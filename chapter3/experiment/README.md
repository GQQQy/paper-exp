# 第三章 CleVer 实验工程

本目录用于运行第三章 CleVer 实验，生成 `logs/raw_experiment_log.json`，并为 `../visualization/` 提供图 7-12 所需的数据输入。

## 实验内容

- 任务 runner：运行 Fibonacci、Poly-Chain、Sort-Large、DP-Large 四类任务样本，记录执行步数、快照大小和状态承诺。
- EVM 执行采样：调用 Geth `evm --bench run`，记录不同任务字节码的 Gas、执行时间和内存分配情况。
- 链上 Gas 测量：调用 Foundry `forge test --gas-report`，统计 `VerSeg` 和对比协议路径的 Gas 开销。
- 对比协议实验：`comparisons/model.go` 将 Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD、CleVer 的对比路径绑定到具体 Solidity benchmark 函数、参数和时间线模型。
- 时间线模拟：`cmd/timeline-exp` 运行本地 dispute timeline 状态机，输出 `logs/timeline_simulation.json`；图 12 的 `3/25/35/33/30` 个槽数均来自该模拟日志中的事件数量。
- 参数实验：按照第三章 3.5 节参数生成 SafeCut、段预算、快照、裁决阈值、质押博弈和时间线 trace。

## 对比复现边界

`comparison_protocols` 中每个方案都会写入本地 benchmark 函数、参数、Gas、`dispute_flow` 和复现命令。第三章的对比由本地 Solidity benchmark 与 timeline 状态机产物支撑，不使用第三方部署结果；外部项目资料只用于确定本地复现路径的阶段边界。

| 方案 | 本地复现函数 | 复现路径 |
| --- | --- | --- |
| Arbitrum Classic | `arbitrumOptimisticPath`、`arbitrumClassicPath` | assertion、challenge、bisection、one-step proof |
| TrueBit | `truebitOptimisticPath`、`truebitPath` | solver commitment、verifier challenge、bisection、final judge |
| Cartesi Dave | `cartesiOptimisticPath`、`cartesiDavePath` | tournament、claim/counterclaim、bisection、referee |
| Arbitrum BoLD | `boldOptimisticPath`、`boldPath` | parallel challenge edges、multi-level narrowing、confirmation |
| CleVer | `cleverOptimisticPath`、`cleverPath` | two-layer slice localization、bounded VerSeg replay |

Gas 数值由 `forge test --gas-report` 在统一本地 EVM 环境下测量；时间线槽数由 `cmd/timeline-exp` 写入 `logs/timeline_simulation.json`。`cmd/comparison-exp` 只从 raw log 中的 Foundry 结果和 timeline 模拟日志重新组织对比报告。

图 12 的关键槽数来自 `logs/timeline_simulation.json` 的状态机事件 trace：BoLD `25`、Cartesi Dave `35`、TrueBit `33`、Arbitrum Classic `30`、CleVer `3`。`cmd/clever-exp`、`cmd/comparison-exp` 和 `../visualization/generate_chapter3_data.py` 都会校验这些槽数、事件列表、状态转移和公式输入，避免 raw log、standalone report 与结构化可视化产物之间出现漂移。

## 目录说明

- `cmd/clever-exp/main.go`：实验采集入口。
- `cmd/comparison-exp/main.go`：单独导出争议协议对比报告。
- `cmd/timeline-exp/main.go`：单独运行争议时间线状态机模拟。
- `comparisons/model.go`：Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD、CleVer 的对比实验配置。
- `src/BenchmarkTasks.sol`：四类基准任务合约。
- `src/CleVerVerifier.sol`：最小裁决单元 `VerSeg` 的链上重放裁决接口。
- `src/DisputeProtocolBenchmarks.sol`：Arbitrum/TrueBit/Cartesi/BoLD/CleVer 对比路径 Gas 基准。
- `test/Benchmarks.t.sol`：Foundry Gas 测试入口。
- `logs/raw_experiment_log.json`：实验原始日志。
- `logs/timeline_simulation.json`：图 12 时间线槽数的状态机模拟事件日志。

## 环境配置

需要安装 Go、Foundry、Geth `evm`、Python，以及 Python 包 `matplotlib`、`numpy`。检查命令：

```bash
go version
forge --version
evm --help
python3 -c "import matplotlib, numpy; print('python deps ok')"
```

如果缺少 Python 包：

```bash
cd ..
python3 -m pip install -r requirements.txt
cd experiment
```

## 运行实验

在 `chapter3/experiment` 目录下运行：

```bash
go test ./...
forge test --gas-report
go run ./cmd/timeline-exp --out logs/timeline_simulation.json
go run ./cmd/clever-exp --quick --timeline-out logs/timeline_simulation.json --out logs/raw_experiment_log.json
go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --timeline logs/timeline_simulation.json --out logs/comparison_protocols.json
```

命令与产物说明：

