# 第三章实验复现说明

本目录对应论文第三章“基于有状态任务切片的链下计算验证”。实验覆盖论文 3.5 节与图 7-12，目标是复现 CleVer 的任务切片、争议定位、链上 Gas、超线性质押和伴随式验证时间线结果。

实验链路为：

```text
experiment/cmd/clever-exp/main.go
  -> experiment/logs/raw_experiment_log.json
  -> visualization/generate_chapter3_data.py
  -> visualization/chapter3_experiment_data.json
  -> visualization/chapter3_all_figures.py
  -> visualization/正确图片输出/*.png
```

图 10 的协议对比不读取外部表格，也不复用第三方主网部署数据。本仓库在 `experiment/src/DisputeProtocolBenchmarks.sol` 中实现同一 EVM 口径下的 submit/challenge/localize/adjudicate 可执行路径，再由 `forge test --gas-report` 测量函数级 Gas；raw log 记录每个方案对应的 Solidity 函数、参数、Gas、时间线推导和复现实验命令。

## 实验总览

默认参数如下。

| 参数 | 值 | 用途 |
| --- | --- | --- |
| 段预算 `B` | `1e8 Gas` | 控制每个任务切片段的最大 Gas 权重 |
| 裁决阈值 `b` | `1e6 Gas` | 控制二次细分后单个 VerSeg 裁决单元大小 |
| 密度控制参数 `alpha` | `0.8` | SafeCut 提前切片触发阈值 |
| 最大质押轮次 `g` | `10` | 超线性累计质押博弈的最大轮数 |
| 累计质押曲线 | `D(r)=d0*r^2`, `d0=0.5 ETH` | 错误方延迟退出时的递增质押成本 |

本章使用四类基准任务。

| 任务 | 论文尺度 Gas | 归一化执行时间 | runner 实测样本 |
| --- | ---: | ---: | --- |
| Fibonacci | `1e9` | `10 s` | `5000` steps，快照 `784 B` |
| Poly-Chain | `1e10` | `100 s` | `4000` steps，快照 `784 B` |
| Sort-Large | `1e11` | `1000 s` | `73536` steps，快照 `784 B` |
| DP-Large | `1e12` | `10000 s` | `11999` steps，快照 `784 B` |

runner 样本用于证明任务执行、快照序列化和 commitment 生成链路可复现；长任务曲线由 raw log 中的 instrumentation trace 按表 3 参数展开。

## 对比复现边界

第三章的外部方案对比是“同口径本地 benchmark”：本仓库不声称移植第三方完整生产实现，而是把各方案在论文中要比较的争议阶段落到本地 Solidity 函数，统一用 Foundry 测量。这样图 10 展示的是同一测试环境下的提交、挑战、定位、裁决路径开销。

| 方案 | 本地复现函数 | 在本实验中的路径 |
| --- | --- | --- |
| Arbitrum Classic | `arbitrumOptimisticPath`、`arbitrumClassicPath` | assertion 提交、一对一 challenge、二分定位和 one-step proof |
| TrueBit | `truebitOptimisticPath`、`truebitPath` | solver 提交、verifier challenge、交互式二分和 final judge |
| Cartesi Dave | `cartesiOptimisticPath`、`cartesiDavePath` | tournament setup、claim/counterclaim 二分和 referee |
| Arbitrum BoLD | `boldOptimisticPath`、`boldPath` | parallel challenge edges、多层定位和 lowest-level edge confirmation |
| CleVer | `cleverOptimisticPath`、`cleverPath`、`CleVerVerifier.verSeg` | 两层切片定位和有界 VerSeg 重放 |

## 测试内容与结果

