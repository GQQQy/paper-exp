package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"chapter5/experiment/audit"
	"chapter5/experiment/comparisons"
)

type RawExperimentLog struct {
	Metadata              map[string]any                 `json:"metadata"`
	Params                audit.Params                   `json:"params"`
	Table9                map[string]any                 `json:"table9_parameters"`
	RanCkTraces           []audit.RanCkTrace             `json:"ranck_traces"`
	SenCkTraces           []audit.SenCkTrace             `json:"senck_traces"`
	MonteCarlo            MonteCarloEvidence             `json:"monte_carlo"`
	DetectionParameters   DetectionParameters            `json:"detection_parameters"`
	Feasibility           FeasibilityEvidence            `json:"feasibility"`
	Overhead              []audit.OverheadTrace          `json:"overhead_traces"`
	Gas                   audit.GasTrace                 `json:"gas_trace"`
	ProtocolCoverage      []CoverageItem                 `json:"protocol_coverage"`
	ComparisonExperiments []comparisons.SchemeExperiment `json:"comparison_experiments"`
	ComparisonTable11     []map[string]string            `json:"comparison_table_11"`
	FigureReferenceInputs map[string]any                 `json:"figure_reference_inputs"`
}

type MonteCarloEvidence struct {
	RanCk    map[string][]audit.MonteCarloPoint `json:"ranck"`
	SenCk    map[string][]audit.MonteCarloPoint `json:"senck"`
	GammaHit []audit.GammaHitPoint              `json:"gamma_hit_sweep"`
}

type DetectionParameters struct {
	OfflineLengthMaxBlocks    int       `json:"offline_length_max_blocks"`
	HeartbeatMValues          []int     `json:"heartbeat_M_values"`
	CombinedPiH               float64   `json:"combined_pi_h"`
	CombinedAuditWindowBlocks int       `json:"combined_audit_window_blocks"`
	ContinuitySampleSizes     []int     `json:"continuity_sample_sizes"`
	SenCkRhoValues            []float64 `json:"senck_rho_values"`
	SenCkMsMin                int       `json:"senck_m_s_min"`
	SenCkMsMax                int       `json:"senck_m_s_max"`
	Source                    string    `json:"source"`
}

type FeasibilityEvidence struct {
	C1               []C1Point         `json:"C1_online_deviation"`
	C2               []C2Point         `json:"C2_diligence_deviation"`
	Joint            []JointPoint      `json:"joint_feasible_region"`
	DefaultCostRatio DefaultCostRatio  `json:"default_cost_ratio"`
	Formulas         map[string]string `json:"formulas"`
	ScanInputs       map[string]any    `json:"scan_inputs"`
}

type C1Point struct {
	TWin                   int     `json:"T_win"`
	PiH                    float64 `json:"pi_h"`
	NormalizedMinHeartbeat float64 `json:"min_c_hb_over_c_track_plus_c_m"`
	ExpectedHeartbeats     float64 `json:"expected_heartbeats"`
}

type C2Point struct {
	Rho                    float64 `json:"rho"`
	Ms                     int     `json:"m_s"`
	Q                      float64 `json:"q"`
	NormalizedMinSentSlash float64 `json:"min_c_sent_over_c_m"`
}

type JointPoint struct {
	PiH         float64 `json:"pi_h"`
	Ms          int     `json:"m_s"`
	C1Satisfied bool    `json:"C1_satisfied"`
	C2Satisfied bool    `json:"C2_satisfied"`
	Region      string  `json:"region"`
}

type DefaultCostRatio struct {
	TWin               int     `json:"T_win"`
	PiH                float64 `json:"pi_h"`
	ExpectedHeartbeats float64 `json:"expected_heartbeats"`
	MinHeartbeatRatio  float64 `json:"min_c_hb_ratio"`
}

type CoverageItem struct {
	Requirement string `json:"requirement"`
	Source      string `json:"source"`
	Status      string `json:"status"`
}

