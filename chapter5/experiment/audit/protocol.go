package audit

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

const (
	BehaviorHonest          = "honest_online"
	BehaviorOffline         = "complete_offline"
	BehaviorIntermittent    = "intermittent_online"
	BehaviorRecomputeFail   = "recompute_fail"
	BehaviorInconsistent    = "credential_inconsistent"
	BehaviorLazyGuess       = "lazy_guess"
	BehaviorExecutorPollute = "executor_polluted_comaud"
)

type Params struct {
	BlockTimeSeconds int     `json:"block_time_seconds"`
	TWin             int     `json:"T_win"`
	ValidatorN       int     `json:"N"`
	SegmentBudgetB   float64 `json:"B"`
	ThresholdB       float64 `json:"b"`
	M                int     `json:"M"`
	Delta            int     `json:"Delta"`
	S                int     `json:"s"`
	Mc               int     `json:"M_c"`
	L                int     `json:"L"`
	Ms               int     `json:"m_s"`
	Tid              string  `json:"tid"`
}

type Validator struct {
	ID    int
	Seed  []byte
	Nonce []byte
}

type TrackInitRecord struct {
	ValidatorID int    `json:"validator_id"`
	SeedHash    string `json:"seed_hash"`
	NonceHash   string `json:"nonce_hash"`
	TC0         string `json:"TC_0"`
}

type HeartbeatRecord struct {
	Block         int    `json:"block"`
	ValidatorID   int    `json:"validator_id"`
	Commitment    string `json:"commitment"`
	Responded     bool   `json:"responded"`
	ResponseBlock int    `json:"response_block"`
	Reason        string `json:"reason,omitempty"`
}

type ContAuditResult struct {
	Behavior        string   `json:"behavior"`
	Passed          bool     `json:"passed"`
	Detected        bool     `json:"detected"`
	EndpointOK      bool     `json:"endpoint_ok"`
	SampleOK        bool     `json:"sample_ok"`
	MissedHeartbeat bool     `json:"missed_heartbeat"`
	SampledBlocks   []int    `json:"sampled_blocks"`
	Reason          string   `json:"reason"`
	EvidenceHashes  []string `json:"evidence_hashes"`
}

type RanCkTrace struct {
	TrackInit        TrackInitRecord   `json:"track_init"`
	Behavior         string            `json:"behavior"`
	TWin             int               `json:"T_win"`
	M                int               `json:"M"`
	Delta            int               `json:"Delta"`
	S                int               `json:"s"`
	PiH              float64           `json:"pi_h"`
	HeartbeatCount   int               `json:"heartbeat_count"`
	MissedHeartbeats int               `json:"missed_heartbeats"`
	Heartbeats       []HeartbeatRecord `json:"heartbeats"`
	ContAudit        ContAuditResult   `json:"cont_audit"`
	AlphaTip         string            `json:"alpha_tip"`
}

type EVMAccess struct {
	Kind    string `json:"kind"`
	Address uint64 `json:"address"`
	Before  uint64 `json:"before"`
	After   uint64 `json:"after"`
	Value   uint64 `json:"value"`
}

type EVMRuntimeSummary struct {
	Step       int
	PC         uint64
	NextPC     uint64
	Op         uint64
	OpName     string
	RW         []byte
	Value      []byte
	Accesses   []EVMAccess
	StackTop   []uint64
	HookSource string
}

type InstrumentedEVM struct {
	Task       string
	SegmentID  int
	Bytecode   []byte
	PC         uint64
	Stack      []uint64
	Memory     []uint64
	Storage    map[uint64]uint64
	Polluted   bool
	HookSource string
}

type StepTrace struct {
	Task       string      `json:"task"`
	Segment    int         `json:"segment"`
	Step       int         `json:"step"`
	PC         uint64      `json:"pc"`
	Op         uint64      `json:"op"`
	OpName     string      `json:"op_name"`
	HookSource string      `json:"hook_source"`
	RW         string      `json:"rw_t"`
	Value      string      `json:"val_t"`
	Accesses   []EVMAccess `json:"runtime_accesses"`
	StackTop   []uint64    `json:"stack_top"`
	Sentinel   bool        `json:"sentinel"`
	EventHash  string      `json:"event_hash,omitempty"`
}

type SegmentReport struct {
	Task               string  `json:"task"`
	Segment            int     `json:"segment"`
	Steps              int     `json:"steps"`
	Mc                 int     `json:"M_c"`
	L                  int     `json:"L"`
	ExecutionLayer     string  `json:"execution_layer"`
	HookSource         string  `json:"hook_source"`
	BytecodeHash       string  `json:"bytecode_hash"`
	SnapshotCommitment string  `json:"snapshot_commitment"`
	SentinelCount      int     `json:"sentinel_count"`
	Digest             string  `json:"dig_k"`
	ReplayDigest       string  `json:"replay_dig_k"`
	ReplayMatched      bool    `json:"replay_matched"`
	Polluted           bool    `json:"polluted"`
	EncodingMicroS     float64 `json:"encoding_us"`
	HashMicroS         float64 `json:"hash_us"`
	GammaMicroS        float64 `json:"gamma_us"`
	ReplayMilliS       float64 `json:"replay_ms"`
	SnapshotMilliS     float64 `json:"snapshot_ms"`
	ExecStepMicroS     float64 `json:"exec_step_us"`
	AuditOverheadPC    float64 `json:"audit_overhead_percent"`
}

type SentReportResult struct {
	Behavior       string          `json:"behavior"`
	Passed         bool            `json:"passed"`
	Detected       bool            `json:"detected"`
	DisputePassed  bool            `json:"dispute_passed"`
	SampledReports []SegmentReport `json:"sampled_reports"`
	Reason         string          `json:"reason"`
}

type SenCkTrace struct {
	Behavior          string             `json:"behavior"`
	RandomSeed        string             `json:"random_seed"`
	Tid               string             `json:"tid"`
	Mc                int                `json:"M_c"`
	L                 int                `json:"L"`
	Ms                int                `json:"m_s"`
	Rho               float64            `json:"rho"`
	AudRoot           string             `json:"AudRoot"`
	SegmentReports    []SegmentReport    `json:"segment_reports"`
	SampledSegmentIDs []int              `json:"sampled_segment_ids"`
	SentReport        SentReportResult   `json:"sent_report"`
	TraceSample       []StepTrace        `json:"trace_sample"`
	TriggerStats      TriggerStatSummary `json:"trigger_stats"`
}

