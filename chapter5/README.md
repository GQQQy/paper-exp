# 第五章实验复现说明

本目录对应论文第五章“面向链下计算的验证者工作审计”。实验覆盖论文 5.5 节、图 24-31 与表 9-11，目标是复现 RanCk 在线跟踪审计、SenCk 语义勤勉检测、激励可行域、检测概率、旁路审计开销和链上 Gas 对比结果。

实验链路为：

```text
experiment/audit/protocol.go
  -> experiment/cmd/audit-exp/main.go
  -> experiment/logs/raw_experiment_log.json
  -> visualization/chapter5_all_figures.py
  -> visualization/chapter5_experiment_data.json
  -> visualization/正确图片输出/*.png
```

表 10、图 30 和图 31 的 Gas 数据由 Foundry gas report 生成，并在 raw log 的 `measurement_provenance.measurement_status=checked` 中记录函数级 Gas、PoD 月度派生公式、PoD watchtower trace 和 `podMineBounty` benchmark。根目录的 `scripts/check_experiment_coverage.py` 会检查这些结果是否由可运行实验产物支撑。

第五章的关键数值由四类实验产物汇总形成：

- RanCk/SenCk 行为结果来自 `audit.SimulateRanCk()`、`audit.SimulateSenCk()` 和 `protocol_test.go` 覆盖的协议 trace，汇总到 `ranck_traces`、`senck_traces` 和 `protocol_coverage`。
- 图 24-28 的可行域、检测概率和 Monte Carlo 对照来自公式 trace、固定随机种子 Monte Carlo trace 与 Gamma 命中率 sweep，汇总到 `feasibility`、`detection` 和 `monte_carlo`。
- 图 29 的旁路开销来自 instrumented EVM opcode hook、rw 编码、哈希、Gamma、快照加载和局部重放 trace，汇总到 `overhead_traces`。
- 表 10、图 30-31 和 PoD baseline 来自 `ValidatorAudit.t.sol` 的 Foundry gas report、`podMineBounty` benchmark 与 deterministic watchtower epoch trace，汇总到 `gas_trace`。

结构化 JSON 与输出图片均由上述 raw experiment log 校验后生成；表 9-11 的字段也保留在 `chapter5_experiment_data.json` 中，便于和论文正文逐项对照。

## 实验总览

默认参数如下。

| 参数 | 值 | 用途 |
| --- | --- | --- |
| 区块时间 | `12 s` | 心跳窗口和检测延迟的时间基准 |
| 窗口长度 `T_win` | `7200` blocks | 每个审计窗口长度 |
| 验证者规模 `N` | `20` | 图 30/31 的系统级 Gas 对比规模 |
| 段预算 `B` | `1e8 Gas` | 沿用第三章切片段预算 |
| 裁决阈值 `b` | `1e6 Gas` | 沿用第三章 VerSeg 裁决阈值 |
| RanCk 心跳间隔 `M` | `100` | 默认心跳触发概率 `pi_h=1/M=0.01` |
| 响应时限 `Delta` | `7` blocks | 心跳响应允许延迟 |
| 连续性抽样规模 `s` | `10` | RanCk 凭证连续性抽样点数 |
| SenCk 哨兵周期 `M_c` | `100` | 单步门控函数触发周期 |
| 重放长度 `L` | `5000` steps | 每个抽样段局部重放步数 |
| 抽样段数 `m_s` | `10` | SenCk 勤勉检测抽样段数 |

## 测试内容与结果

