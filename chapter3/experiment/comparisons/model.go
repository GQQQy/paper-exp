package comparisons

import (
	"fmt"
	"math"
	"sort"
)

type PathSpec struct {
	Function string         `json:"function"`
	Test     string         `json:"test"`
	Args     map[string]int `json:"args"`
}

type ReferenceSource struct {
	Project  string `json:"project"`
	URL      string `json:"url"`
	Evidence string `json:"evidence"`
	UsedFor  string `json:"used_for"`
}

type ProtocolSpec struct {
	Scheme            string            `json:"scheme"`
	Mechanism         string            `json:"mechanism"`
	Family            string            `json:"family"`
	OnchainRounds     int               `json:"onchain_rounds"`
	Optimistic        PathSpec          `json:"optimistic_benchmark"`
	Dispute           PathSpec          `json:"dispute_benchmark"`
	DisputeFlow       []string          `json:"dispute_flow"`
	Timeline          TimelineModel     `json:"timeline_model"`
	Concurrent        bool              `json:"concurrent"`
	ModelScope        string            `json:"model_scope"`
	SourceKind        string            `json:"source_kind"`
	ReferenceSources  []ReferenceSource `json:"reference_sources"`
	MeasurementMethod string            `json:"measurement_method"`
}

type TimelineModel struct {
	Strategy              string `json:"strategy"`
	InitialOffchainRounds int    `json:"initial_offchain_rounds,omitempty"`
	RoundsPerSlot         int    `json:"rounds_per_slot,omitempty"`
	Levels                int    `json:"levels,omitempty"`
	LevelBoundarySlots    int    `json:"level_boundary_slots,omitempty"`
	ConcurrentSlots       int    `json:"concurrent_slots,omitempty"`
}

type PathResult struct {
	Function string         `json:"function"`
	Test     string         `json:"test"`
	Args     map[string]int `json:"args"`
	Gas      uint64         `json:"gas"`
	GasK     float64        `json:"gas_k"`
}

type ProtocolResult struct {
	Scheme            string             `json:"scheme"`
	Mechanism         string             `json:"mechanism"`
	Family            string             `json:"family"`
	OnchainRounds     int                `json:"onchain_rounds"`
	OptimisticGasK    float64            `json:"optimistic_gas_k"`
	DisputeGasK       float64            `json:"dispute_gas_k"`
	Optimistic        PathResult         `json:"optimistic_benchmark"`
	Dispute           PathResult         `json:"dispute_benchmark"`
	DisputeFlow       []string           `json:"dispute_flow"`
	TimelineSlots     int                `json:"timeline_slots"`
	Timeline          TimelineDerivation `json:"timeline_derivation"`
	Concurrent        bool               `json:"concurrent"`
	ModelScope        string             `json:"model_scope"`
	SourceKind        string             `json:"source_kind"`
	ReferenceSources  []ReferenceSource  `json:"reference_sources"`
	MeasurementMethod string             `json:"measurement_method"`
	Reproduce         string             `json:"reproduce"`
	Notes             string             `json:"notes"`
	ValidationClaim   string             `json:"validation_claim"`
}

type TimelineParams struct {
	TExec float64 `json:"t_exec"`
	TSlot float64 `json:"t_slot"`
	Gamma float64 `json:"gamma"`
	TSeg  float64 `json:"t_seg"`
	Eta   float64 `json:"eta"`
}

type TimelineResult struct {
	Name         string             `json:"name"`
	DisputeSlots int                `json:"dispute_slots"`
	TotalTime    float64            `json:"total_time"`
	Formula      string             `json:"formula"`
	Derivation   TimelineDerivation `json:"timeline_derivation"`
}

type TimelineEvent struct {
	Slot              int            `json:"slot"`
	Phase             string         `json:"phase"`
	LogicalRoundStart int            `json:"logical_round_start,omitempty"`
	LogicalRoundEnd   int            `json:"logical_round_end,omitempty"`
	Detail            string         `json:"detail"`
	StateBefore       map[string]int `json:"state_before,omitempty"`
	StateAfter        map[string]int `json:"state_after,omitempty"`
	Transition        string         `json:"transition,omitempty"`
}