| 测试/图表 | 测试对象 | 数据来源 | 当前关键结果 | 输出 |
| --- | --- | --- | --- | --- |
| 图 7 段权重占预算百分比 | 自适应切片是否严格满足 `W(Seg_i) <= B` | `paper_evidence.budget_compliance`，16 组任务/预算 trace | 段权重占预算比例范围为 `80%` 到 `99.9%`，未超过 `100%` | `visualization/正确图片输出/fig1_budget_compliance.png` |
| 图 8 任务切片执行开销 | SafeCut 检查、快照序列化、承诺计算引入的额外执行时间 | `paper_evidence.slicing_overhead` | 四类任务开销分别为 `2.6%`、`2.7%`、`3.5%`、`9.0%`，均小于 `10%` | `visualization/正确图片输出/fig2_overhead.png` |
| 图 9a 段预算影响 | 固定 `b=1e6` 时，`B` 对快照数量和链下存储的影响 | `paper_evidence.parameter_sensitivity.snapshot_count/storage_mb` | `B=1e6,1e7,1e8,1e9` 时快照数为 `130000,13000,1300,130`；存储为 `6100,610,61,6.1 MB` | `visualization/正确图片输出/fig_param_sensitivity_v2.png` |
| 图 9b 裁决阈值影响 | 固定 `B=1e8` 时，`b` 对子段数和 VerSeg Gas 的影响 | `paper_evidence.parameter_sensitivity.subsegment_count/verseg_gas_k` | `b=1e4,1e5,1e6,1e7` 时 `L=10000,1000,100,10`；VerSeg 为 `12K,120K,1200K,12000K Gas` | `visualization/正确图片输出/fig_param_sensitivity_v2.png` |
| 图 10 链上开销对比 | Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD、CleVer 的乐观/争议路径 Gas | `comparison_protocols`，由 Foundry Gas report 派生 | CleVer 乐观路径 `389.652K Gas`；争议路径 `551.974K Gas`，较其他四种争议路径平均 `4445.867K Gas` 降低 `87.58%` | `visualization/正确图片输出/fig_gas_comparison_v2.png` |
| 图 11 超线性质押博弈 | 失败信念 `p`、指数 `beta` 对错误方退出轮次和收益的影响 | `paper_evidence.staking_analysis` | `p` 或 `beta` 越高，错误方越早退出；二次质押曲线下坚持到最终裁决收益固定，退出收益随轮次单调下降 | `visualization/正确图片输出/fig_staking_analysis_v2.png` |
| 图 12 伴随式验证时间 | CleVer 与事后验证方案在 `eta=0.4` 错误注入下的归一化端到端时间 | `paper_evidence.timeline` | CleVer 为 `1.044 T_exec`；BoLD、Cartesi、TrueBit、Arbitrum Classic 分别为 `2.650`、`2.730`、`2.714`、`2.690 T_exec` | `visualization/正确图片输出/fig_timeline_v3.png` |

## 代码与论文结果定位

论文第三章 3.5 节的表 3、表 4 和图 7-12 可按下表定位到具体代码。整体流程是先由 Go/Foundry 生成 `experiment/logs/raw_experiment_log.json`，再由 `visualization/generate_chapter3_data.py` 校验并导出 `visualization/chapter3_experiment_data.json`，最后由 `visualization/chapter3_all_figures.py` 出图。

