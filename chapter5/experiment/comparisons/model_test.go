package comparisons

import "testing"

func TestCapabilityExperimentsGenerateTable11Rows(t *testing.T) {
	experiments := RunCapabilityExperiments()
	rows := Table11Rows(experiments)
	if len(experiments) != 4 || len(rows) != 4 {
		t.Fatalf("expected 4 comparison experiments and rows, got %d/%d", len(experiments), len(rows))
	}
	byScheme := map[string]map[string]string{}
	for _, row := range rows {
		byScheme[row["scheme"]] = row
	}
	want := map[string]map[string]string{
		"TrueBit": {
			"online_audit":        "no",
			"semantic_diligence":  "partial",
			"unpredictable_audit": "-",
			"external_network":    "no",
		},
		"Arbitrum": {
			"online_audit":        "no",
			"semantic_diligence":  "no",
			"unpredictable_audit": "-",
			"external_network":    "no",
		},
		"PoD": {
			"online_audit":        "yes",
			"semantic_diligence":  "no",
			"unpredictable_audit": "limited",
			"external_network":    "yes",
		},
		"RanCk+SenCk": {
			"online_audit":        "yes",
			"semantic_diligence":  "yes",
			"unpredictable_audit": "yes",
			"external_network":    "no",
		},
	}
	for scheme, fields := range want {
		row, ok := byScheme[scheme]
		if !ok {
			t.Fatalf("missing scheme %s", scheme)
		}
		for key, value := range fields {
			if row[key] != value {
				t.Fatalf("%s %s = %s, want %s", scheme, key, row[key], value)
			}
		}
	}
	for _, exp := range experiments {
		if len(exp.Scenarios) != 3 {
			t.Fatalf("%s scenario count = %d, want 3", exp.Scheme, len(exp.Scenarios))
		}
		if len(exp.ProtocolFlow) < 4 {
			t.Fatalf("%s protocol flow is incomplete: %+v", exp.Scheme, exp.ProtocolFlow)
		}
		if exp.BenchmarkMode == "" || exp.SourceKind == "" {
			t.Fatalf("%s missing benchmark/source provenance: %+v", exp.Scheme, exp)
		}
		if exp.Reproduce == "" || exp.ValidationClaim == "" {
			t.Fatalf("%s missing reproducibility metadata", exp.Scheme)
		}
	}
}

func TestBuildPoDBaselineUsesFoundryGas(t *testing.T) {
	pod, err := BuildPoDBaseline([]FoundryEntry{
		{Test: "testPoDMineBountyGas", Function: "podMineBounty", Gas: 380000},
	}, 20, 4320, 0.9)
	if err != nil {
		t.Fatalf("BuildPoDBaseline returned error: %v", err)
	}
	if pod.MonthlyGas != 29548800000 {
		t.Fatalf("monthly gas = %d, want 29548800000", pod.MonthlyGas)
	}
	if pod.Function != "podMineBounty" || pod.GasPerEpoch != 380000 {
		t.Fatalf("unexpected PoD benchmark: %+v", pod)
	}
	if pod.PaperAverageGas != 380000 || len(pod.ProtocolSteps) == 0 {
		t.Fatalf("missing PoD paper/protocol provenance: %+v", pod)
	}
	if pod.Reproduce == "" || pod.ValidationClaim == "" {
		t.Fatal("missing PoD reproducibility metadata")
	}
}

func TestBuildPoDBaselineRejectsMissingFoundryGas(t *testing.T) {
	if _, err := BuildPoDBaseline(nil, 20, 4320, 0.9); err == nil {
		t.Fatal("expected missing Foundry gas error")
	}
}