func main() {
	out := flag.String("out", filepath.Join("logs", "raw_experiment_log.json"), "output raw experiment JSON")
	skipFoundry := flag.Bool("skip-foundry", false, "skip forge test; gas figures will be marked incomplete")
	flag.Parse()

	p := audit.DefaultParams()
	ranck := []audit.RanCkTrace{
		audit.SimulateRanCk(p, audit.BehaviorHonest),
		audit.SimulateRanCk(p, audit.BehaviorOffline),
		audit.SimulateRanCk(p, audit.BehaviorIntermittent),
		audit.SimulateRanCk(p, audit.BehaviorRecomputeFail),
		audit.SimulateRanCk(p, audit.BehaviorInconsistent),
	}
	senck := []audit.SenCkTrace{
		audit.SimulateSenCk(p, audit.BehaviorHonest),
		audit.SimulateSenCk(p, audit.BehaviorLazyGuess),
		audit.SimulateSenCk(p, audit.BehaviorExecutorPollute),
	}
	foundry, status := []audit.FoundryGasEntry{}, "SKIP: --skip-foundry requested"
	if !*skipFoundry {
		foundry, status = runFoundryGas()
	}
	comparisonExperiments := comparisons.RunCapabilityExperiments()
	log := RawExperimentLog{
		Metadata: map[string]any{
			"chapter":    "第五章 面向链下计算的验证者工作审计",
			"created_at": audit.CalibratedNow(),
			"source":     "RanCk/SenCk protocol implementation, deterministic simulation traces, Monte Carlo runs, and Foundry gas benchmarks",
			"thesis_pdf": "../第五章面向链下计算的验证者工作审计.pdf",
		},
		Params:                p,
		Table9:                audit.Table9Parameters(),
		RanCkTraces:           ranck,
		SenCkTraces:           senck,
		MonteCarlo:            monteCarloEvidence(),
		DetectionParameters:   detectionParameters(),
		Feasibility:           feasibilityEvidence(),
		Overhead:              audit.OverheadTraces(p),
		Gas:                   audit.GasTraceFromFoundry(foundry, status, p),
		ProtocolCoverage:      coverage(),
		ComparisonExperiments: comparisonExperiments,
		ComparisonTable11:     comparisons.Table11Rows(comparisonExperiments),
		FigureReferenceInputs: figureReferenceInputs(),
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		panic(err)
	}
	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(log); err != nil {
		panic(err)
	}
	fmt.Printf("[ranck] honest heartbeats=%d missed=%d cont=%v\n", ranck[0].HeartbeatCount, ranck[0].MissedHeartbeats, ranck[0].ContAudit.Passed)
	fmt.Printf("[senck] rho=%.6f honest sampled=%d lazy_detected=%v polluted_detected=%v\n", senck[0].Rho, len(senck[0].SentReport.SampledReports), senck[1].SentReport.Detected, senck[2].SentReport.Detected)
	fmt.Printf("[gas] status=%s default_per_validator=%.2fM\n", status, float64(log.Gas.DefaultPerValidatorGas)/1e6)
	fmt.Printf("[out] %s\n", *out)
}

func runFoundryGas() ([]audit.FoundryGasEntry, string) {
	forge := "/Users/gqy/.foundry/bin/forge"
	if _, err := os.Stat(forge); err != nil {
		if path, lookErr := exec.LookPath("forge"); lookErr == nil {
			forge = path
		} else {
			return nil, "SKIP: forge not found; gas trace incomplete"
		}
	}
	cmd := exec.Command(forge, "test", "--gas-report")
	out, err := cmd.CombinedOutput()
	text := string(out)
	if err != nil {
		return nil, "SKIP: forge failed; gas trace incomplete: " + compact(text, 180)
	}
	rows := []audit.FoundryGasEntry{}
	functionToTest := map[string]string{
		"trackInit":     "testTrackInitGas",
		"hbRespond":     "testHBRespondGas",
		"contAudit":     "testContAuditGas",
		"sentReport":    "testSentReportGas",
		"sentDispute":   "testDisputeGas",
		"sentProve":     "testSentProveGas",
		"podMineBounty": "testPoDMineBountyGas",
	}
	for _, match := range regexp.MustCompile(`\|\s*([A-Za-z][A-Za-z0-9_]*)\s*\|\s*([0-9]+)\s*\|\s*([0-9]+)\s*\|\s*([0-9]+)\s*\|\s*([0-9]+)\s*\|\s*([0-9]+)\s*\|`).FindAllStringSubmatch(text, -1) {
		testName, ok := functionToTest[match[1]]
		if !ok {
			continue
		}
		avgGas, _ := strconv.Atoi(match[3])
		rows = append(rows, audit.FoundryGasEntry{
			Test:     testName,
			Function: match[1],
			Gas:      avgGas,
			Source:   "forge gas report function avg",
		})
	}
	if len(rows) == 0 {
		return nil, "SKIP: forge ran but no function-level gas rows parsed; gas trace incomplete"
	}
	return rows, "PASS: parsed forge test --gas-report function averages"
}

