# Chapter 3 Experiment Reproduction Guide

本目录对应论文第三章“基于有状态任务切片的链下计算验证”的实验与可视化。实验链路的核心约束是：论文图表使用的结构化数据必须来自 `chapter3/experiment/logs/raw_experiment_log.json`，而 raw log 由 Go 本地执行采样、Geth `evm --bench run`、Foundry Gas 测试和 paper-scale instrumentation trace 生成，不应在绘图脚本中脱离实验硬编码。

## 实验内容

第三章实验覆盖论文 3.5 节的三个问题：

- 自适应切片机制：验证段预算不越界、切片开销小于 10%、段预算 `B` 与裁决阈值 `b` 对系统开销的影响。
- 争议定位解决效率：比较 Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD 和 CleVer 在乐观路径与争议路径下的链上 Gas。
- 伴随式验证时效性：比较 CleVer 与事后验证方案在错误注入后的归一化端到端时间。

论文默认参数与代码一致：

- 段预算 `B = 1e8 Gas`
- 裁决阈值 `b = 1e6 Gas`
- 密度控制参数 `alpha = 0.8`
- 最大质押轮次 `g = 10`
- 二次型累计质押曲线 `D(r) = d0 * r^2`，`d0 = 0.5 ETH`

## 目录结构

- `第三章基于有状态任务切片的链下计算验证.pdf`：第三章论文正文，实验部分为 3.5 节。
- `experiment/README.md`：实验工程的简要说明。
- `experiment/cmd/clever-exp/main.go`：raw log 生成入口。它执行本地任务样本、调用 Geth EVM benchmark、调用 Foundry Gas report，并生成论文尺度的 instrumentation trace。
- `experiment/src/*.sol`：Solidity 基准任务、CleVer `VerSeg` 裁决合约和对比协议 Gas 基准。
- `experiment/test/Benchmarks.t.sol`：Foundry Gas 测试入口。
- `experiment/logs/raw_experiment_log.json`：实验原始日志。
- `visualization/generate_chapter3_data.py`：从 raw log 校验并生成结构化可视化数据。
- `visualization/chapter3_experiment_data.json`：结构化数据，绘图脚本唯一数据入口。
- `visualization/chapter3_all_figures.py`：生成论文图 7-12。
- `visualization/verify_paper_alignment.py`：生成论文对齐报告。
- `visualization/paper_alignment_report.md`：图表、参数、关键结论与论文文字的对齐检查结果。
- `visualization/正确图片输出/`：生成的 PNG 图片。

## 环境依赖

需要以下工具：

- Go 1.21 或更高版本。
- Foundry，包括 `forge`。
- Geth 的 `evm` 命令。
- Python 3.10 或兼容版本。
- Python 包：`matplotlib`、`numpy`。

检查命令：

```bash
go version
forge --version
evm --help
python3 --version
python3 -c "import matplotlib, numpy; print('python deps ok')"
```

如果网络不通，安装 Python 依赖前可按仓库说明使用代理：

```bash
export https_proxy=http://127.0.0.1:33210 http_proxy=http://127.0.0.1:33210 all_proxy=socks5://127.0.0.1:33211
python3 -m pip install matplotlib numpy
```

在某些 macOS 沙盒环境中，Go 或 Matplotlib 默认缓存目录不可写，可改用临时目录：

```bash
export GOCACHE=/private/tmp/go-build-ch3
export MPLCONFIGDIR=/private/tmp/mplconfig_ch3
mkdir -p "$GOCACHE" "$MPLCONFIGDIR"
```

## 运行完整实验

从仓库根目录运行：

```bash
cd chapter3/experiment
go test ./...
forge test --gas-report
go run ./cmd/clever-exp --quick --out logs/raw_experiment_log.json

cd ../..
python3 chapter3/visualization/generate_chapter3_data.py
python3 chapter3/visualization/chapter3_all_figures.py
python3 chapter3/visualization/verify_paper_alignment.py
```

