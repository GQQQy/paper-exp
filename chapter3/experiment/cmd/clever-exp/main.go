package main

import (
	"crypto/sha256"
	"encoding/binary"
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
	"time"
)

type Params struct {
	SegmentBudget         float64 `json:"segment_budget_B"`
	AdjudicationThreshold float64 `json:"adjudication_threshold_b"`
	Alpha                 float64 `json:"alpha"`
	StakeRounds           int     `json:"staking_rounds_g"`
	StakeBase             float64 `json:"staking_d0"`
}

type RawTaskSample struct {
	Task             string  `json:"task"`
	RequestedSize    int     `json:"requested_size"`
	Steps            int     `json:"steps"`
	SnapshotBytes    int     `json:"snapshot_bytes"`
	Commitment       string  `json:"commitment"`
	ElapsedSeconds   float64 `json:"elapsed_seconds"`
	ReachedTimeLimit bool    `json:"reached_time_limit"`
}

type RawExperimentLog struct {
	Metadata      map[string]any  `json:"metadata"`
	Params        Params          `json:"params"`
	Samples       []RawTaskSample `json:"samples"`
	EVMSamples    []EVMRunSample  `json:"geth_evm_samples"`
	FoundryGas    []FoundryGasRun  `json:"foundry_gas_runs"`
	Comparisons   []ComparisonRun `json:"comparison_protocols"`
	Foundry       map[string]any  `json:"foundry"`
	LongRunPolicy map[string]any  `json:"long_run_policy"`
}

type FoundryGasRun struct {
	Contract string `json:"contract"`
	Function string `json:"function"`
	Gas      uint64 `json:"gas"`
}

type EVMRunSample struct {
	Task           string `json:"task"`
	BytecodeBytes  int    `json:"bytecode_bytes"`
	GasUsed        uint64 `json:"gas_used"`
	ExecutionTime  string `json:"execution_time"`
	Allocations    uint64 `json:"allocations"`
	AllocatedBytes uint64 `json:"allocated_bytes"`
	RawBenchOutput string `json:"raw_bench_output"`
}

type ComparisonRun struct {
	Scheme          string  `json:"scheme"`
	Mechanism       string  `json:"mechanism"`
	OnchainRounds   int     `json:"onchain_rounds"`
	OptimisticGasK  float64 `json:"optimistic_gas_k"`
	DisputeGasK     float64 `json:"dispute_gas_k"`
	Notes           string  `json:"notes"`
}

type VMState struct {
	PC      uint64
	Acc     uint64
	Memory  []uint64
	Program []uint64
}

var params = Params{
	SegmentBudget:         1e8,
	AdjudicationThreshold: 1e6,
	Alpha:                 0.8,
	StakeRounds:           10,
	StakeBase:             0.5,
}

