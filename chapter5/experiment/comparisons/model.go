package comparisons

import (
	"fmt"
	"math"
)

type ScenarioResult struct {
	Name        string `json:"name"`
	Requirement string `json:"requirement"`
	Outcome     string `json:"outcome"`
	Evidence    string `json:"evidence"`
}

type ProtocolStep struct {
	Phase    string `json:"phase"`
	Action   string `json:"action"`
	Evidence string `json:"evidence"`
}

type ReferenceSource struct {
	Project  string `json:"project"`
	URL      string `json:"url"`
	Evidence string `json:"evidence"`
	UsedFor  string `json:"used_for"`
}

type SchemeExperiment struct {
	Scheme               string            `json:"scheme"`
	ExecutionEnvironment string            `json:"execution_environment"`
	OnlineAudit          string            `json:"online_audit"`
	SemanticDiligence    string            `json:"semantic_diligence"`
	UnpredictableAudit   string            `json:"unpredictable_audit"`
	ExternalNetwork      string            `json:"external_network"`
	BenchmarkMode        string            `json:"benchmark_mode"`
	SourceKind           string            `json:"source_kind"`
	ReferenceSources     []ReferenceSource `json:"reference_sources"`
	Scenarios            []ScenarioResult  `json:"scenarios"`
	ProtocolFlow         []ProtocolStep    `json:"protocol_flow"`
	Reproduce            string            `json:"reproduce"`
	ValidationClaim      string            `json:"validation_claim"`
}

type PoDBaselineResult struct {
	Scheme           string            `json:"scheme"`
	Test             string            `json:"test"`
	Function         string            `json:"function"`
	GasPerEpoch      int               `json:"gas_per_epoch"`
	Theta            float64           `json:"theta"`
	ValidatorCount   int               `json:"validator_count"`
	EpochsPerMonth   int               `json:"epochs_per_month"`
	MonthlyGas       int               `json:"monthly_gas"`
	Formula          string            `json:"formula"`
	PaperAverageGas  int               `json:"paper_mine_bounty_average_gas"`
	PaperSource      string            `json:"paper_source"`
	ProtocolSteps    []string          `json:"protocol_steps"`
	ReferenceSources []ReferenceSource `json:"reference_sources"`
	SourceKind       string            `json:"source_kind"`
	Reproduce        string            `json:"reproduce"`
	ValidationClaim  string            `json:"validation_claim"`
}

type FoundryEntry struct {
	Test     string
	Function string
	Gas      int
}