type TimelineSimulationInput struct {
	TotalGas              float64 `json:"total_gas"`
	LogicalRounds         int     `json:"logical_rounds"`
	Strategy              string  `json:"strategy"`
	InitialOffchainRounds int     `json:"initial_offchain_rounds,omitempty"`
	RoundsPerSlot         int     `json:"rounds_per_slot,omitempty"`
	Levels                int     `json:"levels,omitempty"`
	LevelBoundarySlots    int     `json:"level_boundary_slots,omitempty"`
	ConcurrentSlots       int     `json:"concurrent_slots,omitempty"`
}

type TimelineDerivation struct {
	Strategy       string                  `json:"strategy"`
	TotalGas       float64                 `json:"total_gas"`
	LogicalRounds  int                     `json:"logical_rounds"`
	Slots          int                     `json:"slots"`
	Events         []TimelineEvent         `json:"events"`
	Formula        string                  `json:"slot_formula"`
	SimulationKind string                  `json:"simulation_kind"`
	Inputs         TimelineSimulationInput `json:"simulation_inputs"`
	Resolved       bool                    `json:"resolved"`
}

type TimelineExperimentReport struct {
	Source    string                       `json:"source"`
	Reproduce string                       `json:"reproduce"`
	TotalGas  float64                      `json:"total_gas"`
	Params    TimelineParams               `json:"timeline_params"`
	Protocols []TimelineExperimentProtocol `json:"protocols"`
}

type TimelineExperimentProtocol struct {
	Scheme             string             `json:"scheme"`
	Mechanism          string             `json:"mechanism"`
	Family             string             `json:"family"`
	DisputeBenchmark   PathSpec           `json:"dispute_benchmark"`
	TimelineModel      TimelineModel      `json:"timeline_model"`
	TimelineDerivation TimelineDerivation `json:"timeline_derivation"`
}

