# 第三章实验复现说明

本目录对应论文第三章“基于有状态任务切片的链下计算验证”。实验覆盖论文 3.5 节与图 7-12，工程链路为：运行 CleVer 实验入口生成 raw log，再从 raw log 生成结构化数据和 PNG 图片。

## 实验内容

- 自适应切片机制：验证段预算不越界、切片开销小于 10%、段预算 `B` 与裁决阈值 `b` 对系统开销的影响。
- 争议定位解决效率：比较 Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD 和 CleVer 在乐观路径与争议路径下的链上 Gas。
- 伴随式验证时效性：比较 CleVer 与事后验证方案在错误注入后的归一化端到端时间。

默认参数：

| 参数 | 值 |
| --- | --- |
| 段预算 `B` | `1e8 Gas` |
| 裁决阈值 `b` | `1e6 Gas` |
| 密度控制参数 `alpha` | `0.8` |
| 最大质押轮次 `g` | `10` |
| 累计质押曲线 | `D(r) = d0 * r^2`, `d0 = 0.5 ETH` |

## 目录结构

- `第三章基于有状态任务切片的链下计算验证.pdf`：第三章论文正文。
- `experiment/cmd/clever-exp/main.go`：raw log 生成入口。
- `experiment/src/*.sol`：Solidity 基准任务、CleVer 裁决合约和对比协议 Gas 基准。
- `experiment/test/Benchmarks.t.sol`：Foundry Gas 测试入口。
- `experiment/logs/raw_experiment_log.json`：实验原始日志。
- `visualization/generate_chapter3_data.py`：从 raw log 校验并生成结构化可视化数据。
- `visualization/chapter3_experiment_data.json`：绘图脚本的数据入口。
- `visualization/chapter3_all_figures.py`：生成图 7-12 的 PNG。
- `visualization/正确图片输出/`：生成图片目录。

## 环境依赖

需要以下工具：

- Go 1.21 或更高版本。
- Foundry，包括 `forge`。
- Geth 的 `evm` 命令。
- Python 3.10 或兼容版本。
- Python 包：`matplotlib`、`numpy`。

进入 `chapter3` 目录后，使用当前 `python3` 安装 Python 依赖：

```bash
which python3
python3 --version
python3 -m pip install -r requirements.txt
```

如果遇到 `externally-managed-environment`，说明当前 Python 不允许直接写入系统环境，可显式允许 pip 安装到当前 Python 环境：

```bash
python3 -m pip install --break-system-packages -r requirements.txt
```

如果 Go 或 Matplotlib 缓存目录不可写：

```bash
export GOCACHE="${TMPDIR:-/tmp}/go-build-ch3"
export MPLCONFIGDIR="${TMPDIR:-/tmp}/mplconfig-ch3"
mkdir -p "$GOCACHE" "$MPLCONFIGDIR"
```

检查命令：

```bash
go version
forge --version
evm --help
python3 --version
python3 -c "import matplotlib, numpy; print('python deps ok')"
```

## 运行完整实验

从 `chapter3` 目录运行：

```bash
cd experiment
go test ./...
forge test --gas-report
go run ./cmd/clever-exp --quick --out logs/raw_experiment_log.json

cd ..
python3 visualization/generate_chapter3_data.py
python3 visualization/chapter3_all_figures.py
```

说明：

- `go test ./...` 检查 Go 实验 runner 和本地任务逻辑。
- `forge test --gas-report` 独立检查 Solidity 基准与 Gas report。
- `go run ./cmd/clever-exp ...` 重新生成 `raw_experiment_log.json`，内部会调用 Foundry Gas report 和 Geth `evm --bench run`。
- `generate_chapter3_data.py` 从 raw log 读取并校验 `samples`、`geth_evm_samples`、`comparison_protocols` 和 `paper_evidence`。
- `chapter3_all_figures.py` 只读取 `chapter3_experiment_data.json` 绘图。

如需运行更大的本地样本：

```bash
cd experiment
go run ./cmd/clever-exp --full --max-seconds 5 --out logs/raw_experiment_log.json
```

`--full` 仍受 `--max-seconds` 限制；论文尺度任务通过 raw log 中的 instrumentation trace 记录和校准，避免强制跑完小时级任务。

## 只重新生成可视化

如果 `experiment/logs/raw_experiment_log.json` 已存在：

```bash
python3 visualization/generate_chapter3_data.py
python3 visualization/chapter3_all_figures.py
```

如果只改了绘图样式，且不需要刷新 JSON：

```bash
python3 visualization/chapter3_all_figures.py
```

## 输出文件

- Raw log：`experiment/logs/raw_experiment_log.json`
- 结构化数据：`visualization/chapter3_experiment_data.json`
- 生成图片：
  - `visualization/正确图片输出/fig1_budget_compliance.png`
  - `visualization/正确图片输出/fig2_overhead.png`
  - `visualization/正确图片输出/fig_param_sensitivity_v2.png`
  - `visualization/正确图片输出/fig_gas_comparison_v2.png`
  - `visualization/正确图片输出/fig_staking_analysis_v2.png`
  - `visualization/正确图片输出/fig_timeline_v3.png`

## 数据链路

```text
experiment/cmd/clever-exp/main.go
  -> experiment/logs/raw_experiment_log.json
  -> visualization/generate_chapter3_data.py
  -> visualization/chapter3_experiment_data.json
  -> visualization/chapter3_all_figures.py
  -> visualization/正确图片输出/*.png
```

raw log 的主要来源：

- `samples`：Go 本地确定性任务执行，记录 steps、snapshot bytes 和 SHA-256 commitment。
- `geth_evm_samples`：`evm --bench run` 对四类代表性 EVM bytecode 的 Gas、执行时间和内存分配采样。
- `foundry_gas_runs`：`forge test --gas-report` 解析得到的 Solidity 基准 Gas。
- `comparison_protocols`：由 Foundry Gas 结果派生的五种协议乐观路径和争议路径 Gas。
- `paper_evidence.instrumentation_trace`：SafeCut、切片开销、参数敏感性、质押博弈和时间线 trace。

快速检查 raw log：

```bash
jq '.samples | length' experiment/logs/raw_experiment_log.json
jq '.geth_evm_samples | length' experiment/logs/raw_experiment_log.json
jq '.comparison_protocols | length' experiment/logs/raw_experiment_log.json
jq '.paper_evidence.instrumentation_trace | {mode, budget_runs:(.budget_runs|length), overhead_runs:(.overhead_runs|length), parameter_runs:(.parameter_runs|length)}' experiment/logs/raw_experiment_log.json
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

## 注意事项

- `evm --bench run` 与 `forge test --gas-report` 的执行时间可能随机器环境波动，Gas 数值通常稳定。
- `evm` 缺失会导致 raw log 中 Geth EVM 样本无法通过校验。
- 重新生成 raw log 后，应重新运行数据生成与绘图脚本。
