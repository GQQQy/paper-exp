package comparisons

import (
	"math"
	"testing"
)

func TestBuildResultsCoversPaperComparisonProtocols(t *testing.T) {
	gas := map[string]uint64{
		"arbitrumOptimisticPath": 277926,
		"truebitOptimisticPath":  303640,
		"cartesiOptimisticPath":  327854,
		"boldOptimisticPath":     297357,
		"cleverOptimisticPath":   389652,
		"arbitrumClassicPath":    4438574,
		"truebitPath":            3828334,
		"cartesiDavePath":        5956440,
		"boldPath":               3560120,
		"cleverPath":             551974,
	}
	results, err := BuildResults(gas, 1e11)
	if err != nil {
		t.Fatalf("BuildResults returned error: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 protocol results, got %d", len(results))
	}
	want := map[string]struct {
		dispute float64
		fn      string
	}{
		"Arbitrum\nClassic": {4438.574, "arbitrumClassicPath"},
		"TrueBit":           {3828.334, "truebitPath"},
		"Cartesi\nDave":     {5956.440, "cartesiDavePath"},
		"Arbitrum\nBoLD":    {3560.120, "boldPath"},
		"CleVer\n(ours)":    {551.974, "cleverPath"},
	}
	for _, result := range results {
		expect, ok := want[result.Scheme]
		if !ok {
			t.Fatalf("unexpected scheme %q", result.Scheme)
		}
		if math.Abs(result.DisputeGasK-expect.dispute) > 1e-9 {
			t.Fatalf("%s dispute gas = %.3fK, want %.3fK", result.Scheme, result.DisputeGasK, expect.dispute)
		}
		if result.Dispute.Function != expect.fn {
			t.Fatalf("%s dispute function = %s, want %s", result.Scheme, result.Dispute.Function, expect.fn)
		}
		if len(result.DisputeFlow) < 4 {
			t.Fatalf("%s missing protocol dispute flow: %+v", result.Scheme, result.DisputeFlow)
		}
		if result.Reproduce == "" || result.ValidationClaim == "" {
			t.Fatalf("%s missing reproducibility metadata", result.Scheme)
		}
		if result.TimelineSlots != result.Timeline.Slots {
			t.Fatalf("%s timeline slots = %d, derivation slots = %d", result.Scheme, result.TimelineSlots, result.Timeline.Slots)
		}
		if len(result.Timeline.Events) != result.TimelineSlots {
			t.Fatalf("%s timeline event count = %d, want %d", result.Scheme, len(result.Timeline.Events), result.TimelineSlots)
		}
		if result.Timeline.Formula == "" || result.Timeline.Strategy == "" {
			t.Fatalf("%s missing timeline derivation metadata", result.Scheme)
		}
	}
}

func TestBuildTimelineMatchesFigure12OrderingAndValues(t *testing.T) {
	params := TimelineParams{TExec: 1, TSlot: 0.008, Gamma: 0.85, TSeg: 0.02, Eta: 0.4}
	timeline := BuildTimeline(params, 1e11)
	wantNames := []string{"CleVer\n（并发）", "Arbitrum\nBoLD", "Cartesi\nDave", "TrueBit", "Arbitrum\nClassic"}
	wantTotals := []float64{1.044, 2.65, 2.73, 2.714, 2.69}
	if len(timeline) != len(wantNames) {
		t.Fatalf("timeline length = %d, want %d", len(timeline), len(wantNames))
	}
	for i := range wantNames {
		if timeline[i].Name != wantNames[i] {
			t.Fatalf("timeline[%d] name = %q, want %q", i, timeline[i].Name, wantNames[i])
		}
		if math.Abs(timeline[i].TotalTime-wantTotals[i]) > 1e-9 {
			t.Fatalf("timeline[%d] total = %.3f, want %.3f", i, timeline[i].TotalTime, wantTotals[i])
		}
		if timeline[i].Formula == "" {
			t.Fatalf("timeline[%d] missing formula", i)
		}
		if timeline[i].Derivation.Slots != timeline[i].DisputeSlots {
			t.Fatalf("timeline[%d] derivation slots = %d, want %d", i, timeline[i].Derivation.Slots, timeline[i].DisputeSlots)
		}
		if len(timeline[i].Derivation.Events) != timeline[i].DisputeSlots {
			t.Fatalf("timeline[%d] event count = %d, want %d", i, len(timeline[i].Derivation.Events), timeline[i].DisputeSlots)
		}
	}
}

func TestBuildResultsRejectsMissingFoundryGas(t *testing.T) {
	_, err := BuildResults(map[string]uint64{"cleverPath": 551676}, 1e11)
	if err == nil {
		t.Fatal("expected missing Foundry gas error")
	}
}