func Specs(totalGas float64) []ProtocolSpec {
	rounds := int(math.Ceil(math.Log2(totalGas)))
	foundryMethod := "src/DisputeProtocolBenchmarks.sol implements a local same-metric protocol path; forge test --gas-report measures the EVM average gas for each named path"
	return []ProtocolSpec{
		{
			Scheme:        "Arbitrum\nClassic",
			Mechanism:     "1v1 bisection",
			Family:        "post-execution interactive fraud proof",
			OnchainRounds: rounds,
			Optimistic:    PathSpec{Function: "arbitrumOptimisticPath", Test: "testArbitrumOptimisticPathGas", Args: map[string]int{"rounds": 660}},
			Dispute:       PathSpec{Function: "arbitrumClassicPath", Test: "testArbitrumClassicDisputePathGas", Args: map[string]int{"rounds": 8001}},
			DisputeFlow: []string{
				"submit assertion state root",
				"open one-vs-one challenge",
				"bisect execution interval until one instruction remains",
				"verify AVM-style one-step proof and record adjudication",
			},
			Timeline:   TimelineModel{Strategy: "one_slot_per_remaining_bisection_round", InitialOffchainRounds: 7},
			ModelScope: fmt.Sprintf("O(log N) bisection path for a %.0e-gas task, with %d logical dispute rounds; benchmark rounds are calibrated to the Solidity submit/challenge/localize/adjudicate path", totalGas, rounds),
			SourceKind: "solidity_protocol_benchmark",
			ReferenceSources: []ReferenceSource{
				{
					Project:  "OffchainLabs/arbitrum-classic",
					URL:      "https://github.com/OffchainLabs/arbitrum-classic",
					Evidence: "Arbitrum Classic repository used to define the assertion/challenge/bisection dispute shape",
					UsedFor:  "reference boundary for the local assertion, challenge, bisection, and one-step-proof benchmark stages",
				},
			},
			MeasurementMethod: foundryMethod,
		},
		{
			Scheme:        "TrueBit",
			Mechanism:     "solver-verifier bisection",
			Family:        "post-execution solver-verifier verification game",
			OnchainRounds: rounds,
			Optimistic:    PathSpec{Function: "truebitOptimisticPath", Test: "testTrueBitOptimisticPathGas", Args: map[string]int{"rounds": 742}},
			Dispute:       PathSpec{Function: "truebitPath", Test: "testTrueBitDisputePathGas", Args: map[string]int{"rounds": 4992}},
			DisputeFlow: []string{
				"solver posts task commitment",
				"verifier opens challenge",
				"solver and verifier alternate bisection commitments",
				"single-step judge resolves the isolated instruction",
			},
			Timeline:   TimelineModel{Strategy: "solver_verifier_remaining_bisection_rounds", InitialOffchainRounds: 4},
			ModelScope: fmt.Sprintf("solver/verifier bisection path for a %.0e-gas task, including alternating verifier-response branches and final judge", totalGas),
			SourceKind: "solidity_protocol_benchmark",
			ReferenceSources: []ReferenceSource{
				{
					Project:  "TrueBitFoundation/truebit-eth",
					URL:      "https://github.com/TrueBitFoundation/truebit-eth",
					Evidence: "TrueBit repository used to define the solver/verifier verification-game shape",
					UsedFor:  "reference boundary for the local solver commitment, verifier challenge, bisection, and final judge benchmark stages",
				},
				{
					Project:  "Truebit documentation",
					URL:      "https://docs.truebit.io/v1docs",
					Evidence: "TrueBit documentation used to confirm task and role workflow terminology",
					UsedFor:  "reference boundary for the local TrueBit task/verification workflow model",
				},
			},
			MeasurementMethod: foundryMethod,
		},
		{
			Scheme:        "Cartesi\nDave",
			Mechanism:     "tournament + bisection",
			Family:        "post-execution tournament dispute",
			OnchainRounds: rounds,
			Optimistic:    PathSpec{Function: "cartesiOptimisticPath", Test: "testCartesiOptimisticPathGas", Args: map[string]int{"rounds": 818}},
			Dispute:       PathSpec{Function: "cartesiDavePath", Test: "testCartesiDaveDisputePathGas", Args: map[string]int{"rounds": 4927, "tournament_slots": 633}},
			DisputeFlow: []string{
				"set up Dave tournament bracket",
				"select surviving claim/counterclaim pair",
				"run machine-state bisection",
				"submit one-step machine proof to the referee",
			},
			Timeline:   TimelineModel{Strategy: "tournament_then_remaining_bisection_rounds", InitialOffchainRounds: 2},
			ModelScope: "Dave-style tournament setup plus claim/counterclaim bisection path, measured by the Solidity benchmark",
			SourceKind: "solidity_protocol_benchmark",
			ReferenceSources: []ReferenceSource{
				{
					Project:  "cartesi/dave",
					URL:      "https://github.com/cartesi/dave",
					Evidence: "Dave repository used to define the tournament plus bisection dispute shape",
					UsedFor:  "reference boundary for the local tournament setup, claim/counterclaim bisection, and referee benchmark stages",
				},
			},
			MeasurementMethod: foundryMethod,
		},
		{
			Scheme:        "Arbitrum\nBoLD",
			Mechanism:     "three-level bisection",
			Family:        "bounded-liquidity multi-level dispute",
			OnchainRounds: rounds,
			Optimistic:    PathSpec{Function: "boldOptimisticPath", Test: "testBoLDOptimisticPathGas", Args: map[string]int{"rounds": 722}},
			Dispute:       PathSpec{Function: "boldPath", Test: "testBoLDDisputePathGas", Args: map[string]int{"rounds": 1896, "levels": 3}},
			DisputeFlow: []string{
				"submit top-level assertion tree",
				"open parallel challenge edges",
				"narrow disputes across three BoLD levels",
				"confirm the lowest-level edge with one-step adjudication",
			},
			Timeline:   TimelineModel{Strategy: "bounded_liquidity_parallel_levels", RoundsPerSlot: 2, Levels: 3, LevelBoundarySlots: 6},
			ModelScope: "BoLD-style three-level dispute path, measured by the Solidity benchmark",
			SourceKind: "solidity_protocol_benchmark",
			ReferenceSources: []ReferenceSource{
				{
					Project:  "OffchainLabs/bold",
					URL:      "https://github.com/OffchainLabs/bold",
					Evidence: "BoLD repository used to define the bounded-liquidity multi-level dispute shape",
					UsedFor:  "reference boundary for the local parallel edge, multi-level narrowing, and lowest-level confirmation benchmark stages",
				},
			},
			MeasurementMethod: foundryMethod,
		},
		{
			Scheme:        "CleVer\n(ours)",
			Mechanism:     "two-layer slicing + VerSeg",
			Family:        "concurrent sliced verification",
			OnchainRounds: 3,
			Optimistic:    PathSpec{Function: "cleverOptimisticPath", Test: "testCleVerOptimisticPathGas", Args: map[string]int{"rounds": 1006}},
			Dispute:       PathSpec{Function: "cleverPath", Test: "testCleVerDisputePathGas", Args: map[string]int{"subsegments": 1135, "replay_steps": 197}},
			DisputeFlow: []string{
				"submit two-layer slice root",
				"challenge outer and inner slice commitments",
				"localize a bounded subsegment concurrently",
				"run VerSeg replay and record adjudication",
			},
			Timeline:   TimelineModel{Strategy: "concurrent_two_layer_slice_then_verseg", ConcurrentSlots: 3},
			Concurrent: true,
			ModelScope: "CleVer two-layer slicing and bounded VerSeg path, measured by the verifier benchmark",
			SourceKind: "project_solidity_benchmark",
			ReferenceSources: []ReferenceSource{
				{
					Project:  "paper-exp chapter3",
					URL:      "chapter3/experiment/src/CleVerVerifier.sol",
					Evidence: "local CleVer verifier benchmark used by the thesis implementation",
					UsedFor:  "two-layer localization and bounded VerSeg replay benchmark stages",
				},
			},
			MeasurementMethod: "local CleVer verifier path is measured by forge test --gas-report and parsed into the raw log",
		},
	}
}