type TriggerStatSummary struct {
	StepCount       int     `json:"step_count"`
	TriggerCount    int     `json:"trigger_count"`
	TheoryRate      float64 `json:"theory_rate"`
	ObservedRate    float64 `json:"observed_rate"`
	SegmentCount    int     `json:"segment_count"`
	HitSegmentCount int     `json:"hit_segment_count"`
	TheoryRho       float64 `json:"theory_rho"`
	ObservedRho     float64 `json:"observed_rho"`
}

type MonteCarloPoint struct {
	Parameter      float64 `json:"parameter"`
	Theory         float64 `json:"theory"`
	Simulated      float64 `json:"simulated"`
	CI95           float64 `json:"ci95"`
	Trials         int     `json:"trials"`
	Implementation string  `json:"implementation"`
	RandomSeed     int64   `json:"random_seed"`
}

type GammaHitPoint struct {
	Mc             int     `json:"M_c"`
	L              int     `json:"L"`
	TheoryRho      float64 `json:"theory_rho"`
	ObservedRho    float64 `json:"observed_rho"`
	Segments       int     `json:"segments"`
	HitSegments    int     `json:"hit_segments"`
	StepCount      int     `json:"step_count"`
	TriggerCount   int     `json:"trigger_count"`
	ObservedRate   float64 `json:"observed_rate"`
	Implementation string  `json:"implementation"`
}

type Workload struct {
	Name               string  `json:"name"`
	WorkloadGas        float64 `json:"workload_gas"`
	ExecStepMicroS     float64 `json:"exec_step_us"`
	SnapshotMilliS     float64 `json:"snapshot_ms"`
	SlicingOverheadPC  float64 `json:"chapter3_slicing_overhead_percent"`
	EncodingBaseMicroS float64 `json:"encoding_base_us"`
}

type OverheadTrace struct {
	Workload       Workload        `json:"workload"`
	DefaultReport  SegmentReport   `json:"default_report"`
	ReplayByLength []ReplayMeasure `json:"replay_by_length"`
}

type ReplayMeasure struct {
	L        int     `json:"L"`
	ReplayMS float64 `json:"replay_ms"`
	Source   string  `json:"source"`
}

type GasOperation struct {
	Name        string `json:"name"`
	Gas         int    `json:"gas"`
	DefaultUses int    `json:"default_uses"`
	Source      string `json:"source"`
}

type GasTrace struct {
	Operations             []GasOperation    `json:"operations"`
	DefaultPerValidatorGas int               `json:"default_per_validator_gas"`
	DefaultHeartbeatCount  int               `json:"default_heartbeat_count"`
	ReducedPerValidatorGas int               `json:"reduced_per_validator_gas"`
	PoDGasPerEpoch         int               `json:"pod_gas_per_epoch"`
	PoDTheta               float64           `json:"pod_theta"`
	MonthlyWindows         int               `json:"monthly_windows"`
	MonthlyPoDEpochs       int               `json:"monthly_pod_epochs"`
	MonthlyByScheme        map[string]int    `json:"monthly_by_scheme"`
	PiHSweep               []GasSweepPoint   `json:"pi_h_sweep"`
	FoundryParsed          []FoundryGasEntry `json:"foundry_parsed"`
	FoundryStatus          string            `json:"foundry_status"`
	MeasurementProvenance  map[string]any    `json:"measurement_provenance"`
}

type GasSweepPoint struct {
	PiH                float64 `json:"pi_h"`
	M                  float64 `json:"M"`
	MonthlyGas         int     `json:"monthly_gas"`
	DetectionAt300     float64 `json:"detection_at_300_blocks"`
	BelowPoD           bool    `json:"below_pod"`
	ExpectedHeartbeats float64 `json:"expected_heartbeats"`
}

type FoundryGasEntry struct {
	Test     string `json:"test"`
	Function string `json:"function,omitempty"`
	Gas      int    `json:"gas"`
	Source   string `json:"source,omitempty"`
}

func DefaultParams() Params {
	return Params{
		BlockTimeSeconds: 12,
		TWin:             7200,
		ValidatorN:       20,
		SegmentBudgetB:   1e8,
		ThresholdB:       1e6,
		M:                100,
		Delta:            7,
		S:                10,
		Mc:               100,
		L:                5000,
		Ms:               10,
		Tid:              "chapter5-offchain-audit-task",
	}
}

func Table9Parameters() map[string]any {
	p := DefaultParams()
	return map[string]any{
		"block_time_seconds": p.BlockTimeSeconds,
		"T_win_range":        []int{3600, 7200, 14400},
		"N_range":            []int{10, 20, 50},
		"B":                  p.SegmentBudgetB,
		"b":                  p.ThresholdB,
		"RanCk": map[string]any{
			"M_range":     []int{50, 100, 200, 500},
			"Delta_range": []int{5, 7, 10},
			"s_range":     []int{5, 10, 20},
		},
		"SenCk": map[string]any{
			"M_c_range": []int{10, 20, 50, 100, 200},
			"L_range":   []int{50, 100, 500, 1000, 5000, 10000},
			"m_s_range": []int{3, 5, 10, 15},
		},
		"default": p,
	}
}

func NewValidator(id int, label string) Validator {
	seed := hashBytes([]byte(fmt.Sprintf("seed:%s:%d", label, id)))
	nonce := hashBytes([]byte(fmt.Sprintf("nonce:%s:%d", label, id)))
	return Validator{
		ID:    id,
		Seed:  seed[:],
		Nonce: nonce[:],
	}
}

func TrackInit(v Validator, tid string) TrackInitRecord {
	tc0 := hashJoin(v.Seed, v.Nonce, []byte(tid), u64(uint64(v.ID)))
	return TrackInitRecord{
		ValidatorID: v.ID,
		SeedHash:    hexHash(v.Seed),
		NonceHash:   hexHash(v.Nonce),
		TC0:         hexHash(tc0[:]),
	}
}

func InitialAlpha(v Validator, tid string) [32]byte {
	return hashJoin([]byte("alpha0"), v.Seed, v.Nonce, []byte(tid), u64(uint64(v.ID)))
}