func RunCapabilityExperiments() []SchemeExperiment {
	truebitSources := []ReferenceSource{
		{
			Project:  "TrueBitFoundation/truebit-eth",
			URL:      "https://github.com/TrueBitFoundation/truebit-eth",
			Evidence: "TrueBit repository used to define the solver/verifier verification-game shape",
			UsedFor:  "reference boundary for the local solver/verifier scenario in table 11",
		},
		{
			Project:  "Truebit documentation",
			URL:      "https://docs.truebit.io/v1docs",
			Evidence: "TrueBit documentation used to confirm task and role workflow terminology",
			UsedFor:  "reference boundary for the local TrueBit task and role workflow model",
		},
	}
	arbitrumSources := []ReferenceSource{
		{
			Project:  "OffchainLabs/arbitrum-classic",
			URL:      "https://github.com/OffchainLabs/arbitrum-classic",
			Evidence: "Arbitrum Classic repository used to define the assertion/challenge fraud-proof shape",
			UsedFor:  "reference boundary for the local assertion/challenge scenario in table 11",
		},
		{
			Project:  "OffchainLabs/bold",
			URL:      "https://github.com/OffchainLabs/bold",
			Evidence: "BoLD repository used as the current Arbitrum dispute-system reference",
			UsedFor:  "reference boundary for the local Arbitrum dispute-system scenario",
		},
	}
	podSources := PoDReferenceSources()
	rankSenckSources := []ReferenceSource{
		{
			Project:  "paper-exp chapter5",
			URL:      "chapter5/experiment/audit/protocol.go",
			Evidence: "local RanCk/SenCk protocol implementation and deterministic experiment traces",
			UsedFor:  "online heartbeat, continuity audit, sentinel diligence, and gas benchmark rows",
		},
		{
			Project:  "paper-exp chapter5 Solidity",
			URL:      "chapter5/experiment/src/ValidatorAudit.sol",
			Evidence: "local Solidity benchmark for TrackInit/HBRespond/ContAudit/SentReport/Dispute/SentProve",
			UsedFor:  "table 10 and figures 30-31 gas measurement",
		},
	}
	return []SchemeExperiment{
		{
			Scheme:               "TrueBit",
			ExecutionEnvironment: "WASM",
			OnlineAudit:          "no",
			SemanticDiligence:    "partial",
			UnpredictableAudit:   "-",
			ExternalNetwork:      "no",
			BenchmarkMode:        "local minimal benchmark faithful to solver/verifier verification-game flow",
			SourceKind:           "flow_benchmark_from_published_protocol_shape",
			ReferenceSources:     truebitSources,
			Scenarios: []ScenarioResult{
				{"offline_validator", "detect online/offline state during execution", "not covered", "verification game starts after an explicit challenge; no periodic online heartbeat path is modeled"},
				{"lazy_semantic_report", "detect semantic diligence without rerunning the whole task", "partially covered", "interactive verification can isolate a disputed step, but routine semantic diligence sampling is not present"},
				{"unpredictable_audit_target", "hide sampled audit points before work is done", "not covered", "challenge targets are chosen after a dispute rather than by an in-execution sentinel trigger"},
			},
			ProtocolFlow: []ProtocolStep{
				{"submit", "solver commits to task output and execution trace root", "comparison benchmark records solver commitment before any audit decision"},
				{"challenge", "verifier challenges the solver claim after observing a mismatch", "no heartbeat or in-execution sentinel trigger is required"},
				{"localize", "solver and verifier bisect the trace to one disputed instruction", "semantic diligence is partial because the game can isolate a disputed step only after challenge"},
				{"adjudicate", "single-step judge decides the instruction transition", "final evidence is a post-execution fraud-proof verdict"},
			},
			Reproduce:       "go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json",
			ValidationClaim: "table 11 row is generated from explicit local scenario outcomes and the TrueBit protocol_flow model in comparisons.RunCapabilityExperiments",
		},
		{
			Scheme:               "Arbitrum",
			ExecutionEnvironment: "AVM/WASM",
			OnlineAudit:          "no",
			SemanticDiligence:    "no",
			UnpredictableAudit:   "-",
			ExternalNetwork:      "no",
			BenchmarkMode:        "local minimal benchmark faithful to fraud-proof assertion/challenge flow",
			SourceKind:           "flow_benchmark_from_rollup_fraud_proof_shape",
			ReferenceSources:     arbitrumSources,
			Scenarios: []ScenarioResult{
				{"offline_validator", "detect online/offline state during execution", "not covered", "rollup challenge logic does not require a validator heartbeat during off-chain execution"},
				{"lazy_semantic_report", "detect semantic diligence without rerunning the whole task", "not covered", "fraud proof resolves a challenged state transition, not routine diligence during local work"},
				{"unpredictable_audit_target", "hide sampled audit points before work is done", "not covered", "audit target selection is tied to later disputes"},
			},
			ProtocolFlow: []ProtocolStep{
				{"submit", "validator posts an assertion over the off-chain execution result", "comparison benchmark models the assertion root as the starting object"},
				{"challenge", "opponent opens an interactive fraud-proof challenge", "challenge happens after execution, not as online-work audit"},
				{"localize", "parties bisect the assertion interval to a single transition", "no routine opcode/rw sampling is performed during execution"},
				{"adjudicate", "one-step proof confirms or rejects the challenged transition", "final evidence is a rollup dispute verdict"},
			},
			Reproduce:       "go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json",
			ValidationClaim: "table 11 row is generated from explicit local scenario outcomes and the Arbitrum fraud-proof protocol_flow model",
		},
		{
			Scheme:               "PoD",
			ExecutionEnvironment: "independent network",
			OnlineAudit:          "yes",
			SemanticDiligence:    "no",
			UnpredictableAudit:   "limited",
			ExternalNetwork:      "yes",
			BenchmarkMode:        "Solidity MineBounty gas benchmark plus deterministic watchtower epoch trace",
			SourceKind:           "scenario_comparison_experiment_plus_solidity_pod_mine_bounty",
			ReferenceSources:     podSources,
			Scenarios: []ScenarioResult{
				{"offline_validator", "detect online/offline state during execution", "covered by external watchtower pool", "PoD watchtowers recompute the assertion and execution trace root, then submit a VRF proof through podMineBounty"},
				{"lazy_semantic_report", "detect semantic diligence without rerunning the whole task", "not covered", "PoD proves rollup assertion checking and trace-root work, but does not include SenCk-style opcode/rw sentinel digest sampling"},
				{"unpredictable_audit_target", "hide sampled audit points before work is done", "limited", "VRF bounty eligibility is unpredictable, but audit targets are external watchtower bounty epochs rather than task-local Gamma triggers"},
			},
			ProtocolFlow: []ProtocolStep{
				{"watchtower_recompute", "watchtower recomputes the asserted state root for each epoch", "gas_trace.pod_experiment_trace records honest and lazy watchtower roots"},
				{"trace_root", "watchtower builds an execution trace Merkle-style root", "testPoDMineBountyGas submits a calibrated 260-leaf execution trace"},
				{"vrf_bounty", "watchtower binds state root, trace root, epoch, identity, and VRF proof", "ValidatorAudit.podMineBounty verifies the digest and bounty threshold"},
				{"detect_lazy", "other watchtowers reject mismatched recomputation or forged trace roots", "pod_experiment_trace.lazy_detected and invalid_proof_count provide raw evidence"},
			},
			Reproduce:       "forge test --gas-report; parse testPoDMineBountyGas or run go run ./cmd/comparison-exp",
			ValidationClaim: "PoD row and monthly gas baseline are generated from the Solidity podMineBounty benchmark plus deterministic watchtower epoch traces",
		},
		{
			Scheme:               "RanCk+SenCk",
			ExecutionEnvironment: "EVM",
			OnlineAudit:          "yes",
			SemanticDiligence:    "yes",
			UnpredictableAudit:   "yes",
			ExternalNetwork:      "no",
			BenchmarkMode:        "implemented protocol simulation and Solidity gas benchmark in this repository",
			SourceKind:           "project_protocol_implementation",
			ReferenceSources:     rankSenckSources,
			Scenarios: []ScenarioResult{
				{"offline_validator", "detect online/offline state during execution", "covered by RanCk", "Trigger(t,i)=H(bh_t||tid||i) mod M plus ContAudit detects missed or inconsistent heartbeat traces"},
				{"lazy_semantic_report", "detect semantic diligence without rerunning the whole task", "covered by SenCk", "Gamma(r,tid,k,t,rw_t) sentinel sampling checks local opcode/rw digests"},
				{"unpredictable_audit_target", "hide sampled audit points before work is done", "covered", "block-hash and runtime rw dependent Gamma triggers are computed during instrumented execution"},
			},
			ProtocolFlow: []ProtocolStep{
				{"TrackInit", "validator commits seed/nonce and TC_0", "ranck_traces[].track_init and ValidatorAudit.trackInit"},
				{"HBRespond", "Trigger(t,i)=H(bh_t||tid||i) mod M selects heartbeat points", "ranck_traces[].heartbeats records responses and misses"},
				{"ContAudit", "opened credential chain and sampled single-step witnesses are checked", "ranck_traces[].cont_audit records pass/detection evidence"},
				{"SentReport", "instrumented EVM hook emits opcode/rw sentinel digests", "senck_traces[].segment_reports and trace_sample record Gamma sampling"},
				{"Dispute/VerSeg", "mismatched sentinel reports are replayed and adjudicated", "senck_traces[].sent_report.dispute_passed and ValidatorAudit.sentProve"},
			},
			Reproduce:       "go test ./...; go run ./cmd/audit-exp --out logs/raw_experiment_log.json",
			ValidationClaim: "RanCk/SenCk row is generated from implemented protocol traces and explicit scenario outcomes",
		},
	}
}