func RunTimelineExperiment(totalGas float64, params TimelineParams) TimelineExperimentReport {
	specs := Specs(totalGas)
	protocols := make([]TimelineExperimentProtocol, 0, len(specs))
	for _, spec := range specs {
		protocols = append(protocols, TimelineExperimentProtocol{
			Scheme:             spec.Scheme,
			Mechanism:          spec.Mechanism,
			Family:             spec.Family,
			DisputeBenchmark:   spec.Dispute,
			TimelineModel:      spec.Timeline,
			TimelineDerivation: RunTimelineSimulation(spec, totalGas),
		})
	}
	return TimelineExperimentReport{
		Source:    "local dispute timeline state-machine simulation; slots are counted from emitted state-transition events",
		Reproduce: "go run ./cmd/timeline-exp --total-gas 1e11 --out logs/timeline_simulation.json",
		TotalGas:  totalGas,
		Params:    params,
		Protocols: protocols,
	}
}

func TimelineDerivationsByScheme(report TimelineExperimentReport) map[string]TimelineDerivation {
	out := make(map[string]TimelineDerivation, len(report.Protocols))
	for _, protocol := range report.Protocols {
		out[protocol.Scheme] = protocol.TimelineDerivation
	}
	return out
}

func BuildResults(gasByFunction map[string]uint64, totalGas float64) ([]ProtocolResult, error) {
	return BuildResultsWithTimelines(gasByFunction, totalGas, nil)
}