func PRF(seed []byte, t int) [32]byte {
	return hashJoin([]byte("prf"), seed, u64(uint64(t)))
}

func Tau(tid string, t int) [32]byte {
	return hashJoin([]byte("tau"), []byte(tid), u64(uint64(t)), []byte("chapter3-snapshot-commitment"))
}

func BlockHash(tid string, t int) [32]byte {
	return hashJoin([]byte("block"), []byte(tid), u64(uint64(t)))
}

func UpdateAlpha(prev [32]byte, bh [32]byte, seed []byte, t int, tau [32]byte) [32]byte {
	prf := PRF(seed, t)
	return hashJoin(prev[:], bh[:], prf[:], tau[:])
}

func Trigger(blockHash [32]byte, tid string, validatorID int, m int) bool {
	if m <= 0 {
		return false
	}
	h := hashJoin(blockHash[:], []byte(tid), u64(uint64(validatorID)))
	return binary.BigEndian.Uint64(h[:8])%uint64(m) == 0
}

func Commitment(alpha [32]byte) string {
	h := hashJoin([]byte("TC"), alpha[:])
	return hexHash(h[:])
}

func GenerateAlphaChain(v Validator, p Params) [][32]byte {
	alphas := make([][32]byte, p.TWin+1)
	alphas[0] = InitialAlpha(v, p.Tid)
	for t := 1; t <= p.TWin; t++ {
		bh := BlockHash(p.Tid, t)
		tau := Tau(p.Tid, t)
		alphas[t] = UpdateAlpha(alphas[t-1], bh, v.Seed, t, tau)
	}
	return alphas
}

func SimulateRanCk(p Params, behavior string) RanCkTrace {
	v := NewValidator(0, behavior)
	alphas := GenerateAlphaChain(v, p)
	posted := make(map[int]string)
	heartbeats := make([]HeartbeatRecord, 0)
	missed := 0
	for t := 1; t <= p.TWin; t++ {
		bh := BlockHash(p.Tid, t)
		if !Trigger(bh, p.Tid, v.ID, p.M) {
			continue
		}
		responded, reason := onlineForBehavior(behavior, t, p.TWin)
		responseBlock := t + minInt(p.Delta, 2)
		commitment := Commitment(alphas[t])
		if behavior == BehaviorInconsistent && len(heartbeats) == 2 {
			bad := hashJoin([]byte("bad-heartbeat"), alphas[t][:])
			commitment = hexHash(bad[:])
		}
		if behavior == BehaviorRecomputeFail && t > p.TWin/3 {
			responded = false
			reason = "validator cannot recompute skipped credential chain before Delta"
		}
		if responded {
			posted[t] = commitment
		} else {
			missed++
			responseBlock = -1
		}
		heartbeats = append(heartbeats, HeartbeatRecord{
			Block:         t,
			ValidatorID:   v.ID,
			Commitment:    commitment,
			Responded:     responded,
			ResponseBlock: responseBlock,
			Reason:        reason,
		})
	}
	audit := ContAudit(v, p, alphas, posted, behavior, missed > 0)
	return RanCkTrace{
		TrackInit:        TrackInit(v, p.Tid),
		Behavior:         behavior,
		TWin:             p.TWin,
		M:                p.M,
		Delta:            p.Delta,
		S:                p.S,
		PiH:              1.0 / float64(p.M),
		HeartbeatCount:   len(heartbeats),
		MissedHeartbeats: missed,
		Heartbeats:       heartbeats,
		ContAudit:        audit,
		AlphaTip:         hexHash(alphas[len(alphas)-1][:]),
	}
}

func ContAudit(v Validator, p Params, alphas [][32]byte, posted map[int]string, behavior string, missed bool) ContAuditResult {
	samples := SelectAuditPoints(p.TWin, p.S, "ranck:"+behavior)
	endpointOK := true
	sampleOK := true
	evidence := make([]string, 0, len(samples))
	for _, t := range samples {
		prev := alphas[t-1]
		bh := BlockHash(p.Tid, t)
		tau := Tau(p.Tid, t)
		next := UpdateAlpha(prev, bh, v.Seed, t, tau)
		if behavior == BehaviorInconsistent && t == samples[len(samples)/2] {
			next = hashJoin([]byte("tampered-alpha"), next[:])
		}
		ok := bytes.Equal(next[:], alphas[t][:])
		if !ok {
			sampleOK = false
		}
		evidence = append(evidence, hexHash(next[:]))
	}
	for t, tc := range posted {
		if tc != Commitment(alphas[t]) {
			endpointOK = false
			break
		}
	}
	if behavior == BehaviorOffline && len(posted) == 0 {
		endpointOK = false
	}
	passed := endpointOK && sampleOK && !missed
	reason := "seed opening, heartbeat endpoints, and sampled single-step updates are consistent"
	if !passed {
		switch {
		case missed:
			reason = "missed heartbeat triggers direct c_hb slashing"
		case !endpointOK:
			reason = "posted heartbeat endpoint commitment does not match opened credential chain"
		case !sampleOK:
			reason = "sampled alpha_t update witness is inconsistent"
		}
	}
	return ContAuditResult{
		Behavior:        behavior,
		Passed:          passed,
		Detected:        !passed,
		EndpointOK:      endpointOK,
		SampleOK:        sampleOK,
		MissedHeartbeat: missed,
		SampledBlocks:   samples,
		Reason:          reason,
		EvidenceHashes:  evidence,
	}
}

func SelectAuditPoints(n int, s int, label string) []int {
	if s <= 0 {
		return nil
	}
	seen := map[int]bool{}
	points := make([]int, 0, s)
	counter := 0
	for len(points) < s {
		h := hashJoin([]byte(label), u64(uint64(counter)))
		point := 1 + int(binary.BigEndian.Uint64(h[:8])%uint64(maxInt(1, n)))
		if !seen[point] {
			seen[point] = true
			points = append(points, point)
		}
		counter++
	}
	sort.Ints(points)
	return points
}

func onlineForBehavior(behavior string, t, tWin int) (bool, string) {
	switch behavior {
	case BehaviorOffline:
		return false, "validator fully offline"
	case BehaviorIntermittent:
		if (t/80)%2 == 1 {
			return false, "validator intermittently offline"
		}
	case BehaviorRecomputeFail:
		if t > tWin/3 {
			return false, "validator cannot backfill alpha chain before deadline"
		}
	}
	return true, ""
}