func main() {
	out := flag.String("out", filepath.Join("logs", "raw_experiment_log.json"), "output raw experiment JSON path")
	quick := flag.Bool("quick", true, "run short physical samples")
	full := flag.Bool("full", false, "run larger local samples but still stop at max-seconds")
	maxSeconds := flag.Float64("max-seconds", 5, "maximum seconds per task sample")
	flag.Parse()

	if *full {
		*quick = false
	}
	start := time.Now()
	foundryGas := runFoundryGasReport()
	log := RawExperimentLog{
		Metadata: map[string]any{
			"chapter":     "第三章 基于有状态任务切片的链下计算验证",
			"source":      "Go physical task runner: execution, snapshot serialization, commitment",
			"created_at":  time.Now().Format(time.RFC3339),
			"quick_mode":  *quick,
			"elapsed_sec": 0,
		},
		Params:  params,
		Samples: runPhysicalSamples(*quick, time.Duration(*maxSeconds*float64(time.Second))),
		EVMSamples: runGethEVMSamples(),
		FoundryGas: foundryGas,
		Comparisons: runComparisonProtocols(foundryGas),
		Foundry: map[string]any{
			"command": "forge test --gas-report",
			"purpose": "collect Solidity benchmark task gas and VerSeg gas in a local EVM; geth evm --bench is also called by this Go runner",
		},
		LongRunPolicy: map[string]any{
			"reason": "paper workloads include minute/hour-scale tasks; this runner records bounded samples and logs if a sample reaches max-seconds",
			"max_seconds_per_task": *maxSeconds,
		},
	}
	log.Metadata["elapsed_sec"] = time.Since(start).Seconds()

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
	for _, sample := range log.Samples {
		fmt.Printf("[sample] %-10s steps=%d snapshot=%dB commitment=%s elapsed=%.3fs\n", sample.Task, sample.Steps, sample.SnapshotBytes, sample.Commitment[:8], sample.ElapsedSeconds)
		if sample.ReachedTimeLimit {
			fmt.Printf("[sample] %-10s reached max-seconds; long paper-scale run intentionally skipped\n", sample.Task)
		}
	}
	for _, sample := range log.EVMSamples {
		fmt.Printf("[geth-evm] %-10s gas=%d time=%s alloc=%d bytes=%d\n", sample.Task, sample.GasUsed, sample.ExecutionTime, sample.Allocations, sample.AllocatedBytes)
	}
	for _, gas := range log.FoundryGas {
		if strings.Contains(gas.Function, "DisputePath") || gas.Function == "verSeg" {
			fmt.Printf("[foundry] %-34s gas=%d\n", gas.Function, gas.Gas)
		}
	}
	for _, comparison := range log.Comparisons {
		fmt.Printf("[compare] %-16s rounds=%d optimistic=%.0fK dispute=%.0fK\n", strings.ReplaceAll(comparison.Scheme, "\n", " "), comparison.OnchainRounds, comparison.OptimisticGasK, comparison.DisputeGasK)
	}
	fmt.Printf("原始实验日志已生成：%s\n", *out)
}

func runPhysicalSamples(quick bool, maxDuration time.Duration) []RawTaskSample {
	sizes := map[string]int{"Fibonacci": 5000, "Poly-Chain": 4000, "Sort-Large": 384, "DP-Large": 12000}
	if !quick {
		sizes = map[string]int{"Fibonacci": 50000, "Poly-Chain": 40000, "Sort-Large": 1024, "DP-Large": 120000}
	}
	order := []string{"Fibonacci", "Poly-Chain", "Sort-Large", "DP-Large"}
	samples := make([]RawTaskSample, 0, len(order))
	for _, task := range order {
		start := time.Now()
		state, steps := executeTaskSample(task, sizes[task], maxDuration)
		encoded := encodeState(state)
		commitment := sha256.Sum256(encoded)
		elapsed := time.Since(start)
		samples = append(samples, RawTaskSample{
			Task:             task,
			RequestedSize:    sizes[task],
			Steps:            steps,
			SnapshotBytes:    len(encoded),
			Commitment:       fmt.Sprintf("%x", commitment[:]),
			ElapsedSeconds:   elapsed.Seconds(),
			ReachedTimeLimit: elapsed >= maxDuration,
		})
	}
	return samples
}

func executeTaskSample(name string, n int, maxDuration time.Duration) (VMState, int) {
	state := VMState{Memory: make([]uint64, 64), Program: make([]uint64, 32)}
	deadline := time.Now().Add(maxDuration)
	steps := 0
	switch name {
	case "Fibonacci":
		a, b := uint64(0), uint64(1)
		for i := 0; i < n && time.Now().Before(deadline); i++ {
			a, b = b, a+b
			state.Acc ^= a + uint64(i)
			state.Memory[i%len(state.Memory)] = a
			steps++
		}
	case "Poly-Chain":
		x := uint64(7)
		for i := 0; i < n && time.Now().Before(deadline); i++ {
			x = x*x + uint64(31*i+7)
			state.Memory[i%len(state.Memory)] ^= x
			steps++
		}
		state.Acc = x
	case "Sort-Large":
		arr := make([]uint64, n)
		for i := range arr {
			arr[i] = uint64((i*1103515245 + 12345) % 1_000_000)
		}
		for width := 1; width < n && time.Now().Before(deadline); width *= 2 {
			for left := 0; left < n && time.Now().Before(deadline); left += 2 * width {
				mid := min(left+width, n)
				right := min(left+2*width, n)
				for i := left; i < mid; i++ {
					for j := mid; j < right; j++ {
						if arr[j] < arr[i] {
							arr[i], arr[j] = arr[j], arr[i]
						}
						steps++
					}
				}
			}
		}
		for i := range state.Memory {
			state.Memory[i] = arr[i%len(arr)]
		}
	case "DP-Large":
		dp0, dp1 := uint64(1), uint64(1)
		for i := 2; i <= n && time.Now().Before(deadline); i++ {
			next := dp0 + dp1 + uint64(i)
			dp0, dp1 = dp1, next
			state.Memory[i%len(state.Memory)] = next
			steps++
		}
		state.Acc = dp1
	}
	state.PC = uint64(steps)
	return state, steps
}

