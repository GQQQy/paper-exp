package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"chapter5/experiment/comparisons"
)

type rawLog struct {
	Gas struct {
		FoundryParsed []struct {
			Test     string `json:"test"`
			Function string `json:"function"`
			Gas      int    `json:"gas"`
		} `json:"foundry_parsed"`
	} `json:"gas_trace"`
}

type report struct {
	Source                string                         `json:"source"`
	ComparisonExperiments []comparisons.SchemeExperiment `json:"comparison_experiments"`
	ComparisonTable11     []map[string]string            `json:"comparison_table_11"`
	PoDBaseline           comparisons.PoDBaselineResult  `json:"pod_baseline"`
}

func main() {
	rawPath := flag.String("raw", filepath.Join("logs", "raw_experiment_log.json"), "chapter 5 raw experiment log")
	outPath := flag.String("out", filepath.Join("logs", "comparison_experiments.json"), "standalone comparison report")
	validatorCount := flag.Int("validators", 20, "validator count for PoD monthly baseline")
	epochsPerMonth := flag.Int("epochs-per-month", 6*24*30, "PoD epochs per month")
	theta := flag.Float64("theta", 0.9, "PoD active attestation ratio")
	flag.Parse()

	rawBytes, err := os.ReadFile(*rawPath)
	if err != nil {
		panic(fmt.Sprintf("read raw log: %v; run go run ./cmd/audit-exp first", err))
	}
	var raw rawLog
	if err := json.Unmarshal(rawBytes, &raw); err != nil {
		panic(err)
	}
	foundry := make([]comparisons.FoundryEntry, 0, len(raw.Gas.FoundryParsed))
	for _, row := range raw.Gas.FoundryParsed {
		foundry = append(foundry, comparisons.FoundryEntry{Test: row.Test, Function: row.Function, Gas: row.Gas})
	}
	pod, err := comparisons.BuildPoDBaseline(foundry, *validatorCount, *epochsPerMonth, *theta)
	if err != nil {
		panic(err)
	}
	experiments := comparisons.RunCapabilityExperiments()
	out := report{
		Source:                "scenario comparison experiment plus Foundry PoD gas parsed from chapter5/experiment/logs/raw_experiment_log.json",
		ComparisonExperiments: experiments,
		ComparisonTable11:     comparisons.Table11Rows(experiments),
		PoDBaseline:           pod,
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
	fmt.Printf("[comparison] table11_rows=%d pod_epoch_gas=%d pod_monthly_gas=%d out=%s\n",
		len(out.ComparisonTable11), pod.GasPerEpoch, pod.MonthlyGas, *outPath)
	for _, item := range experiments {
		fmt.Printf("[comparison] %-12s online=%s semantic=%s unpredictable=%s external=%s\n",
			item.Scheme, item.OnlineAudit, item.SemanticDiligence, item.UnpredictableAudit, item.ExternalNetwork)
	}
}