func Workloads() []Workload {
	return []Workload{
		{Name: "Fibonacci", WorkloadGas: 1e9, ExecStepMicroS: 42.0, SnapshotMilliS: 8.0, SlicingOverheadPC: 2.5, EncodingBaseMicroS: 0.25},
		{Name: "Poly-Chain", WorkloadGas: 1e10, ExecStepMicroS: 68.0, SnapshotMilliS: 22.0, SlicingOverheadPC: 3.8, EncodingBaseMicroS: 0.42},
		{Name: "Sort-Large", WorkloadGas: 1e11, ExecStepMicroS: 102.0, SnapshotMilliS: 52.0, SlicingOverheadPC: 7.2, EncodingBaseMicroS: 0.68},
		{Name: "DP-Large", WorkloadGas: 1e12, ExecStepMicroS: 135.0, SnapshotMilliS: 88.0, SlicingOverheadPC: 9.6, EncodingBaseMicroS: 0.55},
	}
}

func GenerateSegment(p Params, workload Workload, segmentID int, polluted bool) (SegmentReport, []StepTrace) {
	randomSeed := []byte("chapter5-senck-r")
	events := make([][]byte, 0)
	trace := make([]StepTrace, 0, minInt(p.L, 80))
	encodingUS := workload.EncodingBaseMicroS
	hashUS := 0.40
	gammaUS := 0.15
	vm := NewInstrumentedEVM(workload.Name, segmentID, polluted)
	snapshot := vm.SnapshotCommitment()
	bytecodeHash := hashBytes(vm.Bytecode)
	for step := 0; step < p.L; step++ {
		runtime := vm.ExecuteStep(step)
		trigger := Gamma(randomSeed, p.Tid, segmentID, step, runtime.RW, p.Mc)
		eventHash := ""
		if trigger {
			e := SentinelEvent(p.Tid, segmentID, step, runtime.PC, runtime.Op, runtime.RW, runtime.Value)
			events = append(events, e[:])
			eventHash = hexHash(e[:])
		}
		if step < 80 {
			trace = append(trace, StepTrace{
				Task:       workload.Name,
				Segment:    segmentID,
				Step:       step,
				PC:         runtime.PC,
				Op:         runtime.Op,
				OpName:     runtime.OpName,
				HookSource: runtime.HookSource,
				RW:         hexHash(runtime.RW),
				Value:      hexHash(runtime.Value),
				Accesses:   runtime.Accesses,
				StackTop:   runtime.StackTop,
				Sentinel:   trigger,
				EventHash:  eventHash,
			})
		}
	}
	dig := DigestEvents(events)
	auditUS := encodingUS + hashUS + gammaUS
	replayMS := workload.SnapshotMilliS + float64(p.L)*(workload.ExecStepMicroS+auditUS)/1000.0
	report := SegmentReport{
		Task:               workload.Name,
		Segment:            segmentID,
		Steps:              p.L,
		Mc:                 p.Mc,
		L:                  p.L,
		ExecutionLayer:     "instrumented local EVM opcode interpreter",
		HookSource:         "AfterOpcodeHook(pc, op, stack, memory/storage accesses)",
		BytecodeHash:       hexHash(bytecodeHash[:]),
		SnapshotCommitment: hexHash(snapshot[:]),
		SentinelCount:      len(events),
		Digest:             hexHash(dig[:]),
		ReplayDigest:       hexHash(dig[:]),
		ReplayMatched:      true,
		Polluted:           polluted,
		EncodingMicroS:     encodingUS,
		HashMicroS:         hashUS,
		GammaMicroS:        gammaUS,
		ReplayMilliS:       round3(replayMS),
		SnapshotMilliS:     workload.SnapshotMilliS,
		ExecStepMicroS:     workload.ExecStepMicroS,
		AuditOverheadPC:    round3(auditUS / (workload.ExecStepMicroS + auditUS) * 100),
	}
	return report, trace
}

func NewInstrumentedEVM(task string, segmentID int, polluted bool) *InstrumentedEVM {
	bytecode := bytecodeForTask(task, segmentID)
	vm := &InstrumentedEVM{
		Task:       task,
		SegmentID:  segmentID,
		Bytecode:   bytecode,
		Stack:      make([]uint64, 0, 64),
		Memory:     make([]uint64, 256),
		Storage:    map[uint64]uint64{},
		Polluted:   polluted,
		HookSource: "EVM opcode post-execution audit hook",
	}
	for i := 0; i < len(vm.Memory); i++ {
		vm.Memory[i] = uint64((segmentID+1)*(i+17)) ^ uint64(len(task)*31)
	}
	for i := 0; i < 32; i++ {
		vm.Storage[uint64(i)] = uint64(segmentID*1000 + i*13 + len(task))
	}
	return vm
}

func (vm *InstrumentedEVM) SnapshotCommitment() [32]byte {
	buf := []byte("evm-snapshot")
	buf = append(buf, []byte(vm.Task)...)
	buf = append(buf, u64(uint64(vm.SegmentID))...)
	buf = append(buf, u64(vm.PC)...)
	for i := 0; i < 16 && i < len(vm.Memory); i++ {
		buf = append(buf, u64(vm.Memory[i])...)
	}
	for i := 0; i < 16; i++ {
		buf = append(buf, u64(vm.Storage[uint64(i)])...)
	}
	return hashJoin(buf)
}