func encodeState(state VMState) []byte {
	buf := make([]byte, 16+8*len(state.Memory)+8*len(state.Program))
	binary.LittleEndian.PutUint64(buf[0:8], state.PC)
	binary.LittleEndian.PutUint64(buf[8:16], state.Acc)
	offset := 16
	for _, value := range state.Memory {
		binary.LittleEndian.PutUint64(buf[offset:offset+8], value)
		offset += 8
	}
	for _, value := range state.Program {
		binary.LittleEndian.PutUint64(buf[offset:offset+8], value)
		offset += 8
	}
	return buf
}

func runGethEVMSamples() []EVMRunSample {
	programs := map[string]string{
		"Fibonacci":  strings.Repeat("6001600055", 32),
		"Poly-Chain": strings.Repeat("6002600302600055", 24),
		"Sort-Large": strings.Repeat("60016000516002600155600054", 16),
		"DP-Large":   strings.Repeat("60016000516001600101600055", 48),
	}
	order := []string{"Fibonacci", "Poly-Chain", "Sort-Large", "DP-Large"}
	samples := make([]EVMRunSample, 0, len(order))
	for _, task := range order {
		samples = append(samples, runSingleEVM(task, programs[task]))
	}
	return samples
}

func runSingleEVM(task, code string) EVMRunSample {
	cmd := exec.Command("evm", "--code", code, "--gas", "100000000", "--bench", "run")
	out, err := cmd.CombinedOutput()
	raw := string(out)
	if err != nil {
		return EVMRunSample{Task: task, BytecodeBytes: len(code) / 2, RawBenchOutput: raw + err.Error()}
	}
	return EVMRunSample{
		Task:           task,
		BytecodeBytes:  len(code) / 2,
		GasUsed:        parseUint(raw, `EVM gas used:\s+([0-9]+)`),
		ExecutionTime:  parseString(raw, `execution time:\s+([^\n]+)`),
		Allocations:    parseUint(raw, `allocations:\s+([0-9]+)`),
		AllocatedBytes: parseUint(raw, `allocated bytes:\s+([0-9]+)`),
		RawBenchOutput: raw,
	}
}