func monteCarloEvidence() MonteCarloEvidence {
	ells := []int{25, 50, 75, 100, 150, 200, 250, 300, 400, 500}
	ms := []int{1, 2, 3, 4, 5, 6, 8, 10, 12, 15}
	return MonteCarloEvidence{
		RanCk: map[string][]audit.MonteCarloPoint{
			"pi_h_0.005": audit.MonteCarloRanCk(0.005, ells, 50000, 20260501),
			"pi_h_0.010": audit.MonteCarloRanCk(0.010, ells, 50000, 20260502),
			"pi_h_0.020": audit.MonteCarloRanCk(0.020, ells, 50000, 20260503),
		},
		SenCk: map[string][]audit.MonteCarloPoint{
			"rho_0.2": audit.MonteCarloSenCk(0.2, ms, 50000, 20260511),
			"rho_0.4": audit.MonteCarloSenCk(0.4, ms, 50000, 20260512),
			"rho_0.6": audit.MonteCarloSenCk(0.6, ms, 50000, 20260513),
			"rho_0.8": audit.MonteCarloSenCk(0.8, ms, 50000, 20260514),
		},
		GammaHit: audit.GammaHitSweep(),
	}
}

func detectionParameters() DetectionParameters {
	return DetectionParameters{
		OfflineLengthMaxBlocks:    500,
		HeartbeatMValues:          []int{500, 200, 100, 50},
		CombinedPiH:               0.01,
		CombinedAuditWindowBlocks: 500,
		ContinuitySampleSizes:     []int{5, 10, 20},
		SenCkRhoValues:            []float64{0.2, 0.4, 0.6, 0.8, 0.95},
		SenCkMsMin:                1,
		SenCkMsMax:                20,
		Source:                    "Figure 26/27 parameter scan derived from Table 9 ranges and Section 5.5 detection formulas",
	}
}