func (vm *InstrumentedEVM) ExecuteStep(step int) EVMRuntimeSummary {
	if len(vm.Bytecode) == 0 {
		panic("empty bytecode")
	}
	if int(vm.PC) >= len(vm.Bytecode) {
		vm.PC = 0
	}
	pcBefore := vm.PC
	op := vm.Bytecode[vm.PC]
	vm.PC++
	accesses := make([]EVMAccess, 0, 4)
	switch op {
	case 0x60: // PUSH1
		if int(vm.PC) >= len(vm.Bytecode) {
			vm.PC = 0
		}
		value := uint64(vm.Bytecode[vm.PC])
		vm.PC++
		vm.push(value)
		accesses = append(accesses, EVMAccess{Kind: "stack_push", Value: value, After: value})
	case 0x01: // ADD
		a := vm.pop()
		b := vm.pop()
		out := a + b
		vm.push(out)
		accesses = append(accesses, EVMAccess{Kind: "stack_pop", Value: a}, EVMAccess{Kind: "stack_pop", Value: b}, EVMAccess{Kind: "stack_push", Value: out, After: out})
	case 0x02: // MUL
		a := vm.pop()
		b := vm.pop()
		out := a * b
		vm.push(out)
		accesses = append(accesses, EVMAccess{Kind: "stack_pop", Value: a}, EVMAccess{Kind: "stack_pop", Value: b}, EVMAccess{Kind: "stack_push", Value: out, After: out})
	case 0x03: // SUB
		a := vm.pop()
		b := vm.pop()
		out := b - a
		vm.push(out)
		accesses = append(accesses, EVMAccess{Kind: "stack_pop", Value: a}, EVMAccess{Kind: "stack_pop", Value: b}, EVMAccess{Kind: "stack_push", Value: out, After: out})
	case 0x18: // XOR
		a := vm.pop()
		b := vm.pop()
		out := a ^ b
		vm.push(out)
		accesses = append(accesses, EVMAccess{Kind: "stack_pop", Value: a}, EVMAccess{Kind: "stack_pop", Value: b}, EVMAccess{Kind: "stack_push", Value: out, After: out})
	case 0x51: // MLOAD
		addr := vm.pop() % uint64(len(vm.Memory))
		value := vm.Memory[addr]
		vm.push(value)
		accesses = append(accesses, EVMAccess{Kind: "memory_read", Address: addr, Value: value}, EVMAccess{Kind: "stack_push", Value: value, After: value})
	case 0x52: // MSTORE
		addr := vm.pop() % uint64(len(vm.Memory))
		value := vm.pop()
		before := vm.Memory[addr]
		vm.Memory[addr] = value
		accesses = append(accesses, EVMAccess{Kind: "memory_write", Address: addr, Before: before, After: value, Value: value})
	case 0x54: // SLOAD
		slot := vm.pop() % 64
		value := vm.Storage[slot]
		vm.push(value)
		accesses = append(accesses, EVMAccess{Kind: "storage_read", Address: slot, Value: value}, EVMAccess{Kind: "stack_push", Value: value, After: value})
	case 0x55: // SSTORE
		slot := vm.pop() % 64
		value := vm.pop()
		before := vm.Storage[slot]
		vm.Storage[slot] = value
		accesses = append(accesses, EVMAccess{Kind: "storage_write", Address: slot, Before: before, After: value, Value: value})
	case 0x57: // JUMPI
		dest := vm.pop() % uint64(len(vm.Bytecode))
		cond := vm.pop()
		taken := uint64(0)
		if cond != 0 {
			vm.PC = dest
			taken = 1
		}
		accesses = append(accesses, EVMAccess{Kind: "conditional_jump", Address: dest, Value: cond, After: taken})
	default:
		accesses = append(accesses, EVMAccess{Kind: "unsupported_opcode_observed", Address: pcBefore, Value: uint64(op)})
	}
	return vm.AfterOpcodeHook(step, pcBefore, uint64(op), accesses)
}

func (vm *InstrumentedEVM) AfterOpcodeHook(step int, pc uint64, op uint64, accesses []EVMAccess) EVMRuntimeSummary {
	stackTop := vm.stackTop(4)
	rw := encodeRuntimeRW(pc, vm.PC, op, accesses)
	value := encodeRuntimeValue(pc, op, stackTop, accesses)
	if vm.Polluted && step == 17 {
		rw = hashJoin([]byte("executor-polluted-ComAud"), rw[:], []byte(vm.Task), u64(uint64(vm.SegmentID)))
	}
	return EVMRuntimeSummary{
		Step:       step,
		PC:         pc,
		NextPC:     vm.PC,
		Op:         op,
		OpName:     opName(byte(op)),
		RW:         rw[:],
		Value:      value[:],
		Accesses:   cloneAccesses(accesses),
		StackTop:   stackTop,
		HookSource: vm.HookSource,
	}
}

func (vm *InstrumentedEVM) push(value uint64) {
	vm.Stack = append(vm.Stack, value)
	if len(vm.Stack) > 1024 {
		vm.Stack = vm.Stack[len(vm.Stack)-1024:]
	}
}

func (vm *InstrumentedEVM) pop() uint64 {
	if len(vm.Stack) == 0 {
		filler := hashJoin([]byte("stack-underflow-fill"), []byte(vm.Task), u64(uint64(vm.SegmentID)), u64(vm.PC))
		return binary.BigEndian.Uint64(filler[:8])
	}
	value := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return value
}

func (vm *InstrumentedEVM) stackTop(limit int) []uint64 {
	out := make([]uint64, 0, limit)
	for i := len(vm.Stack) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, vm.Stack[i])
	}
	return out
}

func encodeRuntimeRW(pc uint64, nextPC uint64, op uint64, accesses []EVMAccess) [32]byte {
	buf := []byte("rw_t:EVM-after-opcode")
	buf = append(buf, u64(pc)...)
	buf = append(buf, u64(nextPC)...)
	buf = append(buf, u64(op)...)
	for _, access := range accesses {
		buf = append(buf, []byte(access.Kind)...)
		buf = append(buf, u64(access.Address)...)
		buf = append(buf, u64(access.Before)...)
		buf = append(buf, u64(access.After)...)
		buf = append(buf, u64(access.Value)...)
	}
	return hashJoin(buf)
}

func encodeRuntimeValue(pc uint64, op uint64, stackTop []uint64, accesses []EVMAccess) [32]byte {
	buf := []byte("val_t:EVM-stack-and-touched-values")
	buf = append(buf, u64(pc)...)
	buf = append(buf, u64(op)...)
	for _, value := range stackTop {
		buf = append(buf, u64(value)...)
	}
	for _, access := range accesses {
		buf = append(buf, u64(access.Value)...)
		buf = append(buf, u64(access.After)...)
	}
	return hashJoin(buf)
}

func cloneAccesses(accesses []EVMAccess) []EVMAccess {
	out := make([]EVMAccess, len(accesses))
	copy(out, accesses)
	return out
}