| 论文结果 | 负责生成或测量的代码 | JSON 数据字段 | 最终输出 |
| --- | --- | --- | --- |
| 表 3 测试基准任务 | `experiment/cmd/clever-exp/main.go` 中的 `runPaperEvidence()` 定义论文尺度任务；`runPhysicalSamples()` 和 `executeTaskSample()` 跑本地确定性样本；`experiment/src/BenchmarkTasks.sol` 是 Solidity 任务基准 | `paper_evidence.workloads`、`samples`、`foundry_gas_runs` | README 中的“本章使用四类基准任务”和 Foundry 函数级 Gas 表 |
| 表 4 典型方案对比 | `experiment/comparisons/model.go` 定义五种方案的路径参数、时间线槽数和复现命令；`experiment/src/DisputeProtocolBenchmarks.sol` 实现被测路径；`experiment/cmd/comparison-exp` 可单独导出对比报告 | `comparison_protocols`、`logs/comparison_protocols.json` | README 中的方案说明；图 10 使用同一组方案顺序 |
| 图 7 段权重占预算百分比 | `runPaperEvidence()` 调度 4 类任务和 4 个 `B`；`runInstrumentedBudgetTrace()` 与 `runSafeCutGasSegment()` 生成 SafeCut 段权重 trace；`fig_budget_compliance()` 绘制箱线图 | `paper_evidence.budget_compliance` -> `budget_compliance.samples_percent` | `visualization/正确图片输出/fig1_budget_compliance.png` |
| 图 8 任务切片执行开销 | `runInstrumentedOverheadTrace()` 生成无切片/切片时间、SafeCut、快照序列化、承诺计算占比；`summarizeOverhead()` 汇总；`fig_overhead()` 绘图 | `paper_evidence.slicing_overhead` -> `slicing_overhead` | `visualization/正确图片输出/fig2_overhead.png` |
| 图 9 段预算与裁决阈值影响 | `deriveParameterSensitivity()` 扫描 `B={1e6,1e7,1e8,1e9}` 与 `b={1e4,1e5,1e6,1e7}`；`plot_param_sensitivity()` 绘图 | `paper_evidence.parameter_sensitivity` -> `parameter_sensitivity` | `visualization/正确图片输出/fig_param_sensitivity_v2.png` |
| 图 10 链上开销对比 | `experiment/test/Benchmarks.t.sol` 触发 Foundry gas report；`DisputeProtocolBenchmarks.sol` 与 `CleVerVerifier.sol` 提供被测函数；`experiment/comparisons/model.go` 将对比路径绑定到具体基准函数和参数 | `foundry_gas_runs`、`comparison_protocols` -> `gas_comparison` | `visualization/正确图片输出/fig_gas_comparison_v2.png` |
| 图 11 超线性质押博弈 | `deriveStakingAnalysis()` 计算 `p`、`beta`、退出轮次、退出收益和坚持收益；`fig_staking_analysis()` 绘图 | `paper_evidence.staking_analysis` -> `staking_analysis` | `visualization/正确图片输出/fig_staking_analysis_v2.png` |
| 图 12 伴随式验证时间 | `deriveTimeline()` 固定 `eta=0.4`、`t_seg=0.02`、`t_slot=0.008` 并计算各方案归一化端到端时间；`fig_timeline()` 绘图 | `paper_evidence.timeline` -> `timeline` | `visualization/正确图片输出/fig_timeline_v3.png` |

链上 Solidity 基准还会测试以下函数级 Gas。

| Foundry 函数 | 测试内容 | 当前 Gas |
| --- | --- | ---: |
| `fibonacci` | Fibonacci 基准任务合约执行 | `16959` |
| `polyChain` | 多项式链式计算任务 | `41088` |
| `sortLarge` | 排序任务 | `94288` |
| `dpLarge` | 动态规划任务 | `264957` |
| `verSeg` | CleVer 最小裁决单元链上重放接口 | `55444` |
| `arbitrumClassicPath` | Arbitrum Classic 提交、挑战、二分定位和单步裁决路径 | `4438574` |
| `truebitPath` | TrueBit solver/verifier 挑战、二分定位和单步 judge 路径 | `3828334` |
| `cartesiDavePath` | Cartesi Dave tournament setup、claim/counterclaim 二分和 referee 路径 | `5956440` |
| `boldPath` | Arbitrum BoLD 多层并行争议边定位和确认路径 | `3560120` |
| `cleverPath` | CleVer 两层定位和有界 VerSeg 裁决路径 | `551974` |

Geth EVM 采样用于确认基准 bytecode 的执行可测性，当前四个样本 Gas 为 Fibonacci `5492`、Poly-Chain `4836`、Sort-Large `7591`、DP-Large `7911`。

## 目录结构

- `第三章基于有状态任务切片的链下计算验证.pdf`：第三章论文正文。
- `experiment/cmd/clever-exp/main.go`：实验采集入口，生成 raw log。
- `experiment/cmd/comparison-exp/main.go`：从 raw log 中的 Foundry gas 单独导出外部争议协议对比实验报告。
- `experiment/comparisons/model.go`：Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD、CleVer 的对比实验配置、路径参数和时间线模型。
- `experiment/src/BenchmarkTasks.sol`：四类链上基准任务。
- `experiment/src/CleVerVerifier.sol`：最小裁决单元 `VerSeg` 的链上重放裁决接口。
- `experiment/src/DisputeProtocolBenchmarks.sol`：五种协议乐观路径和争议路径 Gas 对比基准。
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

如果从 `chapter3` 目录安装 Python 依赖：