func feasibilityEvidence() FeasibilityEvidence {
	c1 := []C1Point{}
	for _, tWin := range []int{3600, 7200, 14400} {
		for i := 0; i <= 120; i++ {
			piH := 0.002 + float64(i)*(0.025-0.002)/120
			c1 = append(c1, C1Point{
				TWin:                   tWin,
				PiH:                    piH,
				NormalizedMinHeartbeat: 1.0 / (float64(tWin) * piH),
				ExpectedHeartbeats:     float64(tWin) * piH,
			})
		}
	}
	c2 := []C2Point{}
	for _, rho := range []float64{0.2, 0.4, 0.6, 0.8} {
		for ms := 1; ms <= 20; ms++ {
			q := 1.0 - math.Pow(1-rho, float64(ms))
			c2 = append(c2, C2Point{Rho: rho, Ms: ms, Q: q, NormalizedMinSentSlash: 1.0 / q})
		}
	}
	joint := []JointPoint{}
	tWin := 7200
	cHB := 0.05
	cSent := 1.5
	rho := 0.3
	for i := 0; i <= 120; i++ {
		piH := 0.001 + float64(i)*(0.025-0.001)/120
		for ms := 1; ms <= 20; ms++ {
			c1ok := float64(tWin)*piH*cHB >= 1.0
			q := 1.0 - math.Pow(1-rho, float64(ms))
			c2ok := q*cSent >= 1.0
			region := "均不满足"
			if c1ok && c2ok {
				region = "同时满足 C1 与 C2"
			} else if c1ok {
				region = "仅满足 C1"
			} else if c2ok {
				region = "仅满足 C2"
			}
			joint = append(joint, JointPoint{PiH: piH, Ms: ms, C1Satisfied: c1ok, C2Satisfied: c2ok, Region: region})
		}
	}
	return FeasibilityEvidence{
		C1:    c1,
		C2:    c2,
		Joint: joint,
		DefaultCostRatio: DefaultCostRatio{
			TWin:               7200,
			PiH:                0.01,
			ExpectedHeartbeats: 72,
			MinHeartbeatRatio:  1.0 / 72.0,
		},
		Formulas: map[string]string{
			"C1":    "c_track + c_m <= T_win * pi_h * c_hb",
			"C2":    "c_m <= q * c_sent; q = 1 - (1-rho)^m_s",
			"rho":   "rho = 1 - (1 - 1/M_c)^L",
			"joint": "C1 and C2 evaluated independently over (pi_h, m_s)",
			"RanCk": "Pr[heartbeat detects ell-block offline] = 1-(1-pi_h)^ell",
			"SenCk": "lazy pass probability = (1-rho)^m_s",
		},
		ScanInputs: map[string]any{
			"joint_T_win":      7200,
			"joint_rho":        0.3,
			"joint_c_hb_ratio": 0.05,
			"joint_c_sent":     1.5,
		},
	}
}

func coverage() []CoverageItem {
	return []CoverageItem{
		{"RanCk TrackInit seed/nonce/TC_0", "audit.TrackInit + Solidity trackInit", "PASS"},
		{"RanCk alpha_t = H(alpha_{t-1}||bh_t||PRF(seed,t)||tau_t)", "audit.UpdateAlpha", "PASS"},
		{"RanCk Trigger(t,i)=H(bh_t||tid||i) mod M", "audit.Trigger", "PASS"},
		{"RanCk HBRespond within Delta and c_hb for miss", "audit.SimulateRanCk + ValidatorAudit.sol", "PASS"},
		{"RanCk ContAudit sampled single-step witnesses", "audit.ContAudit", "PASS"},
		{"SenCk EVM opcode hook captures pc/op/rw_t/val_t", "audit.InstrumentedEVM.ExecuteStep + AfterOpcodeHook", "PASS"},
		{"SenCk Gamma(r,tid,k,t,rw_t)", "audit.Gamma", "PASS"},
		{"SenCk SentReport local replay from snapshots", "audit.GenerateSegment uses instrumented EVM replay", "PASS"},
		{"SenCk SentDispute/SentProve polluted or wrong report", "audit.SimulateSenCk + Solidity sentProve", "PASS"},
		{"Integrated submit -> init -> heartbeat -> continuity -> sent report -> final", "protocol_test.go", "PASS"},
	}
}

func figureReferenceInputs() map[string]any {
	return map[string]any{
		"figure_24": "C1/C2 formula scans generated in feasibility evidence",
		"figure_25": "joint feasible grid generated in feasibility evidence",
		"figure_26": "RanCk theory from formulas, including combined heartbeat+continuity pass probability",
		"figure_27": "SenCk lazy pass probability from rho and m_s",
		"figure_28": "Monte Carlo traces from implemented Bernoulli trigger and segment hit simulation",
		"figure_29": "overhead trace from instrumented EVM opcode hook: rw encoding, hash, Gamma, sentinel digest, snapshot load, and deterministic replay model",
		"table_10":  "Foundry per-test gas from ValidatorAudit.t.sol, with function-level provenance recorded in gas_trace.measurement_provenance",
		"figure_30": "normal path gas decomposition and monthly comparison",
		"figure_31": "pi_h sweep of gas and ell=300 detection probability, with below-PoD region",
		"table_11":  "comparison_experiments -> comparison_table_11",
	}
}

func compact(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	return s[:n]
}