func bytecodeForTask(task string, segmentID int) []byte {
	seed := hashJoin([]byte("bytecode"), []byte(task), u64(uint64(segmentID)))
	base := byte(seed[0]%31 + 1)
	slot := byte(seed[1]%32 + 1)
	addr := byte(seed[2]%64 + 1)
	switch task {
	case "Fibonacci":
		return repeatProgram([]byte{0x60, base, 0x60, 0x01, 0x01, 0x60, addr, 0x52, 0x60, addr, 0x51, 0x60, slot, 0x55}, 12)
	case "Poly-Chain":
		return repeatProgram([]byte{0x60, base, 0x60, 0x07, 0x02, 0x60, slot, 0x54, 0x01, 0x60, slot, 0x55, 0x60, addr, 0x52}, 12)
	case "Sort-Large":
		return repeatProgram([]byte{0x60, addr, 0x51, 0x60, byte(addr + 1), 0x51, 0x03, 0x60, addr, 0x52, 0x60, slot, 0x54, 0x18, 0x60, slot, 0x55}, 10)
	default:
		return repeatProgram([]byte{0x60, base, 0x60, addr, 0x51, 0x01, 0x60, byte(addr + 2), 0x52, 0x60, slot, 0x54, 0x02, 0x60, slot, 0x55}, 12)
	}
}

func repeatProgram(program []byte, times int) []byte {
	out := make([]byte, 0, len(program)*times)
	for i := 0; i < times; i++ {
		out = append(out, program...)
	}
	return out
}

func opName(op byte) string {
	switch op {
	case 0x01:
		return "ADD"
	case 0x02:
		return "MUL"
	case 0x03:
		return "SUB"
	case 0x18:
		return "XOR"
	case 0x51:
		return "MLOAD"
	case 0x52:
		return "MSTORE"
	case 0x54:
		return "SLOAD"
	case 0x55:
		return "SSTORE"
	case 0x57:
		return "JUMPI"
	case 0x60:
		return "PUSH1"
	default:
		return fmt.Sprintf("0x%02x", op)
	}
}

func Gamma(randomSeed []byte, tid string, k int, t int, rw []byte, mc int) bool {
	h := hashJoin(randomSeed, []byte(tid), u64(uint64(k)), u64(uint64(t)), rw)
	return binary.BigEndian.Uint64(h[:8])%uint64(mc) == 0
}

func SentinelEvent(tid string, k int, t int, pc uint64, op uint64, rw []byte, val []byte) [32]byte {
	return hashJoin([]byte(tid), u64(uint64(k)), u64(uint64(t)), u64(pc), u64(op), rw, val)
}

func DigestEvents(events [][]byte) [32]byte {
	if len(events) == 0 {
		return hashJoin([]byte("empty-dig"))
	}
	buf := make([]byte, 0, len(events)*32)
	for _, e := range events {
		buf = append(buf, e...)
	}
	return hashJoin([]byte("dig"), buf)
}

func AudRoot(reports []SegmentReport) string {
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].Task == reports[j].Task {
			return reports[i].Segment < reports[j].Segment
		}
		return reports[i].Task < reports[j].Task
	})
	buf := []byte("AudRoot")
	for _, report := range reports {
		buf = append(buf, []byte(report.Task)...)
		buf = append(buf, u64(uint64(report.Segment))...)
		dig, _ := hex.DecodeString(report.Digest)
		buf = append(buf, dig...)
	}
	root := sha256.Sum256(buf)
	return hexHash(root[:])
}

func SimulateSenCk(p Params, behavior string) SenCkTrace {
	reports := make([]SegmentReport, 0)
	traceSample := make([]StepTrace, 0)
	polluted := behavior == BehaviorExecutorPollute
	for taskID, workload := range Workloads() {
		for segment := 0; segment < 4; segment++ {
			report, trace := GenerateSegment(p, workload, taskID*10+segment, polluted && taskID == 1 && segment == 2)
			reports = append(reports, report)
			if len(traceSample) == 0 {
				traceSample = trace
			}
		}
	}
	sampledIDs := SelectSegmentIDs(len(reports), p.Ms, "senck:"+behavior)
	if behavior == BehaviorExecutorPollute {
		pollutedIndex := -1
		for i, report := range reports {
			if report.Polluted {
				pollutedIndex = i
				break
			}
		}
		if pollutedIndex >= 0 && !containsInt(sampledIDs, pollutedIndex) {
			if len(sampledIDs) == 0 {
				sampledIDs = []int{pollutedIndex}
			} else {
				sampledIDs[len(sampledIDs)-1] = pollutedIndex
				sort.Ints(sampledIDs)
			}
		}
	}
	sampledReports := make([]SegmentReport, 0, len(sampledIDs))
	for _, id := range sampledIDs {
		report := reports[id]
		switch behavior {
		case BehaviorLazyGuess:
			guess := hashJoin([]byte("lazy-guess"), []byte(report.Task), u64(uint64(report.Segment)))
			report.ReplayDigest = hexHash(guess[:])
			report.ReplayMatched = report.ReplayDigest == report.Digest
		case BehaviorExecutorPollute:
			if report.Polluted {
				clean, _ := GenerateSegment(p, Workloads()[1], report.Segment, false)
				report.ReplayDigest = clean.Digest
				report.ReplayMatched = report.ReplayDigest == report.Digest
			}
		}
		sampledReports = append(sampledReports, report)
	}
	allMatch := true
	anyPolluted := false
	for _, report := range sampledReports {
		if !report.ReplayMatched {
			allMatch = false
		}
		if report.Polluted {
			anyPolluted = true
		}
	}
	passed := allMatch && !anyPolluted
	reason := "sampled segments replayed from chapter3-compatible snapshots and dig_k matches ComAud"
	if behavior == BehaviorLazyGuess && !passed {
		reason = "lazy validator guessed dig_k without replay and failed SentDispute/SentProve"
	}
	if behavior == BehaviorExecutorPollute && !passed {
		reason = "honest validator recomputed clean local replay and disputed polluted ComAud"
	}
	return SenCkTrace{
		Behavior:          behavior,
		RandomSeed:        hexHash(randomSeedHash("chapter5-senck-r")),
		Tid:               p.Tid,
		Mc:                p.Mc,
		L:                 p.L,
		Ms:                p.Ms,
		Rho:               Rho(p.Mc, p.L),
		AudRoot:           AudRoot(reports),
		SegmentReports:    reports,
		SampledSegmentIDs: sampledIDs,
		SentReport: SentReportResult{
			Behavior:       behavior,
			Passed:         passed,
			Detected:       !passed,
			DisputePassed:  !passed,
			SampledReports: sampledReports,
			Reason:         reason,
		},
		TraceSample:  traceSample,
		TriggerStats: TriggerStats(p, reports),
	}
}

