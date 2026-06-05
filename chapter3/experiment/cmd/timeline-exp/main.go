package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"clever-chapter3-exp/comparisons"
)

func main() {
	outPath := flag.String("out", filepath.Join("logs", "timeline_simulation.json"), "timeline simulation JSON output")
	totalGas := flag.Float64("total-gas", 1e11, "paper-scale task gas used to derive logical dispute rounds")
	tExec := flag.Float64("t-exec", 1.0, "normalized execution time")
	tSlot := flag.Float64("t-slot", 0.008, "normalized dispute slot time")
	gamma := flag.Float64("gamma", 0.85, "post-execution verification delay")
	tSeg := flag.Float64("t-seg", 0.02, "CleVer segment verification time")
	eta := flag.Float64("eta", 0.4, "error injection progress fraction")
	flag.Parse()

	params := comparisons.TimelineParams{
		TExec: *tExec,
		TSlot: *tSlot,
		Gamma: *gamma,
		TSeg:  *tSeg,
		Eta:   *eta,
	}
	report := comparisons.RunTimelineExperiment(*totalGas, params)
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		panic(err)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(*outPath, append(encoded, '\n'), 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("[timeline] simulations=%d out=%s\n", len(report.Protocols), *outPath)
	for _, protocol := range report.Protocols {
		fmt.Printf("[timeline] %-18s slots=%d events=%d formula=%s\n",
			compactScheme(protocol.Scheme),
			protocol.TimelineDerivation.Slots,
			len(protocol.TimelineDerivation.Events),
			protocol.TimelineDerivation.Formula,
		)
	}
}

func compactScheme(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}