说明：

- `go test ./...` 检查 Go 实验 runner 能正常编译。
- `forge test --gas-report` 独立检查 Solidity 基准和 Gas report。
- `go run ./cmd/clever-exp ...` 会重新生成 `raw_experiment_log.json`；该命令内部也会调用 `forge test --gas-report`，并调用 `evm --bench run` 采集 Geth EVM 样本。
- `generate_chapter3_data.py` 从 raw log 读取并校验 `samples`、`geth_evm_samples`、`comparison_protocols` 和 `paper_evidence`，然后写出 `chapter3_experiment_data.json`。
- `chapter3_all_figures.py` 只读取 `chapter3_experiment_data.json`，不直接从脚本中硬编码图表数据。
- `verify_paper_alignment.py` 生成 `paper_alignment_report.md`，用于确认图 7-12 的关键结论与论文 3.5 节一致。

如需运行更大的本地样本：

```bash
cd chapter3/experiment
go run ./cmd/clever-exp --full --max-seconds 5 --out logs/raw_experiment_log.json
```

`--full` 仍会受 `--max-seconds` 限制；论文尺度的长任务通过 raw log 中的 instrumentation trace 记录和校准，避免把小时级任务强制跑完。

## 只重新生成可视化

如果 `experiment/logs/raw_experiment_log.json` 已经存在，只想刷新结构化数据、图片和对齐报告：

```bash
python3 chapter3/visualization/generate_chapter3_data.py
python3 chapter3/visualization/chapter3_all_figures.py
python3 chapter3/visualization/verify_paper_alignment.py
```

如果只改了绘图样式，并且不需要刷新 JSON：

```bash
python3 chapter3/visualization/chapter3_all_figures.py
python3 chapter3/visualization/verify_paper_alignment.py
```

## 输出文件位置

- Raw log：`chapter3/experiment/logs/raw_experiment_log.json`
- 结构化数据：`chapter3/visualization/chapter3_experiment_data.json`
- 论文对齐报告：`chapter3/visualization/paper_alignment_report.md`
- 生成图片：
  - `chapter3/visualization/正确图片输出/fig1_budget_compliance.png`
  - `chapter3/visualization/正确图片输出/fig2_overhead.png`
  - `chapter3/visualization/正确图片输出/fig_param_sensitivity_v2.png`
  - `chapter3/visualization/正确图片输出/fig_gas_comparison_v2.png`
  - `chapter3/visualization/正确图片输出/fig_staking_analysis_v2.png`
  - `chapter3/visualization/正确图片输出/fig_timeline_v3.png`

## 数据链路

完整链路如下：

```text
experiment/cmd/clever-exp/main.go
  -> experiment/logs/raw_experiment_log.json
  -> visualization/generate_chapter3_data.py
  -> visualization/chapter3_experiment_data.json
  -> visualization/chapter3_all_figures.py
  -> visualization/正确图片输出/*.png
  -> visualization/verify_paper_alignment.py
  -> visualization/paper_alignment_report.md
```

raw log 的主要来源：

- `samples`：Go 本地确定性任务执行，记录 steps、snapshot bytes 和 SHA-256 commitment。
- `geth_evm_samples`：`evm --bench run` 对四类代表性 EVM bytecode 的 Gas、执行时间和内存分配采样。
- `foundry_gas_runs`：`forge test --gas-report` 解析得到的 Solidity 基准 Gas。
- `comparison_protocols`：由 Foundry Gas 结果派生的五种协议乐观路径和争议路径 Gas。
- `paper_evidence.instrumentation_trace`：面向论文尺度任务的 SafeCut、切片开销、参数敏感性、质押博弈和时间线 trace。该分支在 Go runner 中计算，然后进入 JSON 和绘图流程。

判断数据是否脱离实验硬编码时，重点检查：