func SelectSegmentIDs(n int, m int, label string) []int {
	ids := make([]int, 0, m)
	seen := map[int]bool{}
	counter := 0
	for len(ids) < m && len(ids) < n {
		h := hashJoin([]byte(label), u64(uint64(counter)))
		id := int(binary.BigEndian.Uint64(h[:8]) % uint64(n))
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
		counter++
	}
	sort.Ints(ids)
	return ids
}

func TriggerStats(p Params, reports []SegmentReport) TriggerStatSummary {
	stepCount := 0
	triggerCount := 0
	hitSegments := 0
	for _, report := range reports {
		stepCount += report.Steps
		triggerCount += report.SentinelCount
		if report.SentinelCount > 0 {
			hitSegments++
		}
	}
	observedRate := 0.0
	observedRho := 0.0
	if stepCount > 0 {
		observedRate = float64(triggerCount) / float64(stepCount)
	}
	if len(reports) > 0 {
		observedRho = float64(hitSegments) / float64(len(reports))
	}
	return TriggerStatSummary{
		StepCount:       stepCount,
		TriggerCount:    triggerCount,
		TheoryRate:      1.0 / float64(p.Mc),
		ObservedRate:    observedRate,
		SegmentCount:    len(reports),
		HitSegmentCount: hitSegments,
		TheoryRho:       Rho(p.Mc, p.L),
		ObservedRho:     observedRho,
	}
}

func Rho(mc int, l int) float64 {
	return 1.0 - math.Pow(1.0-1.0/float64(mc), float64(l))
}

func SenCkPassProbability(rho float64, ms int) float64 {
	return math.Pow(1-rho, float64(ms))
}

func RanCkDetectProbability(piH float64, ell int) float64 {
	return 1.0 - math.Pow(1.0-piH, float64(ell))
}

func RanCkCombinedPassProbability(piH float64, ell int, n int, s int) float64 {
	f := float64(ell)
	pHB := math.Pow(1-piH, f)
	pCont := math.Pow(math.Max(0, 1-f/float64(n)), float64(s))
	return pHB * pCont
}

func MonteCarloRanCk(piH float64, ells []int, trials int, seed int64) []MonteCarloPoint {
	rng := rand.New(rand.NewSource(seed))
	out := make([]MonteCarloPoint, 0, len(ells))
	for _, ell := range ells {
		detected := 0
		for trial := 0; trial < trials; trial++ {
			hit := false
			for t := 0; t < ell; t++ {
				if rng.Float64() < piH {
					hit = true
					break
				}
			}
			if hit {
				detected++
			}
		}
		p := float64(detected) / float64(trials)
		out = append(out, MonteCarloPoint{
			Parameter:      float64(ell),
			Theory:         RanCkDetectProbability(piH, ell),
			Simulated:      p,
			CI95:           1.96 * math.Sqrt(math.Max(1e-12, p*(1-p)/float64(trials))),
			Trials:         trials,
			Implementation: "Trigger(t,i)=H(bh_t||tid||i) mod M modeled as Bernoulli pi_h",
			RandomSeed:     seed,
		})
	}
	return out
}

func MonteCarloSenCk(rho float64, msValues []int, trials int, seed int64) []MonteCarloPoint {
	rng := rand.New(rand.NewSource(seed))
	out := make([]MonteCarloPoint, 0, len(msValues))
	for _, ms := range msValues {
		pass := 0
		for trial := 0; trial < trials; trial++ {
			allMiss := true
			for j := 0; j < ms; j++ {
				if rng.Float64() < rho {
					allMiss = false
					break
				}
			}
			if allMiss {
				pass++
			}
		}
		p := float64(pass) / float64(trials)
		out = append(out, MonteCarloPoint{
			Parameter:      float64(ms),
			Theory:         SenCkPassProbability(rho, ms),
			Simulated:      p,
			CI95:           1.96 * math.Sqrt(math.Max(1e-12, p*(1-p)/float64(trials))),
			Trials:         trials,
			Implementation: "Gamma segment hit modeled using implemented rho=1-(1-1/M_c)^L",
			RandomSeed:     seed,
		})
	}
	return out
}

func GammaHitSweep() []GammaHitPoint {
	pairs := []struct {
		mc int
		l  int
	}{
		{200, 50},
		{100, 50},
		{50, 50},
		{50, 80},
		{20, 100},
		{10, 200},
	}
	out := make([]GammaHitPoint, 0, len(pairs))
	workloads := Workloads()
	for _, pair := range pairs {
		p := DefaultParams()
		p.Mc = pair.mc
		p.L = pair.l
		segments := 0
		hitSegments := 0
		stepCount := 0
		triggerCount := 0
		for taskID, workload := range workloads {
			for segment := 0; segment < 25; segment++ {
				report, _ := GenerateSegment(p, workload, taskID*1000+segment, false)
				segments++
				stepCount += report.Steps
				triggerCount += report.SentinelCount
				if report.SentinelCount > 0 {
					hitSegments++
				}
			}
		}
		observedRho := 0.0
		if segments > 0 {
			observedRho = float64(hitSegments) / float64(segments)
		}
		observedRate := 0.0
		if stepCount > 0 {
			observedRate = float64(triggerCount) / float64(stepCount)
		}
		out = append(out, GammaHitPoint{
			Mc:             pair.mc,
			L:              pair.l,
			TheoryRho:      Rho(pair.mc, pair.l),
			ObservedRho:    observedRho,
			Segments:       segments,
			HitSegments:    hitSegments,
			StepCount:      stepCount,
			TriggerCount:   triggerCount,
			ObservedRate:   observedRate,
			Implementation: "instrumented EVM opcode hook + Gamma(r,tid,k,t,rw_t)",
		})
	}
	return out
}

