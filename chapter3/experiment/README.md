# 第三章 CleVer 实验工程

本目录用于运行第三章 CleVer 实验，生成实验原始日志，并为可视化脚本提供图 7-12 所需的数据输入。

## 实验内容

实验流程包含以下部分：

- 本地任务执行：运行 Fibonacci、Poly-Chain、Sort-Large、DP-Large 四类任务样本，记录执行步数、快照大小和状态承诺。
- EVM 执行采样：调用 Geth `evm --bench run`，记录不同任务字节码的 Gas、执行时间和内存分配情况。
- 链上 Gas 测量：调用 Foundry `forge test --gas-report`，统计 `VerSeg` 和对比协议路径的 Gas 开销。
- 论文参数实验：按照第三章 3.5 节的默认参数生成 SafeCut、段预算、快照、裁决阈值、质押博弈和时间线相关 trace。

## 目录说明

- `cmd/clever-exp/main.go`：实验采集入口，生成 `raw_experiment_log.json`。
- `src/BenchmarkTasks.sol`：四类基准任务合约，对应 Fibonacci、Poly-Chain、Sort-Large、DP-Large。
- `src/CleVerVerifier.sol`：最小裁决单元 `VerSeg` 的链上重放裁决接口。
- `src/DisputeProtocolBenchmarks.sol`：Arbitrum/TrueBit/Cartesi/BoLD/CleVer 对比路径 Gas 基准。
- `test/Benchmarks.t.sol`：Foundry Gas 测试入口。
- `logs/raw_experiment_log.json`：实验原始日志。

## 环境配置

需要安装以下工具：

- Go 1.21 或更高版本。
- Foundry，包括 `forge`。
- Geth 的 `evm` 命令，用于 `evm --bench run`。
- Python 3.10 或更高版本。
- Python 包：`matplotlib`、`numpy`。

检查命令：

```bash
go version
forge --version
evm --help
python3 --version
python3 -c "import matplotlib, numpy; print('python deps ok')"
```

如果缺少 Python 包：

```bash
python3 -m pip install matplotlib numpy
```

## 运行完整实验

在 `chapter3/experiment` 目录下运行：

```bash
go test ./...
forge test --gas-report
go run ./cmd/clever-exp --quick --out logs/raw_experiment_log.json
```

`go run` 会输出本地任务样本、Geth EVM 样本、Foundry Gas 对比结果，以及图表数据中的关键节点。例如：

```text
[sample] Fibonacci  steps=5000 snapshot=784B commitment=cd1f6b79 elapsed=0.001s
[geth-evm] Sort-Large gas=7591 time=42.072µs alloc=38 bytes=3867
[foundry] verSeg gas=55444
[compare] CleVer (ours) rounds=3 optimistic=314K dispute=552K
[paper] figure9 default snapshots=1300 storage=61.0MB L=100 VerSeg=1200K
[paper] figure10 dispute reduction=87.62%
[paper] figure12 CleVer total=1.044 T_exec
```

如需运行更大的本地任务样本，可以使用：

```bash
go run ./cmd/clever-exp --full --max-seconds 5 --out logs/raw_experiment_log.json
```

## 查看实验日志

生成 raw log 后，可查看图表 trace 的概要信息：

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

## 生成可视化结果

在仓库根目录 `第三章` 下运行：

```bash
python3 chapter3/visualization/generate_chapter3_data.py
python3 chapter3/visualization/chapter3_all_figures.py
python3 chapter3/visualization/verify_paper_alignment.py
```

输出文件：

- `chapter3/visualization/chapter3_experiment_data.json`：从 raw log 提取并校验后的可视化数据。
- `chapter3/visualization/正确图片输出/`：图 7-12 的 PNG 文件。
- `chapter3/visualization/paper_alignment_report.md`：论文实验分析对齐报告。

## 结果检查

`verify_paper_alignment.py` 会检查以下关键点：

- 图 7：段权重占预算比例不超过 100%，且多数段集中在 `alpha B` 附近。
- 图 8：四类任务切片开销均小于 10%。
- 图 9：默认 `B=10^8` 时 Sort-Large 为 `1300` 个快照、`61MB`；默认 `b=10^6` 时 `L=100`、`VerSeg=1200K Gas`。
- 图 10：CleVer 争议路径 Gas 相比四个对比方案平均值降低约 87%。
- 图 11：`p>=0.7`、`beta≈2` 时错误方早期退出。
- 图 12：CleVer 总时间为 `1.044 T_exec`。

报告中所有关键项应为 `PASS`。

## 注意事项

- `evm --bench run` 与 `forge test --gas-report` 的输出会受到本地工具版本和机器环境影响，Gas 数值通常稳定，执行时间可能有轻微波动。
- 重新生成数据后，需要按顺序运行数据生成、绘图和对齐检查脚本。
- 若修改实验参数，请先重新生成 `logs/raw_experiment_log.json`，再生成可视化数据和图片。