| 测试/图表 | 测试对象 | 数据来源 | 当前关键结果 | 输出 |
| --- | --- | --- | --- | --- |
| 表 9 参数配置 | 审计窗口、验证者规模、RanCk/SenCk 参数范围和默认值 | `table9_parameters`、`metadata.params` | 默认值为 `T_win=7200,N=20,B=1e8,b=1e6,M=100,Delta=7,s=10,M_c=100,L=5000,m_s=10` | `visualization/chapter5_experiment_data.json` |
| 图 24 C1/C2 归一化边界 | 在线偏离抑制 C1 与勤勉偏离抑制 C2 的罚没条件 | `feasibility.C1_online_deviation`、`C2_diligence_deviation` | `T_win=7200, pi_h≈0.01005` 时期望心跳 `72.36` 次，C1 最小归一化罚没约 `0.0138` | `visualization/正确图片输出/fig24_feasibility_ab.png` |
| 图 25 联合可行域 | 在 `(pi_h,m_s)` 平面上同时满足 C1/C2 的区域 | `feasibility.joint_feasible_region` | 默认罚没参数下，大部分扫描区域同时满足 C1 与 C2 | `visualization/正确图片输出/fig25_joint_feasibility.png` |
| 图 26 RanCk 检测 | 心跳缺失检测和心跳+连续性审计联合通过概率 | `detection.ranck_heartbeat`、`detection.ranck_combined_pass` | `M=100` 时离线 `500` blocks 检出概率 `0.993430`；`s=10`、离线 `200` blocks 联合通过概率 `8.101e-04` | `visualization/正确图片输出/fig26_ranck_detection.png` |
| 图 27 SenCk 检测 | 惰性验证者通过勤勉检测的概率随抽样段数衰减 | `detection.senck_lazy_pass` | `rho=0.4` 时 `m_s=10` 通过概率 `6.047e-03`，`m_s=20` 通过概率 `3.656e-05` | `visualization/正确图片输出/fig27_senck_passthrough.png` |
| 图 28 Monte Carlo 验证 | RanCk/SenCk 理论检测公式与固定随机种子仿真的吻合程度 | `monte_carlo.ranck`、`monte_carlo.senck`、`gamma_hit_sweep` | RanCk 最大理论/仿真偏差约 `0.00708`；SenCk 最大偏差约 `0.00418`，均在统计波动范围内 | `visualization/正确图片输出/fig28_monte_carlo_and_gate.png` |
| 图 29 旁路审计开销 | opcode hook 的读写编码、哈希、Gamma 判断、快照加载和局部重放时间 | `overhead` | `L=5000` 时 Fibonacci/Poly-Chain/Sort-Large/DP-Large 的审计开销分别约 `1.869%/1.406%/1.192%/0.808%` | `visualization/正确图片输出/fig29_overhead.png` |
| 表 10 链上开销 | RanCk/SenCk 正常路径、争议路径和 PoD MineBounty 的函数级 Gas | `gas.operations`、`gas.pod_baseline`、`gas.measurement_provenance` | 默认每验证者每窗口正常路径总 Gas 为 `2.698M`；`SentProve/VerSeg` 为 `1.269M`；PoD MineBounty 单 epoch 为 `382010 Gas` | `visualization/chapter5_experiment_data.json` |
| 图 30 Gas 对比 | 单验证者正常路径 Gas 构成和系统月度 Gas 对比 | `gas.operations`、`gas.monthly_by_scheme`、`gas.pod_baseline` | 本方案 `pi_h=0.01` 月度 `1.619B Gas`，`pi_h=0.005` 月度 `0.868B Gas`，PoD MineBounty `29.705B Gas` | `visualization/正确图片输出/fig30_gas_comparison.png` |
| 图 31 心跳权衡 | 心跳触发概率对月度 Gas 和 `ell=300` 检测概率的影响 | `gas.pi_h_sweep` | `pi_h` 从 `0.001` 到 `0.025` 时月度 Gas 低于当前 PoD MineBounty 基线，检测概率随 `pi_h` 单调上升 | `visualization/正确图片输出/fig31_tradeoff.png` |
| 表 11 综合对比 | TrueBit、Arbitrum、PoD、RanCk+SenCk 的审计能力对比 | `comparison_experiments` -> `comparison_table_11` | RanCk+SenCk 同时覆盖在线状态审计、语义勤勉检测和不可预测审计，不依赖外部网络辅助 | `visualization/chapter5_experiment_data.json` |

## 对比复现边界

第五章表 11 是由本地场景模型、PoD watchtower trace 和 RanCk/SenCk 协议 trace 共同生成的能力对比。TrueBit 和 Arbitrum 行由本仓库的本地场景模型给出在线审计、语义勤勉、不可预测审计和外部网络依赖四类结果；PoD 行额外跑 deterministic watchtower epoch trace，并用 `ValidatorAudit.podMineBounty` 的 Foundry Gas 生成图 30/31 中的月度基线；RanCk+SenCk 行来自本章 Go 协议 trace 和 Solidity benchmark。