```bash
jq '.samples | length' chapter3/experiment/logs/raw_experiment_log.json
jq '.geth_evm_samples | length' chapter3/experiment/logs/raw_experiment_log.json
jq '.comparison_protocols | length' chapter3/experiment/logs/raw_experiment_log.json
jq '.paper_evidence.instrumentation_trace | {mode, budget_runs:(.budget_runs|length), overhead_runs:(.overhead_runs|length), parameter_runs:(.parameter_runs|length)}' chapter3/experiment/logs/raw_experiment_log.json
```

期望分别看到 4 个本地任务样本、4 个 Geth EVM 样本、5 个协议对比项，以及 16 个 budget traces、4 个 overhead traces、8 个 parameter traces。

## 论文图表对应关系

| 论文图表 | 输出图片 | 数据字段 | 生成/测量代码 |
| --- | --- | --- | --- |
| 图 7 段权重占预算百分比 | `fig1_budget_compliance.png` | `budget_compliance.samples_percent`、`means_percent`、`errors_percent` | `runInstrumentedBudgetTrace`、`runSafeCutGasSegment`、`fig_budget_compliance` |
| 图 8 任务切片执行开销 | `fig2_overhead.png` | `slicing_overhead` | `runInstrumentedOverheadTrace`、`summarizeOverhead`、`fig_overhead` |
| 图 9 段预算与裁决阈值影响 | `fig_param_sensitivity_v2.png` | `parameter_sensitivity` | `deriveParameterSensitivity`、`plot_param_sensitivity` |
| 图 10 链上开销对比 | `fig_gas_comparison_v2.png` | `gas_comparison` | `DisputeProtocolBenchmarks.sol`、`Benchmarks.t.sol`、`runFoundryGasReport`、`runComparisonProtocols`、`fig_gas_comparison` |
| 图 11 超线性累计质押博弈 | `fig_staking_analysis_v2.png` | `staking_analysis` | `deriveStakingAnalysis`、`fig_staking_analysis` |
| 图 12 伴随式验证时间对比 | `fig_timeline_v3.png` | `timeline` | `deriveTimeline`、`fig_timeline` |

## 对齐检查要点

`paper_alignment_report.md` 应至少满足：

- 图 7：所有段权重均不超过 `B`，多数段集中在 `alpha B` 附近。
- 图 8：四类任务切片开销均小于 10%。
- 图 9：默认 `B=1e8` 时 Sort-Large 为 1300 个快照、约 61 MB；默认 `b=1e6` 时 `L=100`，`VerSeg` 约 1200 K Gas。
- 图 10：CleVer 争议路径 Gas 相比四个对比方案平均值降低约 87%。
- 图 11：`p >= 0.7` 且 `beta ~= 2` 时错误方在前几轮退出。
- 图 12：CleVer 总时间为 `1.044 T_exec`。

当前报告保留一个需要人工解释的 `WARN`：论文正文写“端到端验证时延降低约 52%”，而图 12 总时间数值按对比方案平均值直接计算约为 61%。报告同时列出这两个口径，避免静默掩盖论文文字与图中算术之间的差异。

## 常见问题和注意事项

- `go test ./...` 或 `go run` 若无法写入默认 Go cache，可设置 `GOCACHE=/private/tmp/go-build-ch3`。
- Matplotlib 若提示默认配置目录不可写，可设置 `MPLCONFIGDIR=/private/tmp/mplconfig_ch3`。
- `evm` 缺失会导致 raw log 中 Geth EVM 样本无法通过校验；安装 Geth 后重新运行完整实验。
- `forge test --gas-report` 输出格式会被 `main.go` 解析，若 Foundry 版本改变导致解析不到函数名，`go run` 会因缺失 Gas 项失败。
- `out/`、`cache/`、`node_modules/`、`__pycache__/` 等可重建产物不要提交；仓库 `.gitignore` 已忽略 Foundry build/cache 和常见缓存目录。
- 重新生成 raw log 后，应按顺序重新生成结构化数据、图片和对齐报告。