```bash
python3 -m pip install -r requirements.txt
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
go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_protocols.json

cd ..
python3 visualization/generate_chapter3_data.py
python3 visualization/chapter3_all_figures.py
```

每条命令的作用：

- `go test ./...`：检查 Go 实验入口和任务执行逻辑是否能编译、运行。
- `forge test --gas-report`：测量 Solidity 基准任务、VerSeg 和五种争议协议路径的函数级 Gas。
- `go run ./cmd/clever-exp --quick ...`：执行本地确定性任务样本、Geth EVM 采样、Foundry Gas 解析和论文尺度 instrumentation trace 生成。
- `go run ./cmd/comparison-exp ...`：从 raw log 中的 Foundry gas 单独导出 Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD、CleVer 的对比实验报告，便于演示。
- `generate_chapter3_data.py`：校验 raw log 中的 `samples`、`geth_evm_samples`、`comparison_protocols`、`paper_evidence`，写出结构化 JSON。
- `chapter3_all_figures.py`：读取结构化 JSON 并生成图 7-12。

如需运行更大的 runner 样本：

```bash
cd experiment
go run ./cmd/clever-exp --full --max-seconds 5 --out logs/raw_experiment_log.json
```

`--full` 仍受 `--max-seconds` 限制；长任务尺度结果由 instrumentation trace 按参数展开，runner 样本负责验证执行、快照和 commitment 链路。

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
- 外部协议对比报告：`experiment/logs/comparison_protocols.json`
- 结构化数据：`visualization/chapter3_experiment_data.json`
- 生成图片：
  - `visualization/正确图片输出/fig1_budget_compliance.png`
  - `visualization/正确图片输出/fig2_overhead.png`
  - `visualization/正确图片输出/fig_param_sensitivity_v2.png`
  - `visualization/正确图片输出/fig_gas_comparison_v2.png`
  - `visualization/正确图片输出/fig_staking_analysis_v2.png`
  - `visualization/正确图片输出/fig_timeline_v3.png`

## 快速检查数据

```bash
jq '.samples | length' experiment/logs/raw_experiment_log.json
jq '.geth_evm_samples | length' experiment/logs/raw_experiment_log.json
jq '.comparison_protocols | length' experiment/logs/raw_experiment_log.json
jq '.comparison_protocols[] | {scheme, optimistic:(.optimistic_benchmark.function), dispute:(.dispute_benchmark.function), dispute_gas_k, dispute_flow, reproduce}' experiment/logs/raw_experiment_log.json
jq '.paper_evidence.instrumentation_trace | {mode, budget_runs:(.budget_runs|length), overhead_runs:(.overhead_runs|length), parameter_runs:(.parameter_runs|length)}' experiment/logs/raw_experiment_log.json
jq '.gas_comparison' visualization/chapter3_experiment_data.json
```

期望看到 4 个任务样本、4 个 Geth EVM 样本、5 个协议对比项，以及 16 个 budget traces、4 个 overhead traces、8 个 parameter traces。

## 注意事项

- `evm --bench run` 与 `forge test --gas-report` 的执行时间可能随机器环境波动，Gas 数值通常稳定。
- `evm` 缺失会导致 raw log 中 Geth EVM 样本无法通过校验。
- 重新生成 raw log 后，应重新运行数据生成与绘图脚本。

## 验收流程

1. 运行 `go test ./...`，确认 Go 任务 runner、对比模型和时间线推导能通过测试。
2. 运行 `forge test --gas-report`，确认 `BenchmarkTasks`、`CleVerVerifier` 和 `DisputeProtocolBenchmarks` 的函数级 Gas 可重新测量。
3. 运行 `go run ./cmd/clever-exp --quick --out logs/raw_experiment_log.json`，重新生成任务样本、Geth EVM 样本、Foundry Gas 解析、图 7-12 所需 trace。
4. 运行 `go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_protocols.json`，单独导出图 10/表 4 的对比报告。
5. 回到 `chapter3` 运行 `python3 visualization/generate_chapter3_data.py && python3 visualization/chapter3_all_figures.py`，生成结构化数据和图片。
6. 回到仓库根目录运行 `python3 scripts/check_experiment_coverage.py`，检查第三章结果是否与 raw log、结构化数据、图片一致。