| 方案 | 本地复现实验 | 验收查看字段 |
| --- | --- | --- |
| TrueBit | `comparisons.RunCapabilityExperiments()` 中的 solver/verifier 场景 | `comparison_experiments[].scenarios`、`protocol_flow` |
| Arbitrum | `comparisons.RunCapabilityExperiments()` 中的 assertion/challenge 场景 | `comparison_experiments[].scenarios`、`protocol_flow` |
| PoD | `SimulatePoDExperiment()` watchtower epoch trace + `podMineBounty` Foundry benchmark | `gas_trace.pod_experiment_trace`、`gas_trace.pod_baseline` |
| RanCk+SenCk | `SimulateRanCk()`、`SimulateSenCk()`、`ValidatorAudit.sol` Gas benchmark | `ranck_traces`、`senck_traces`、`gas_trace.operations` |

## 代码与论文结果定位

论文第五章 5.5 节的表 9-11 和图 24-31 可按下表定位到具体代码。整体流程是先由 `experiment/cmd/audit-exp/main.go` 调用 `experiment/audit/protocol.go` 生成 raw log，再由 `visualization/chapter5_all_figures.py` 校验 raw log、导出结构化 JSON 并渲染图片。

| 论文结果 | 负责生成或测量的代码 | JSON 数据字段 | 最终输出 |
| --- | --- | --- | --- |
| 表 9 参数配置 | `audit.DefaultParams()` 和 `audit.Table9Parameters()` 定义默认值与参数范围；`cmd/audit-exp/main.go` 写入 raw log | `params`、`table9_parameters`、`metadata.params` | README 默认参数表；`visualization/chapter5_experiment_data.json` |
| RanCk 行为 trace | `audit.SimulateRanCk()`、`TrackInit()`、`UpdateAlpha()`、`Trigger()`、`ContAudit()` 生成诚实、离线、间歇在线、补算失败和凭证不一致行为；`protocol_test.go` 覆盖行为测试 | `ranck_traces` | README “RanCk 协议行为测试”；`experiment/logs/raw_experiment_log.json` |
| SenCk 行为 trace | `audit.SimulateSenCk()`、`GenerateSegment()`、`InstrumentedEVM.ExecuteStep()`、`AfterOpcodeHook()`、`Gamma()`、`SentinelEvent()`、`DigestEvents()` 生成哨兵摘要和局部重放 trace；`protocol_test.go` 覆盖惰性/污染检测 | `senck_traces`、`protocol_coverage` | README “SenCk 协议行为测试”；`experiment/logs/raw_experiment_log.json` |
| 图 24 C1/C2 边界 | `feasibilityEvidence()` 扫描 C1/C2 罚没边界；`fig24()` 绘制归一化边界 | `feasibility.C1_online_deviation`、`feasibility.C2_diligence_deviation` | `visualization/正确图片输出/fig24_feasibility_ab.png` |
| 图 25 联合可行域 | `feasibilityEvidence()` 生成 `(pi_h,m_s)` 网格；`fig25()` 绘制 C1/C2 联合可行域 | `feasibility.joint_feasible_region` | `visualization/正确图片输出/fig25_joint_feasibility.png` |
| 图 26 RanCk 检测概率 | `detectionParameters()` 给出扫描范围；`ranck_detect()` 和 `ranck_combined_pass()` 由公式生成心跳检出与联合通过概率；`fig26()` 出图 | `detection_parameters`、`detection.ranck_heartbeat`、`detection.ranck_combined_pass` | `visualization/正确图片输出/fig26_ranck_detection.png` |
| 图 27 SenCk 检测概率 | `senck_pass()` 计算惰性策略通过概率 `(1-rho)^m_s`；`fig27()` 出图 | `detection.senck_lazy_pass` | `visualization/正确图片输出/fig27_senck_passthrough.png` |
| 图 28 Monte Carlo 与 Gamma | `monteCarloEvidence()` 调用 `audit.MonteCarloRanCk()`、`audit.MonteCarloSenCk()`；`audit.GammaHitSweep()` 生成门控命中率；`fig28()` 出图 | `monte_carlo.ranck`、`monte_carlo.senck`、`monte_carlo.gamma_hit_sweep`、`senck_traces[0].trigger_stats` | `visualization/正确图片输出/fig28_monte_carlo_and_gate.png` |
| 图 29 旁路审计开销 | `audit.Workloads()` 给出四类任务模型；`audit.OverheadTraces()` 生成 opcode hook、哈希、Gamma、快照加载和重放时间；`fig29()` 出图 | `overhead_traces` -> `overhead` | `visualization/正确图片输出/fig29_overhead.png` |
| 表 10 链上开销 | `experiment/test/ValidatorAudit.t.sol` 触发链上审计 Gas 测试；`cmd/audit-exp/main.go` 的 `runFoundryGas()` 解析 gas-report 函数平均值；`audit.GasTraceFromFoundry()` 生成默认窗口、月度和 sweep 数据；`experiment/comparisons/model.go` 生成 PoD MineBounty 基线 | `gas_trace.operations`、`gas_trace.pod_baseline`、`gas_trace.measurement_provenance` -> `gas.operations` | README “链上 Gas 测试结果”；`visualization/chapter5_experiment_data.json` |
| 图 30 Gas 对比 | `audit.GasTraceFromFoundry()` 计算正常路径构成和月度对比；PoD MineBounty 基线由 `comparisons.BuildPoDBaseline()` 使用 Foundry `testPoDMineBountyGas`/`podMineBounty` 生成；`fig30()` 出图 | `gas.operations`、`gas.monthly_by_scheme`、`gas.pod_baseline` | `visualization/正确图片输出/fig30_gas_comparison.png` |
| 图 31 心跳权衡 | `audit.GasTraceFromFoundry()` 生成 `pi_h_sweep`；`fig31()` 同时绘制月度 Gas 与 `ell=300` 检测概率 | `gas.pi_h_sweep`、`gas.monthly_by_scheme` | `visualization/正确图片输出/fig31_tradeoff.png` |
| 表 11 综合对比 | `experiment/comparisons/model.go` 用在线审计、语义勤勉、不可预测审计和外部网络依赖四类场景生成 TrueBit、Arbitrum、PoD、RanCk+SenCk 对比项；`cmd/comparison-exp` 可单独导出报告 | `comparison_experiments` -> `comparison_table_11` | `visualization/chapter5_experiment_data.json`、`experiment/logs/comparison_experiments.json` |