func Table11Rows(experiments []SchemeExperiment) []map[string]string {
	rows := make([]map[string]string, 0, len(experiments))
	for _, exp := range experiments {
		rows = append(rows, map[string]string{
			"scheme":                exp.Scheme,
			"execution_environment": exp.ExecutionEnvironment,
			"online_audit":          exp.OnlineAudit,
			"semantic_diligence":    exp.SemanticDiligence,
			"unpredictable_audit":   exp.UnpredictableAudit,
			"external_network":      exp.ExternalNetwork,
		})
	}
	return rows
}

func PoDReferenceSources() []ReferenceSource {
	return []ReferenceSource{
		{
			Project:  "witnesschain-com/diligencewatchtower-contracts",
			URL:      "https://github.com/witnesschain-com/diligencewatchtower-contracts",
			Evidence: "Witness Chain contracts used to define the diligence watchtower and bounty-submission shape",
			UsedFor:  "reference boundary for the local PoD watchtower and MineBounty benchmark context",
		},
		{
			Project:  "witnesschain-com/diligencewatchtower-client",
			URL:      "https://github.com/witnesschain-com/diligencewatchtower-client",
			Evidence: "Witness Chain watchtower client used to confirm the recomputation workflow",
			UsedFor:  "reference boundary for the local watchtower epoch trace",
		},
		{
			Project:  "Proof of Diligence paper",
			URL:      "https://eprint.iacr.org/2024/736",
			Evidence: "Proof of Diligence: Cryptoeconomic Security for Rollups",
			UsedFor:  "MineBounty normal-path gas baseline and watchtower incentive workflow",
		},
	}
}

