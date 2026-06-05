# 第三章 CleVer 实验工程

本目录用于运行第三章 CleVer 实验，生成 `logs/raw_experiment_log.json`，并为 `../visualization/` 提供图 7-12 所需的数据输入。

## 实验内容

- 任务 runner：运行 Fibonacci、Poly-Chain、Sort-Large、DP-Large 四类任务样本，记录执行步数、快照大小和状态承诺。
- EVM 执行采样：调用 Geth `evm --bench run`，记录不同任务字节码的 Gas、执行时间和内存分配情况。
- 链上 Gas 测量：调用 Foundry `forge test --gas-report`，统计 `VerSeg` 和对比协议路径的 Gas 开销。
- 对比协议实验：`comparisons/model.go` 将 Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD、CleVer 的对比路径绑定到具体 Solidity benchmark 函数、参数和时间线模型。
- 参数实验：按照第三章 3.5 节参数生成 SafeCut、段预算、快照、裁决阈值、质押博弈和时间线 trace。

## 对比复现边界

`comparison_protocols` 中每个方案都会写入本地 benchmark 函数、参数、Gas、`dispute_flow` 和复现命令。第三章的对比不读取外部表格，也不使用第三方部署结果；外部项目资料只用于确定本地复现路径的阶段边界。

| 方案 | 本地复现函数 | 复现路径 |
| --- | --- | --- |
| Arbitrum Classic | `arbitrumOptimisticPath`、`arbitrumClassicPath` | assertion、challenge、bisection、one-step proof |
| TrueBit | `truebitOptimisticPath`、`truebitPath` | solver commitment、verifier challenge、bisection、final judge |
| Cartesi Dave | `cartesiOptimisticPath`、`cartesiDavePath` | tournament、claim/counterclaim、bisection、referee |
| Arbitrum BoLD | `boldOptimisticPath`、`boldPath` | parallel challenge edges、multi-level narrowing、confirmation |
| CleVer | `cleverOptimisticPath`、`cleverPath` | two-layer slice localization、bounded VerSeg replay |

Gas 数值由 `forge test --gas-report` 在统一本地 EVM 环境下测量，`cmd/comparison-exp` 只从 raw log 中的 Foundry 结果重新组织对比报告。

## 目录说明

- `cmd/clever-exp/main.go`：实验采集入口。
- `cmd/comparison-exp/main.go`：单独导出争议协议对比报告。
- `comparisons/model.go`：Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD、CleVer 的对比实验配置。
- `src/BenchmarkTasks.sol`：四类基准任务合约。
- `src/CleVerVerifier.sol`：最小裁决单元 `VerSeg` 的链上重放裁决接口。
- `src/DisputeProtocolBenchmarks.sol`：Arbitrum/TrueBit/Cartesi/BoLD/CleVer 对比路径 Gas 基准。
- `test/Benchmarks.t.sol`：Foundry Gas 测试入口。
- `logs/raw_experiment_log.json`：实验原始日志。

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
go run ./cmd/clever-exp --quick --out logs/raw_experiment_log.json
go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_protocols.json
```

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

## 生成可视化

回到 `chapter3` 目录运行：

```bash
cd ..
python3 visualization/generate_chapter3_data.py
python3 visualization/chapter3_all_figures.py
```

输出：

- `visualization/chapter3_experiment_data.json`
- `experiment/logs/comparison_protocols.json`
- `visualization/正确图片输出/*.png`

## 注意事项

- 如果 `go test` 或 `go run` 无法写入默认 Go cache，可设置 `GOCACHE` 到本地临时目录。
- 如果 Matplotlib 提示配置目录不可写，可设置 `MPLCONFIGDIR` 到本地临时目录。
- 重新生成 raw log 后，需要重新运行数据生成和绘图脚本。

## 验收流程

1. 在本目录运行 `go test ./...` 和 `forge test --gas-report`。
2. 运行 `go run ./cmd/clever-exp --quick --out logs/raw_experiment_log.json`，重新生成 raw log。
3. 运行 `go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_protocols.json`，导出对比报告。
4. 回到 `chapter3` 运行 `python3 visualization/generate_chapter3_data.py` 和 `python3 visualization/chapter3_all_figures.py`。
5. 回到仓库根目录运行 `python3 scripts/check_experiment_coverage.py`。