func BuildResultsWithTimelines(gasByFunction map[string]uint64, totalGas float64, timelines map[string]TimelineDerivation) ([]ProtocolResult, error) {
	specs := Specs(totalGas)
	missing := missingFunctions(specs, gasByFunction)
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing Foundry gas measurements for comparison benchmarks: %v", missing)
	}
	out := make([]ProtocolResult, 0, len(specs))
	for _, spec := range specs {
		optimistic := pathResult(spec.Optimistic, gasByFunction[spec.Optimistic.Function])
		dispute := pathResult(spec.Dispute, gasByFunction[spec.Dispute.Function])
		timeline, err := resolveTimeline(spec, totalGas, timelines)
		if err != nil {
			return nil, err
		}
		out = append(out, ProtocolResult{
			Scheme:            spec.Scheme,
			Mechanism:         spec.Mechanism,
			Family:            spec.Family,
			OnchainRounds:     spec.OnchainRounds,
			OptimisticGasK:    optimistic.GasK,
			DisputeGasK:       dispute.GasK,
			Optimistic:        optimistic,
			Dispute:           dispute,
			DisputeFlow:       append([]string(nil), spec.DisputeFlow...),
			TimelineSlots:     timeline.Slots,
			Timeline:          timeline,
			Concurrent:        spec.Concurrent,
			ModelScope:        spec.ModelScope,
			SourceKind:        spec.SourceKind,
			ReferenceSources:  append([]ReferenceSource(nil), spec.ReferenceSources...),
			MeasurementMethod: spec.MeasurementMethod,
			Reproduce:         fmt.Sprintf("forge test --gas-report; parse %s and %s from test/Benchmarks.t.sol", spec.Optimistic.Test, spec.Dispute.Test),
			Notes:             "direct Foundry gas from src/DisputeProtocolBenchmarks.sol; each benchmark executes the local submit/challenge/localize/adjudicate path and records the benchmark function, arguments, and reproduction command",
			ValidationClaim:   "the plotted comparison value is the parsed EVM average gas for the named Solidity benchmark function and its recorded dispute_flow",
		})
	}
	return out, nil
}

func BuildTimeline(p TimelineParams, totalGas float64) []TimelineResult {
	out, err := BuildTimelineWithTimelines(p, totalGas, nil)
	if err != nil {
		panic(err)
	}
	return out
}

func BuildTimelineWithTimelines(p TimelineParams, totalGas float64, timelines map[string]TimelineDerivation) ([]TimelineResult, error) {
	tReexec := 1 - p.Eta
	byScheme := map[string]ProtocolSpec{}
	for _, spec := range Specs(totalGas) {
		byScheme[spec.Scheme] = spec
	}
	order := []string{"CleVer\n(ours)", "Arbitrum\nBoLD", "Cartesi\nDave", "TrueBit", "Arbitrum\nClassic"}
	out := make([]TimelineResult, 0, len(order))
	for _, scheme := range order {
		spec := byScheme[scheme]
		derivation, err := resolveTimeline(spec, totalGas, timelines)
		if err != nil {
			return nil, err
		}
		name := spec.Scheme
		total := p.TExec + p.Gamma + float64(derivation.Slots)*p.TSlot + tReexec
		formula := "T_exec + gamma + dispute_slots*T_slot + T_reexec"
		if spec.Concurrent {
			name = "CleVer\n（并发）"
			total = p.Eta*p.TExec + p.TSeg + float64(derivation.Slots)*p.TSlot + tReexec
			formula = "eta*T_exec + T_seg + dispute_slots*T_slot + T_reexec"
		}
		out = append(out, TimelineResult{Name: name, DisputeSlots: derivation.Slots, TotalTime: total, Formula: formula, Derivation: derivation})
	}
	return out, nil
}

func DeriveTimeline(spec ProtocolSpec, totalGas float64) TimelineDerivation {
	return RunTimelineSimulation(spec, totalGas)
}

func resolveTimeline(spec ProtocolSpec, totalGas float64, timelines map[string]TimelineDerivation) (TimelineDerivation, error) {
	if timelines == nil {
		return DeriveTimeline(spec, totalGas), nil
	}
	timeline, ok := timelines[spec.Scheme]
	if !ok {
		return TimelineDerivation{}, fmt.Errorf("missing timeline simulation result for %s", spec.Scheme)
	}
	if err := validateTimelineDerivation(spec, totalGas, timeline); err != nil {
		return TimelineDerivation{}, err
	}
	return timeline, nil
}