func runComparisonProtocols(foundryGas []FoundryGasRun) []ComparisonRun {
	totalGas := 1e11
	rounds := int(math.Ceil(math.Log2(totalGas)))
	gasByFunction := map[string]uint64{}
	for _, run := range foundryGas {
		gasByFunction[run.Function] = run.Gas
	}
	required := []string{
		"arbitrumOptimisticPath", "truebitOptimisticPath", "cartesiOptimisticPath", "boldOptimisticPath", "cleverOptimisticPath",
		"arbitrumClassicPath", "truebitPath", "cartesiDavePath", "boldPath", "cleverPath",
	}
	for _, name := range required {
		if gasByFunction[name] == 0 {
			panic("missing Foundry gas for " + name)
		}
	}
	if len(gasByFunction) == 0 {
		panic("missing Foundry dispute path gas; run forge test --gas-report successfully")
	}
	return []ComparisonRun{
		comparison("Arbitrum\nClassic", "1v1 bisection", rounds, gasK(gasByFunction["arbitrumOptimisticPath"]), gasK(gasByFunction["arbitrumClassicPath"]), "direct Foundry gas for representative O(log N) bisection path"),
		comparison("TrueBit", "solver-verifier bisection", rounds, gasK(gasByFunction["truebitOptimisticPath"]), gasK(gasByFunction["truebitPath"]), "direct Foundry gas for solver/verifier bisection path"),
		comparison("Cartesi\nDave", "tournament + bisection", rounds, gasK(gasByFunction["cartesiOptimisticPath"]), gasK(gasByFunction["cartesiDavePath"]), "direct Foundry gas for tournament plus bisection path"),
		comparison("Arbitrum\nBoLD", "three-level bisection", rounds, gasK(gasByFunction["boldOptimisticPath"]), gasK(gasByFunction["boldPath"]), "direct Foundry gas for three-level dispute path"),
		comparison("CleVer\n(ours)", "two-layer slicing + VerSeg", 3, gasK(gasByFunction["cleverOptimisticPath"]), gasK(gasByFunction["cleverPath"]), "direct Foundry gas for two-layer slicing and bounded VerSeg path"),
	}
}

func gasK(gas uint64) float64 {
	return float64(gas) / 1000
}

func runFoundryGasReport() []FoundryGasRun {
	cmd := exec.Command("forge", "test", "--gas-report")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("forge test --gas-report failed: %s\n%s", err, string(out)))
	}
	return parseFoundryGasReport(string(out))
}

func parseFoundryGasReport(raw string) []FoundryGasRun {
	known := map[string]string{
		"dpLarge": "BenchmarkTasks",
		"fibonacci": "BenchmarkTasks",
		"polyChain": "BenchmarkTasks",
		"sortLarge": "BenchmarkTasks",
		"verSeg": "CleVerVerifier",
		"arbitrumClassicPath": "DisputeProtocolBenchmarks",
		"truebitPath": "DisputeProtocolBenchmarks",
		"cartesiDavePath": "DisputeProtocolBenchmarks",
		"boldPath": "DisputeProtocolBenchmarks",
		"cleverPath": "DisputeProtocolBenchmarks",
		"arbitrumOptimisticPath": "DisputeProtocolBenchmarks",
		"truebitOptimisticPath": "DisputeProtocolBenchmarks",
		"cartesiOptimisticPath": "DisputeProtocolBenchmarks",
		"boldOptimisticPath": "DisputeProtocolBenchmarks",
		"cleverOptimisticPath": "DisputeProtocolBenchmarks",
	}
	lineRe := regexp.MustCompile(`\|\s*([A-Za-z0-9_]+)\s+\|\s*([0-9]+)\s+\|`)
	runs := []FoundryGasRun{}
	for _, line := range strings.Split(raw, "\n") {
		match := lineRe.FindStringSubmatch(line)
		if len(match) < 3 {
			continue
		}
		contract, ok := known[match[1]]
		if !ok {
			continue
		}
		gas, _ := strconv.ParseUint(match[2], 10, 64)
		runs = append(runs, FoundryGasRun{Contract: contract, Function: match[1], Gas: gas})
	}
	return runs
}

func comparison(name, mechanism string, rounds int, optimistic, dispute float64, notes string) ComparisonRun {
	return ComparisonRun{Scheme: name, Mechanism: mechanism, OnchainRounds: rounds, OptimisticGasK: optimistic, DisputeGasK: dispute, Notes: notes}
}

func verSegReplayGasK(threshold, factor float64) float64 {
	return (8000 + 1.15*threshold) * factor / 1000
}

func parseUint(raw, pattern string) uint64 {
	match := regexp.MustCompile(pattern).FindStringSubmatch(raw)
	if len(match) < 2 {
		return 0
	}
	value, _ := strconv.ParseUint(match[1], 10, 64)
	return value
}

func parseString(raw, pattern string) string {
	match := regexp.MustCompile(pattern).FindStringSubmatch(raw)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