## 实验链路校验

当前工程通过 Go 协议实现、Solidity benchmark 和 Python 可视化脚本重建第五章表 9-11 与图 24-31。`SentProve/VerSeg` 行叠加第三章 `b=1e6` 的 VerSeg 重放模型 `1.2*b`，因为该行表示复用裁决接口的完整争议路径，而不只是 `sentProve` 包装函数。覆盖检查脚本会检查 RanCk/SenCk trace、Foundry 函数级 Gas、PoD MineBounty benchmark、PoD watchtower trace 和图表结构化数据是否互相对应。

## RanCk 协议行为测试

`go test ./...` 和 `go run ./cmd/audit-exp` 会生成 `ranck_traces`，覆盖诚实和四类偏离行为。

| 行为 | 测试内容 | 当前心跳数 | 漏响应数 | 连续性审计结果 |
| --- | --- | ---: | ---: | --- |
| `honest_online` | 正常在线，按 `alpha_t=H(alpha_{t-1}||bh_t||PRF(seed,t)||tau_t)` 更新凭证链并响应心跳 | `68` | `0` | `passed=true, detected=false` |
| `complete_offline` | 完全离线，不响应任何心跳 | `68` | `68` | `passed=false, detected=true` |
| `intermittent_online` | 间歇在线，部分心跳响应缺失 | `68` | `36` | `passed=false, detected=true` |
| `recompute_fail` | 试图补算但凭证链或抽样点无法一致 | `68` | `48` | `passed=false, detected=true` |
| `credential_inconsistent` | 心跳存在但端点凭证/抽样凭证不一致 | `68` | `0` | `passed=false, detected=true` |

这些 trace 对应论文中的 RanCk 在线跟踪和连续性审计，重点验证心跳触发、响应时限、端点凭证和抽样单步凭证是否能区分诚实与偏离行为。

