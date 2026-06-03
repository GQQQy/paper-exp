package audit

import "testing"

func shortParams() Params {
	p := DefaultParams()
	p.TWin = 600
	p.M = 20
	p.S = 8
	p.L = 300
	p.Ms = 5
	p.Mc = 20
	return p
}

func TestRanCkHonestTrackInitHBRespondAndContAudit(t *testing.T) {
	trace := SimulateRanCk(shortParams(), BehaviorHonest)
	if trace.TrackInit.TC0 == "" {
		t.Fatal("TrackInit did not commit TC_0")
	}
	if trace.HeartbeatCount == 0 {
		t.Fatal("expected triggered heartbeats")
	}
	if trace.MissedHeartbeats != 0 {
		t.Fatalf("honest validator missed heartbeats: %d", trace.MissedHeartbeats)
	}
	if !trace.ContAudit.Passed {
		t.Fatalf("honest ContAudit failed: %s", trace.ContAudit.Reason)
	}
}

func TestRanCkMissedHeartbeatSlashesOffline(t *testing.T) {
	trace := SimulateRanCk(shortParams(), BehaviorOffline)
	if trace.MissedHeartbeats == 0 {
		t.Fatal("offline validator should miss heartbeats")
	}
	if trace.ContAudit.Passed || !trace.ContAudit.Detected || !trace.ContAudit.MissedHeartbeat {
		t.Fatalf("offline validator not detected: %+v", trace.ContAudit)
	}
}

func TestRanCkRejectsWrongSeedNonceOrAlphaWitness(t *testing.T) {
	trace := SimulateRanCk(shortParams(), BehaviorInconsistent)
	if trace.ContAudit.Passed {
		t.Fatal("inconsistent credential chain passed ContAudit")
	}
	if trace.ContAudit.EndpointOK && trace.ContAudit.SampleOK {
		t.Fatalf("expected endpoint or sample mismatch, got %+v", trace.ContAudit)
	}
}

func TestRanCkIntermittentAndRecomputeFailDetected(t *testing.T) {
	for _, behavior := range []string{BehaviorIntermittent, BehaviorRecomputeFail} {
		trace := SimulateRanCk(shortParams(), behavior)
		if !trace.ContAudit.Detected {
			t.Fatalf("%s should be detected by heartbeat or continuity audit", behavior)
		}
	}
}

func TestSenCkHonestReplayGeneratesDigest(t *testing.T) {
	trace := SimulateSenCk(shortParams(), BehaviorHonest)
	if trace.AudRoot == "" {
		t.Fatal("missing AudRoot")
	}
	if !trace.SentReport.Passed {
		t.Fatalf("honest replay failed: %s", trace.SentReport.Reason)
	}
	for _, report := range trace.SentReport.SampledReports {
		if !report.ReplayMatched {
			t.Fatalf("digest mismatch for honest report: %+v", report)
		}
	}
}

func TestSenCkTraceComesFromEVMOpcodeHook(t *testing.T) {
	trace := SimulateSenCk(shortParams(), BehaviorHonest)
	if len(trace.TraceSample) == 0 {
		t.Fatal("missing EVM hook trace sample")
	}
	seenMemoryOrStorage := false
	seenStack := false
	for _, step := range trace.TraceSample {
		if step.HookSource != "EVM opcode post-execution audit hook" {
			t.Fatalf("unexpected hook source: %q", step.HookSource)
		}
		if step.OpName == "" || step.RW == "" || step.Value == "" {
			t.Fatalf("opcode hook did not emit pc/op/rw_t/val_t: %+v", step)
		}
		for _, access := range step.Accesses {
			switch access.Kind {
			case "memory_read", "memory_write", "storage_read", "storage_write":
				seenMemoryOrStorage = true
			case "stack_push", "stack_pop":
				seenStack = true
			}
		}
	}
	if !seenMemoryOrStorage {
		t.Fatal("EVM hook trace did not include memory/storage runtime accesses")
	}
	if !seenStack {
		t.Fatal("EVM hook trace did not include stack runtime accesses")
	}
	for _, report := range trace.SegmentReports {
		if report.ExecutionLayer != "instrumented local EVM opcode interpreter" {
			t.Fatalf("segment report is not tied to the instrumented EVM layer: %+v", report)
		}
		if report.BytecodeHash == "" || report.SnapshotCommitment == "" {
			t.Fatalf("segment report missing bytecode/snapshot provenance: %+v", report)
		}
	}
}

func TestSenCkLazyGuessCannotStablyPass(t *testing.T) {
	trace := SimulateSenCk(shortParams(), BehaviorLazyGuess)
	if trace.SentReport.Passed || !trace.SentReport.Detected {
		t.Fatalf("lazy report should be detected: %+v", trace.SentReport)
	}
}

func TestSenCkDisputeDetectsWrongReportAndPollutedComAud(t *testing.T) {
	for _, behavior := range []string{BehaviorLazyGuess, BehaviorExecutorPollute} {
		trace := SimulateSenCk(shortParams(), behavior)
		if !trace.SentReport.DisputePassed {
			t.Fatalf("%s dispute/prove path should detect the fault", behavior)
		}
	}
}

func TestIntegratedAuditFlow(t *testing.T) {
	p := shortParams()
	ranck := SimulateRanCk(p, BehaviorHonest)
	senck := SimulateSenCk(p, BehaviorHonest)
	if !ranck.ContAudit.Passed {
		t.Fatalf("integrated RanCk stage failed: %s", ranck.ContAudit.Reason)
	}
	if !senck.SentReport.Passed {
		t.Fatalf("integrated SenCk stage failed: %s", senck.SentReport.Reason)
	}
	if len(ranck.Heartbeats) == 0 || len(senck.SentReport.SampledReports) == 0 {
		t.Fatal("integrated flow did not produce heartbeat and sentinel evidence")
	}
}