| 命令 | 实验内容 | 产物 | 支撑论文结果 |
| --- | --- | --- | --- |
| `go test ./...` | Go runner、对比协议模型、timeline derivation 测试 | 测试输出 | 验证图 7-12 的模型实现 |
| `forge test --gas-report` | Solidity benchmark 函数级 Gas | Foundry gas report | 图 10、表 4、函数级 Gas 表 |
| `go run ./cmd/timeline-exp --out logs/timeline_simulation.json` | dispute timeline 状态机模拟 | `logs/timeline_simulation.json` | 图 12 槽数和状态转移 |
| `go run ./cmd/clever-exp --quick --timeline-out logs/timeline_simulation.json --out logs/raw_experiment_log.json` | 任务样本、Geth EVM 采样、Foundry Gas 解析、论文尺度 instrumentation trace、timeline 校验 | `logs/raw_experiment_log.json` | 图 7-12、表 3、表 4 |
| `go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --timeline logs/timeline_simulation.json --out logs/comparison_protocols.json` | 将 raw log 与 timeline trace 汇总为 standalone 对比报告 | `logs/comparison_protocols.json` | 图 10、图 12、表 4 的独立验收材料 |

如需运行更大的任务 runner 样本：

```bash
go run ./cmd/clever-exp --full --max-seconds 5 --out logs/raw_experiment_log.json
```

## 查看实验日志

```bash
jq '.paper_evidence.instrumentation_trace | {mode, safe_cut_rule, budget_run_count:(.budget_runs|length), overhead_run_count:(.overhead_runs|length), parameter_run_count:(.parameter_runs|length)}' logs/raw_experiment_log.json
```

期望输出包括：

```json
{
  "mode": "paper-scale-instrumented-evm-loop",
  "budget_run_count": 16,
  "overhead_run_count": 4,
  "parameter_run_count": 8
}
```

查看图 9 默认点的实验 trace：

```bash
jq '[.paper_evidence.instrumentation_trace.parameter_runs[] | select((.value==100000000) or (.value==1000000))]' logs/raw_experiment_log.json
```

期望能看到：

```json
[
  {"kind":"segment_budget","value":100000000,"snapshot_count":1300,"storage_mb":61},
  {"kind":"adjudication_threshold","value":1000000,"subsegment_count":100,"verseg_gas_k":1200}
]
```

查看图 10 对比协议实验：

```bash
jq '.comparison_protocols[] | {scheme, optimistic:(.optimistic_benchmark.function), dispute:(.dispute_benchmark.function), args:.dispute_benchmark.args, dispute_flow, dispute_gas_k, reproduce}' logs/raw_experiment_log.json
jq '.protocols[] | {scheme, dispute:.dispute_benchmark.function, dispute_gas_k, dispute_flow}' logs/comparison_protocols.json
```

查看图 12 时间线模拟：

```bash
jq '.protocols[] | {scheme, slots:.timeline_derivation.slots, events:(.timeline_derivation.events|length), formula:.timeline_derivation.slot_formula}' logs/timeline_simulation.json
jq '.paper_evidence.timeline.schemes[] | {name, slots:.dispute_slots, total_time, source:.timeline_derivation.simulation_kind}' logs/raw_experiment_log.json
```

## 生成可视化

回到 `chapter3` 目录运行：

```bash
cd ..
python3 visualization/generate_chapter3_data.py
python3 visualization/chapter3_all_figures.py
```

输出：

- `visualization/chapter3_experiment_data.json`
- `experiment/logs/timeline_simulation.json`
- `experiment/logs/comparison_protocols.json`
- `visualization/正确图片输出/*.png`

结构化 JSON 与最终图片对应关系如下。

| 结构化字段 | 生成依据 | 输出图片/结果 |
| --- | --- | --- |
| `budget_compliance` | SafeCut budget traces | 图 7 `fig1_budget_compliance.png` |
| `slicing_overhead` | runner instrumentation overhead traces | 图 8 `fig2_overhead.png` |
| `parameter_sensitivity` | `B`、`b` 参数扫描 trace | 图 9 `fig_param_sensitivity_v2.png` |
| `gas_comparison` | Foundry gas report + `comparison_protocols` | 图 10 `fig_gas_comparison_v2.png` |
| `staking_analysis` | 超线性质押公式 trace | 图 11 `fig_staking_analysis_v2.png` |
| `timeline`、`timeline_simulation` | dispute timeline 状态机事件 trace | 图 12 `fig_timeline_v3.png` |

## 注意事项

- 如果 `go test` 或 `go run` 无法写入默认 Go cache，可设置 `GOCACHE` 到本地临时目录。
- 如果 Matplotlib 提示配置目录不可写，可设置 `MPLCONFIGDIR` 到本地临时目录。
- 重新生成 raw log 后，需要重新生成结构化实验产物和图片，使 raw log、timeline trace、standalone report 与 PNG 保持一致。

## 验收流程

1. 在本目录运行 `go test ./...` 和 `forge test --gas-report`。
2. 运行 `go run ./cmd/timeline-exp --out logs/timeline_simulation.json`，重新生成图 12 时间线状态机模拟日志。
3. 运行 `go run ./cmd/clever-exp --quick --timeline-out logs/timeline_simulation.json --out logs/raw_experiment_log.json`，重新生成 raw log。
4. 运行 `go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --timeline logs/timeline_simulation.json --out logs/comparison_protocols.json`，导出对比报告。
5. 回到 `chapter3` 运行 `python3 visualization/generate_chapter3_data.py` 和 `python3 visualization/chapter3_all_figures.py`。
6. 回到仓库根目录运行 `python3 scripts/check_experiment_coverage.py`。