func validateTimelineDerivation(spec ProtocolSpec, totalGas float64, timeline TimelineDerivation) error {
	if timeline.SimulationKind != "local_dispute_timeline_state_machine" || !timeline.Resolved {
		return fmt.Errorf("%s timeline is not a resolved local simulation", spec.Scheme)
	}
	if timeline.Strategy != spec.Timeline.Strategy {
		return fmt.Errorf("%s timeline strategy = %s, want %s", spec.Scheme, timeline.Strategy, spec.Timeline.Strategy)
	}
	if timeline.LogicalRounds != spec.OnchainRounds {
		return fmt.Errorf("%s timeline logical rounds = %d, want %d", spec.Scheme, timeline.LogicalRounds, spec.OnchainRounds)
	}
	if math.Abs(timeline.TotalGas-totalGas) > 1e-6 {
		return fmt.Errorf("%s timeline total gas = %.0f, want %.0f", spec.Scheme, timeline.TotalGas, totalGas)
	}
	if timeline.Slots <= 0 || len(timeline.Events) != timeline.Slots {
		return fmt.Errorf("%s timeline slots/events mismatch: slots=%d events=%d", spec.Scheme, timeline.Slots, len(timeline.Events))
	}
	if timeline.Formula == "" {
		return fmt.Errorf("%s timeline formula missing", spec.Scheme)
	}
	for _, event := range timeline.Events {
		if event.Transition == "" || event.StateBefore == nil || event.StateAfter == nil {
			return fmt.Errorf("%s timeline event %d missing state transition evidence", spec.Scheme, event.Slot)
		}
	}
	return nil
}