## SenCk 协议行为测试

`senck_traces` 使用 instrumented EVM-style opcode interpreter，在每个 opcode 后通过 `AfterOpcodeHook` 记录 `pc/op/rw_t/val_t`、栈、memory/storage 访问，并以 `Gamma(r,tid,k,t,rw_t)` 触发哨兵事件。

| 行为 | 测试内容 | 单段命中率 `rho` | SentReport 结果 |
| --- | --- | ---: | --- |
| `honest_online` | 正常采集 opcode hook、生成哨兵事件、计算 `dig_k`，并从快照段局部重放 | `1.0` | `passed=true, detected=false` |
| `lazy_guess` | 惰性验证者不重放，猜测段摘要 | `1.0` | `passed=false, detected=true` |
| `executor_polluted_comaud` | 执行方污染 `ComAud` 或提交错误审计摘要 | `1.0` | `passed=false, detected=true` |

当前 `protocol_coverage` 中 10 项协议需求均为 `PASS`，包括 TrackInit、alpha 链更新、Trigger、HBRespond、ContAudit、opcode hook、Gamma、SentReport、SentDispute/SentProve 和完整流程。

## 旁路审计开销细节

图 29 使用四类任务的 `overhead` trace。

| 任务 | 论文尺度 Gas | `L=5000` 哨兵数量 | `L=5000` 重放耗时 | 审计开销 |
| --- | ---: | ---: | ---: | ---: |
| Fibonacci | `1e9` | `50` | `222.00 ms` | `1.869%` |
| Poly-Chain | `1e10` | `55` | `366.85 ms` | `1.406%` |
| Sort-Large | `1e11` | `61` | `568.15 ms` | `1.192%` |
| DP-Large | `1e12` | `55` | `768.50 ms` | `0.808%` |

`replay_by_length` 还记录 `L=50,100,200,500,1000,2000,5000,10000` 时的快照加载和局部重放耗时，用于展示重放长度对审计成本的影响。

## 链上 Gas 测试结果

`forge test --gas-report` 覆盖正常路径、争议路径和 PoD MineBounty。

| 操作 | 测试内容 | 单次 Gas | 默认次数/窗口 |
| --- | --- | ---: | ---: |
| `RanCk TrackInit` | 验证者初始化 `TC_0`，写入跟踪状态 | `49359` | `1` |
| `RanCk HBRespond` | 响应心跳并更新最新凭证 commitment | `34781` | `72` |
| `RanCk ContAudit` | 验证端点凭证和抽样单步凭证 | `89918` | `1` |
| `SenCk SentReport` | 提交哨兵段摘要报告 | `54915` | `1` |
| `Dispute` | 对错误哨兵报告发起争议 | `79892` | `0` |
| `SentProve/VerSeg` | `sentProve` 提交开销加第三章 `b=1e6` 的 VerSeg 重放模型 | `1269429` | `0` |
| `PoD MineBounty` | watchtower 提交 state root、trace root、VRF proof 和 bounty receipt | `382010` | 按 epoch |

正常路径默认总量计算为：

```text
49359 + 72 * 34781 + 89918 + 54915 = 2698424 Gas
```

结构化 JSON 中记录的精确值为 `2698424`。图 30/31 的 PoD MineBounty 基线使用 `podMineBounty` Gas `382010`、`theta=0.9`、月度 `4320` 个 epoch 和 `N=20` 验证者规模计算，得到 `29705097600 Gas`。

## 目录结构

- `第五章面向链下计算的验证者工作审计.pdf`：第五章论文正文。
- `experiment/audit/protocol.go`：RanCk/SenCk 协议实现、EVM opcode hook、Monte Carlo、旁路开销和 Gas 数据结构。
- `experiment/audit/protocol_test.go`：RanCk、SenCk 和完整审计流测试。
- `experiment/cmd/audit-exp/main.go`：实验入口，生成 `experiment/logs/raw_experiment_log.json`。
- `experiment/cmd/comparison-exp/main.go`：单独导出 TrueBit、Arbitrum、PoD、RanCk+SenCk 能力对比和 PoD MineBounty 报告。
- `experiment/comparisons/model.go`：表 11 对比场景模型和 PoD MineBounty gas 基线计算。
- `experiment/src/ValidatorAudit.sol`：链上审计 benchmark 合约。
- `experiment/test/ValidatorAudit.t.sol`：Foundry Gas 测试。
- `experiment/logs/raw_experiment_log.json`：实验原始日志。
- `visualization/chapter5_all_figures.py`：校验 raw log 并生成结构化数据和图 24-31。
- `visualization/chapter5_experiment_data.json`：结构化实验数据。
- `visualization/正确图片输出/`：图 24-31 的 PNG 输出。