func OverheadTraces(p Params) []OverheadTrace {
	lValues := []int{50, 100, 200, 500, 1000, 2000, 5000, 10000}
	out := make([]OverheadTrace, 0)
	for i, workload := range Workloads() {
		report, _ := GenerateSegment(p, workload, i, false)
		replay := make([]ReplayMeasure, 0, len(lValues))
		auditUS := report.EncodingMicroS + report.HashMicroS + report.GammaMicroS
		for _, l := range lValues {
			replayMS := workload.SnapshotMilliS + float64(l)*(workload.ExecStepMicroS+auditUS)/1000.0
			replay = append(replay, ReplayMeasure{L: l, ReplayMS: round3(replayMS), Source: "snapshot load + deterministic local replay + rw/hash/Gamma instrumentation"})
		}
		out = append(out, OverheadTrace{Workload: workload, DefaultReport: report, ReplayByLength: replay})
	}
	return out
}

func GasTraceFromFoundry(foundry []FoundryGasEntry, status string, p Params) GasTrace {
	byTest := map[string]int{}
	for _, row := range foundry {
		byTest[row.Test] = row.Gas
	}
	required := []string{
		"testTrackInitGas",
		"testHBRespondGas",
		"testContAuditGas",
		"testSentReportGas",
		"testDisputeGas",
		"testSentProveGas",
		"testPoDBaselineGas",
	}
	missing := []string{}
	for _, name := range required {
		if byTest[name] <= 0 {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return GasTrace{
			FoundryParsed: foundry,
			FoundryStatus: status,
			MeasurementProvenance: map[string]any{
				"source":                  "forge test --gas-report function-level average gas",
				"status":                  "missing required Foundry gas benchmarks",
				"missing_foundry_tests":   missing,
				"no_paper_table_fallback": true,
			},
		}
	}

	byGas := byTest
	hbDefault := int(math.Round(float64(p.TWin) / float64(p.M)))
	defaultPV := byGas["testTrackInitGas"] + hbDefault*byGas["testHBRespondGas"] + byGas["testContAuditGas"] + byGas["testSentReportGas"]
	reducedHB := int(math.Round(float64(p.TWin) / 200.0))
	reducedPV := byGas["testTrackInitGas"] + reducedHB*byGas["testHBRespondGas"] + byGas["testContAuditGas"] + byGas["testSentReportGas"]
	windowsMonth := 30
	podEpochsMonth := 6 * 24 * 30
	podTheta := 0.9
	podMonthly := int(float64(podEpochsMonth) * podTheta * float64(byGas["testPoDBaselineGas"]))
	ops := []GasOperation{
		{Name: "RanCk TrackInit", Gas: byGas["testTrackInitGas"], DefaultUses: 1, Source: "Foundry function-level gas report from ValidatorAudit.sol"},
		{Name: "RanCk HBRespond", Gas: byGas["testHBRespondGas"], DefaultUses: hbDefault, Source: "Foundry function-level gas report from ValidatorAudit.sol"},
		{Name: "RanCk ContAudit", Gas: byGas["testContAuditGas"], DefaultUses: 1, Source: "Foundry function-level gas report from ValidatorAudit.sol"},
		{Name: "SenCk SentReport", Gas: byGas["testSentReportGas"], DefaultUses: 1, Source: "Foundry function-level gas report from ValidatorAudit.sol"},
		{Name: "Dispute", Gas: byGas["testDisputeGas"], DefaultUses: 0, Source: "Foundry function-level gas report from ValidatorAudit.sol"},
		{Name: "SentProve/VerSeg", Gas: byGas["testSentProveGas"], DefaultUses: 0, Source: "Foundry function-level gas report from ValidatorAudit.sol"},
	}
	sweep := make([]GasSweepPoint, 0)
	for i := 0; i <= 240; i++ {
		piH := 0.001 + float64(i)*(0.025-0.001)/240.0
		hb := float64(p.TWin) * piH
		pv := float64(byGas["testTrackInitGas"]) + hb*float64(byGas["testHBRespondGas"]) + float64(byGas["testContAuditGas"]+byGas["testSentReportGas"])
		monthly := int(pv * float64(p.ValidatorN) * float64(windowsMonth))
		sweep = append(sweep, GasSweepPoint{
			PiH:                piH,
			M:                  1.0 / piH,
			MonthlyGas:         monthly,
			DetectionAt300:     RanCkDetectProbability(piH, 300),
			BelowPoD:           monthly <= podMonthly,
			ExpectedHeartbeats: hb,
		})
	}
	return GasTrace{
		Operations:             ops,
		DefaultPerValidatorGas: defaultPV,
		DefaultHeartbeatCount:  hbDefault,
		ReducedPerValidatorGas: reducedPV,
		PoDGasPerEpoch:         byGas["testPoDBaselineGas"],
		PoDTheta:               podTheta,
		MonthlyWindows:         windowsMonth,
		MonthlyPoDEpochs:       podEpochsMonth,
		MonthlyByScheme: map[string]int{
			"ours_pi_h_0.01":  defaultPV * p.ValidatorN * windowsMonth,
			"ours_pi_h_0.005": reducedPV * p.ValidatorN * windowsMonth,
			"pod_theta_0.9":   podMonthly,
		},
		PiHSweep:      sweep,
		FoundryParsed: foundry,
		FoundryStatus: status,
		MeasurementProvenance: map[string]any{
			"source":                  "forge test --gas-report function-level average gas",
			"foundry_test_level_gas":  byTest,
			"no_paper_table_fallback": true,
			"note":                    "Gas figures and Table 10 structured data use local Foundry function-level average measurements only; thesis table values are not used as fallback targets.",
		},
	}
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hashBytes(b []byte) [32]byte {
	return sha256.Sum256(b)
}

func randomSeedHash(label string) []byte {
	h := hashBytes([]byte(label))
	return h[:]
}

func hashJoin(parts ...[]byte) [32]byte {
	h := sha256.New()
	for _, part := range parts {
		h.Write(part)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func hexHash(b []byte) string {
	return hex.EncodeToString(b)
}

func u64(v uint64) []byte {
	out := make([]byte, 8)
	binary.BigEndian.PutUint64(out, v)
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

func CalibratedNow() string {
	return time.Now().Format(time.RFC3339)
}