func RunTimelineSimulation(spec ProtocolSpec, totalGas float64) TimelineDerivation {
	logicalRounds := spec.OnchainRounds
	if logicalRounds <= 0 {
		logicalRounds = int(math.Ceil(math.Log2(totalGas)))
	}
	events := []TimelineEvent{}
	appendEvent := func(phase string, start, end int, detail string, before, after map[string]int, transition string) {
		events = append(events, TimelineEvent{
			Slot:              len(events) + 1,
			Phase:             phase,
			LogicalRoundStart: start,
			LogicalRoundEnd:   end,
			Detail:            detail,
			StateBefore:       before,
			StateAfter:        after,
			Transition:        transition,
		})
	}
	model := spec.Timeline
	formula := ""
	switch model.Strategy {
	case "one_slot_per_remaining_bisection_round", "solver_verifier_remaining_bisection_rounds", "tournament_then_remaining_bisection_rounds":
		completedOffchain := min(max(model.InitialOffchainRounds, 0), logicalRounds)
		remaining := logicalRounds - completedOffchain
		currentRound := completedOffchain + 1
		for remaining > 0 {
			before := map[string]int{
				"completed_logical_rounds": currentRound - 1,
				"remaining_onchain_rounds": remaining,
			}
			after := map[string]int{
				"completed_logical_rounds": currentRound,
				"remaining_onchain_rounds": remaining - 1,
			}
			appendEvent(
				"interactive_bisection_deadline",
				currentRound,
				currentRound,
				fmt.Sprintf("state-machine slot executes logical bisection round %d after %d off-chain/pre-deadline rounds", currentRound, completedOffchain),
				before,
				after,
				"consume one on-chain deadline slot and advance one bisection round",
			)
			currentRound++
			remaining--
		}
		formula = fmt.Sprintf("max(0, logical_rounds(%d)-initial_offchain_rounds(%d))", logicalRounds, model.InitialOffchainRounds)
	case "bounded_liquidity_parallel_levels":
		perSlot := model.RoundsPerSlot
		if perSlot <= 0 {
			perSlot = 1
		}
		remaining := logicalRounds
		nextRound := 1
		for remaining > 0 {
			batch := min(perSlot, remaining)
			start := nextRound
			end := nextRound + batch - 1
			before := map[string]int{
				"remaining_logical_rounds": remaining,
				"rounds_per_slot":          perSlot,
			}
			remaining -= batch
			nextRound += batch
			after := map[string]int{
				"remaining_logical_rounds": remaining,
				"rounds_per_slot":          perSlot,
			}
			appendEvent(
				"parallel_level_bisection",
				start,
				end,
				fmt.Sprintf("BoLD state machine narrows rounds %d-%d in one parallel edge deadline slot", start, end),
				before,
				after,
				"consume one deadline slot and advance a parallel batch of dispute rounds",
			)
		}
		boundaryRemaining := model.LevelBoundarySlots
		for i := 0; i < model.LevelBoundarySlots; i++ {
			level := (i % max(1, model.Levels)) + 1
			before := map[string]int{
				"level":                    level,
				"boundary_slots_remaining": boundaryRemaining,
			}
			boundaryRemaining--
			after := map[string]int{
				"level":                    level,
				"boundary_slots_remaining": boundaryRemaining,
			}
			appendEvent(
				"level_boundary_confirmation",
				0,
				0,
				fmt.Sprintf("BoLD state machine confirms level %d boundary", level),
				before,
				after,
				"consume one level-boundary confirmation slot",
			)
		}
		formula = fmt.Sprintf("ceil(logical_rounds(%d)/rounds_per_slot(%d))+level_boundary_slots(%d)", logicalRounds, perSlot, model.LevelBoundarySlots)
	case "concurrent_two_layer_slice_then_verseg":
		phases := []string{"outer_segment_localization", "inner_subsegment_localization", "bounded_verseg_replay"}
		count := model.ConcurrentSlots
		if count <= 0 {
			count = len(phases)
		}
		for i := 0; i < count; i++ {
			phase := phases[min(i, len(phases)-1)]
			before := map[string]int{
				"completed_phases": i,
				"remaining_phases": count - i,
			}
			after := map[string]int{
				"completed_phases": i + 1,
				"remaining_phases": count - i - 1,
			}
			appendEvent(
				phase,
				i+1,
				i+1,
				fmt.Sprintf("CleVer state machine executes concurrent verification phase %d", i+1),
				before,
				after,
				"consume one concurrent localization/replay slot",
			)
		}
		formula = fmt.Sprintf("concurrent protocol phases(%d)", count)
	default:
		panic(fmt.Sprintf("unknown timeline strategy %q for %s", model.Strategy, spec.Scheme))
	}
	return TimelineDerivation{
		Strategy:       model.Strategy,
		TotalGas:       totalGas,
		LogicalRounds:  logicalRounds,
		Slots:          len(events),
		Events:         events,
		Formula:        formula,
		SimulationKind: "local_dispute_timeline_state_machine",
		Inputs: TimelineSimulationInput{
			TotalGas:              totalGas,
			LogicalRounds:         logicalRounds,
			Strategy:              model.Strategy,
			InitialOffchainRounds: model.InitialOffchainRounds,
			RoundsPerSlot:         model.RoundsPerSlot,
			Levels:                model.Levels,
			LevelBoundarySlots:    model.LevelBoundarySlots,
			ConcurrentSlots:       model.ConcurrentSlots,
		},
		Resolved: true,
	}
}

func RequiredFunctions(totalGas float64) []string {
	specs := Specs(totalGas)
	functions := make([]string, 0, len(specs)*2)
	for _, spec := range specs {
		functions = append(functions, spec.Optimistic.Function, spec.Dispute.Function)
	}
	sort.Strings(functions)
	return functions
}

func missingFunctions(specs []ProtocolSpec, gasByFunction map[string]uint64) []string {
	missing := []string{}
	for _, spec := range specs {
		for _, path := range []PathSpec{spec.Optimistic, spec.Dispute} {
			if gasByFunction[path.Function] == 0 {
				missing = append(missing, path.Function)
			}
		}
	}
	sort.Strings(missing)
	return missing
}

func pathResult(spec PathSpec, gas uint64) PathResult {
	return PathResult{
		Function: spec.Function,
		Test:     spec.Test,
		Args:     spec.Args,
		Gas:      gas,
		GasK:     float64(gas) / 1000,
	}
}