## 环境依赖

需要以下工具：

- Go 1.21 或更高版本。
- Foundry，包括 `forge`。
- Python 3.10 或兼容版本。
- Python 包：`matplotlib`、`numpy`。

检查命令：

```bash
go version
forge --version
python3 --version
python3 -c "import matplotlib, numpy; print('python deps ok')"
```

安装 Python 依赖：

```bash
python3 -m pip install matplotlib numpy
```

## 运行完整实验

从 `chapter5` 目录运行：

```bash
cd experiment
go test ./...
forge test --gas-report
go run ./cmd/audit-exp --out logs/raw_experiment_log.json
go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_experiments.json

cd ..
python3 visualization/chapter5_all_figures.py
```

每条命令的作用：

- `go test ./...`：验证 RanCk/SenCk 协议实现、EVM opcode hook trace、惰性/污染偏离检测和完整审计流。
- `forge test --gas-report`：测量链上审计合约的函数级 Gas。
- `go run ./cmd/audit-exp ...`：生成 RanCk/SenCk trace、Monte Carlo、可行域、旁路开销、Gas trace 和表 11 对比数据。
- `go run ./cmd/comparison-exp ...`：从 raw log 中的 Foundry gas 单独导出表 11 能力对比和 PoD MineBounty 报告，便于演示。
- `chapter5_all_figures.py`：校验协议 trace、检测参数、Monte Carlo/Gamma sweep 和 Gas 数据，生成结构化 JSON 与图 24-31。

命令与产物对应关系：

| 命令 | 实验内容 | 生成或刷新产物 | 支撑论文结果 |
| --- | --- | --- | --- |
| `go test ./...` | RanCk/SenCk 协议行为、PoD trace、能力对比和 PoD baseline 单元/集成测试 | 测试输出 | 表 9-11、图 24-31 的协议实现可信度 |
| `forge test --gas-report` | TrackInit、HBRespond、ContAudit、SentReport、Dispute、SentProve、PoD MineBounty 函数级 Gas | Foundry gas report，被 `audit-exp` 汇总为 `gas_trace.foundry_parsed` | 表 10、图 30、图 31 |
| `go run ./cmd/audit-exp --out logs/raw_experiment_log.json` | RanCk/SenCk trace、Monte Carlo、可行域、旁路开销、Gas trace、表 11 场景对比 | `logs/raw_experiment_log.json` | 表 9-11、图 24-31 |
| `go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_experiments.json` | TrueBit、Arbitrum、PoD、RanCk+SenCk 能力对比和 PoD MineBounty baseline standalone 报告 | `logs/comparison_experiments.json` | 表 11、图 30/31 PoD 对照 |
| `python3 visualization/chapter5_all_figures.py` | raw log 校验、结构化实验产物生成、图片渲染 | `visualization/chapter5_experiment_data.json`、`visualization/正确图片输出/*.png` | 图 24-31、表 9-11 |

## 只重新生成可视化

如果 `experiment/logs/raw_experiment_log.json` 已经存在：

```bash
python3 visualization/chapter5_all_figures.py
```

该脚本仍会校验 raw log 中的协议 trace、检测参数、Monte Carlo/Gamma sweep 和 Gas 数据。若 raw log 版本过旧，会自动重新运行 `go run ./cmd/audit-exp`。

## 输出文件

