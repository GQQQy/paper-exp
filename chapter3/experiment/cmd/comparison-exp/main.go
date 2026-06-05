package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"clever-chapter3-exp/comparisons"
)

type rawLog struct {
	FoundryGas []struct {
		Function string `json:"function"`
		Gas      uint64 `json:"gas"`
	} `json:"foundry_gas_runs"`
}

type report struct {
	Source            string                       `json:"source"`
	RequiredFunctions []string                     `json:"required_functions"`
	Protocols         []comparisons.ProtocolResult `json:"protocols"`
	Timeline          []comparisons.TimelineResult `json:"timeline"`
}

func main() {
	rawPath := flag.String("raw", filepath.Join("logs", "raw_experiment_log.json"), "chapter 3 raw experiment log")
	outPath := flag.String("out", filepath.Join("logs", "comparison_protocols.json"), "standalone comparison report")
	totalGas := flag.Float64("total-gas", 1e11, "paper-scale task gas used to derive logical dispute rounds")
	flag.Parse()

	rawBytes, err := os.ReadFile(*rawPath)
	if err != nil {
		panic(fmt.Sprintf("read raw log: %v; run go run ./cmd/clever-exp first", err))
	}
	var raw rawLog
	if err := json.Unmarshal(rawBytes, &raw); err != nil {
		panic(err)
	}
	gasByFunction := map[string]uint64{}
	for _, row := range raw.FoundryGas {
		gasByFunction[row.Function] = row.Gas
	}
	protocols, err := comparisons.BuildResults(gasByFunction, *totalGas)
	if err != nil {
		panic(err)
	}
	timeline := comparisons.BuildTimeline(comparisons.TimelineParams{TExec: 1, TSlot: 0.008, Gamma: 0.85, TSeg: 0.02, Eta: 0.4}, *totalGas)
	out := report{
		Source:            "local Foundry gas report parsed from chapter3/experiment/logs/raw_experiment_log.json",
		RequiredFunctions: comparisons.RequiredFunctions(*totalGas),
		Protocols:         protocols,
		Timeline:          timeline,
	}
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		panic(err)
	}
	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(*outPath, append(encoded, '\n'), 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("[comparison] protocols=%d out=%s\n", len(protocols), *outPath)
	for _, item := range protocols {
		fmt.Printf("[comparison] %-18s optimistic=%.3fK dispute=%.3fK via %s/%s\n",
			compactScheme(item.Scheme), item.OptimisticGasK, item.DisputeGasK, item.Optimistic.Function, item.Dispute.Function)
	}
}

func compactScheme(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\n' {
			out = append(out, ' ')
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