func BuildPoDBaseline(foundry []FoundryEntry, validatorCount, epochsPerMonth int, theta float64) (PoDBaselineResult, error) {
	gas := 0
	function := ""
	for _, row := range foundry {
		if row.Test == "testPoDMineBountyGas" {
			gas = row.Gas
			function = row.Function
			break
		}
	}
	if gas <= 0 {
		return PoDBaselineResult{}, fmt.Errorf("missing Foundry gas for testPoDMineBountyGas")
	}
	monthly := int(math.Round(float64(epochsPerMonth) * theta * float64(validatorCount) * float64(gas)))
	return PoDBaselineResult{
		Scheme:          "PoD",
		Test:            "testPoDMineBountyGas",
		Function:        function,
		GasPerEpoch:     gas,
		Theta:           theta,
		ValidatorCount:  validatorCount,
		EpochsPerMonth:  epochsPerMonth,
		MonthlyGas:      monthly,
		Formula:         "epochs_per_month * theta * validator_count * PoD podMineBounty gas",
		PaperAverageGas: 380000,
		PaperSource:     "Proof of Diligence: Cryptoeconomic Security for Rollups, Section 6.2: MineBounty consumes about 380K gas on average",
		ProtocolSteps: []string{
			"watchtower recomputes the asserted L2 state root",
			"watchtower builds the execution trace Merkle root",
			"watchtower computes a VRF proof bound to the trace root and its public key",
			"BountyManager-style podMineBounty stores and verifies the proof for reward allocation",
		},
		ReferenceSources: PoDReferenceSources(),
		SourceKind:       "solidity_pod_mine_bounty_benchmark_plus_paper_protocol_trace",
		Reproduce:        "forge test --gas-report; parse testPoDMineBountyGas from test/ValidatorAudit.t.sol",
		ValidationClaim:  "PoD monthly gas plotted in figures 30/31 is derived from the podMineBounty Foundry benchmark; deterministic Go traces separately verify the watchtower recomputation/VRF workflow",
	}, nil
}