- Raw log：`experiment/logs/raw_experiment_log.json`
- 对比实验报告：`experiment/logs/comparison_experiments.json`
- 结构化数据：`visualization/chapter5_experiment_data.json`
- 生成图片：
  - `visualization/正确图片输出/fig24_feasibility_ab.png`
  - `visualization/正确图片输出/fig25_joint_feasibility.png`
  - `visualization/正确图片输出/fig26_ranck_detection.png`
  - `visualization/正确图片输出/fig27_senck_passthrough.png`
  - `visualization/正确图片输出/fig28_monte_carlo_and_gate.png`
  - `visualization/正确图片输出/fig29_overhead.png`
  - `visualization/正确图片输出/fig30_gas_comparison.png`
  - `visualization/正确图片输出/fig31_tradeoff.png`

## 图表产物映射

| 结构化字段 | 实验依据 | 输出图片/结果 |
| --- | --- | --- |
| `table9_parameters`、`metadata.params` | `audit.DefaultParams()`、`audit.Table9Parameters()` | 表 9 参数配置 |
| `feasibility` | C1/C2 罚没边界和联合可行域公式 trace | 图 24、图 25 |
| `detection` | RanCk 心跳/连续性检测公式、SenCk lazy pass 公式 | 图 26、图 27 |
| `monte_carlo`、`senck_traces[0].trigger_stats` | 固定随机种子 Monte Carlo trace 与 Gamma hit sweep | 图 28 |
| `overhead` | instrumented EVM opcode hook、rw 编码、哈希、Gamma、快照加载、局部重放 trace | 图 29 |
| `gas.operations`、`gas.pod_baseline`、`gas.monthly_by_scheme` | Foundry gas report、PoD watchtower trace、月度 Gas 公式 | 表 10、图 30 |
| `gas.pi_h_sweep` | 心跳概率 sweep 与 `ell=300` 检测概率 | 图 31 |
| `comparison_experiments`、`comparison_table_11` | 本地场景模型、PoD epoch trace、RanCk/SenCk 协议 trace | 表 11 |

## 快速检查数据

```bash
jq '.protocol_coverage[] | {requirement, source, status}' experiment/logs/raw_experiment_log.json
jq '.ranck_traces[] | {behavior, heartbeat_count, missed_heartbeats, cont_audit}' experiment/logs/raw_experiment_log.json
jq '.senck_traces[] | {behavior, rho, sent_report}' experiment/logs/raw_experiment_log.json
jq '.monte_carlo.gamma_hit_sweep[] | {M_c, L, theory_rho, observed_rho, segments}' experiment/logs/raw_experiment_log.json
jq '.gas_trace | {foundry_status, operations, monthly_by_scheme, measurement_provenance}' experiment/logs/raw_experiment_log.json
jq '.gas_trace.pod_baseline' experiment/logs/raw_experiment_log.json
jq '.comparison_experiments[] | {scheme, online_audit, semantic_diligence, unpredictable_audit, external_network, scenarios:(.scenarios|length), flow:(.protocol_flow|length)}' experiment/logs/raw_experiment_log.json
```

## 验收流程

1. 运行 `go test ./...`，确认 RanCk/SenCk、PoD trace、能力对比和 PoD baseline 计算通过测试。
2. 运行 `forge test --gas-report`，确认 `ValidatorAudit.sol` 中 TrackInit、HBRespond、ContAudit、SentReport、Dispute、SentProve、PoD MineBounty 的 Gas 可重新测量。
3. 运行 `go run ./cmd/audit-exp --out logs/raw_experiment_log.json`，重新生成 RanCk/SenCk 行为 trace、Monte Carlo、可行域、旁路开销、Gas trace 和表 11 场景对比。
4. 运行 `go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_experiments.json`，单独导出表 11 和 PoD MineBounty 对比报告。
5. 回到 `chapter5` 运行 `python3 visualization/chapter5_all_figures.py`，生成结构化数据和图 24-31。
6. 回到仓库根目录运行 `python3 scripts/check_experiment_coverage.py`，检查第五章结果是否与 raw log、结构化数据、图片一致。

## 注意事项

- `out/`、`cache/`、`__pycache__/`、生成 PDF 和其他可重建产物不提交。
- Foundry per-test gas 可能包含测试 harness/setup 效应；raw log 会保存实测值、测量来源和论文对齐校验字段。
- 重新生成 raw log 后，应重新运行 `chapter5_all_figures.py`，使结构化数据和图片保持一致。
