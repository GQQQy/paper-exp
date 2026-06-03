# 第五章实验复现说明

本目录对应论文第五章“面向链下计算的验证者工作审计”。实验覆盖论文 5.5 节、图 24-31 与表 9-11，工程链路为：运行 RanCk/SenCk Go 实验入口生成 raw log，再从 raw log 生成结构化数据和 PNG 图片。

## 实验内容

- RanCk 在线跟踪审计：生成验证者私有 `seed/nonce`，提交 `TC_0`，按 `alpha_t = H(alpha_{t-1} || bh_t || PRF(seed_i,t) || tau_t)` 维护凭证链，使用 `Trigger(t,i)=H(bh_t||tid||i) mod M` 触发心跳，并对端点凭证与抽样点执行连续性审计。
- SenCk 语义勤勉检测：在本地 instrumented EVM-style opcode interpreter 中通过 `AfterOpcodeHook` 采集 `pc/op/rw_t/val_t`、栈、memory/storage 访问，计算 `Gamma(r,tid,k,t,rw_t)`，生成哨兵事件、段摘要 `dig_k` 和 `AudRoot`，再从快照段局部重放提交 `SentReport`。
- 偏离行为：覆盖诚实在线、完全离线、间歇在线、补算失败、凭证链不一致、惰性猜摘要、执行方污染 `ComAud`。
- 激励与检测分析：按论文 5.4-5.5 的 C1/C2 条件、RanCk/SenCk 检测公式和固定随机种子的 Monte Carlo 仿真生成图 24-28。
- 旁路审计开销：记录 opcode hook 的 `rw_t` 编码、哈希、Gamma 判断、快照加载和局部重放时间模型，生成图 29。
- 链上开销：用 Foundry 测试 `TrackInit`、`HBRespond`、`ContAudit`、`SentReport`、`Dispute`、`SentProve` 和 PoD baseline；raw log 和图 30-31 使用本地 Foundry 函数级 Gas report，不使用论文表 10 目标值作为回退数据。

## 实现范围

RanCk 的 Go 实现位于 `experiment/audit/protocol.go`，包含初始化承诺、凭证链更新、心跳触发、心跳响应模拟、端点一致性检查和抽样单步凭证审计。Solidity 合约 `experiment/src/ValidatorAudit.sol` 提供对应的最小链上 benchmark 接口。

SenCk 的 Go 实现位于 `experiment/audit/protocol.go`，使用本地 opcode interpreter 模拟 EVM 执行层的后置 hook，记录运行时访问并生成 `rw_t/val_t`、哨兵事件、段摘要和补证重放结果。该实现用于复现实验与图表数据，不是完整 Geth 客户端补丁；论文中“注入 EVM 解释器轻量回调钩子”的工程对象在这里以可复现的 instrumented local EVM 形式落地。

## 目录结构

- `第五章面向链下计算的验证者工作审计.pdf`：第五章论文正文。
- `experiment/audit/protocol.go`：RanCk/SenCk 协议实现、EVM opcode hook、Monte Carlo、旁路开销和 Gas 数据结构。
- `experiment/audit/protocol_test.go`：RanCk、SenCk 和完整审计流测试。
- `experiment/cmd/audit-exp/main.go`：实验入口，生成 `experiment/logs/raw_experiment_log.json`。
- `experiment/src/ValidatorAudit.sol`：链上审计 benchmark 合约。
- `experiment/test/ValidatorAudit.t.sol`：Foundry Gas 测试。
- `experiment/logs/raw_experiment_log.json`：实验原始日志。
- `visualization/chapter5_all_figures.py`：从 raw log 生成结构化数据和图 24-31。
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

如果缺少 Python 包：

```bash
python3 -m pip install matplotlib numpy
```

如果 Go 或 Matplotlib 默认缓存目录不可写：

```bash
export GOCACHE="${TMPDIR:-/tmp}/chapter5-gocache"
export MPLCONFIGDIR="${TMPDIR:-/tmp}/mplconfig-ch5"
mkdir -p "$GOCACHE" "$MPLCONFIGDIR"
```

## 运行完整实验

从仓库根目录运行：

```bash
cd chapter5/experiment
go test ./...
forge test --gas-report
go run ./cmd/audit-exp --out logs/raw_experiment_log.json

cd ../..
python3 chapter5/visualization/chapter5_all_figures.py
```

说明：

- `go test ./...` 验证 RanCk/SenCk 协议实现、EVM opcode hook trace、惰性/污染偏离检测和完整审计流。
- `forge test --gas-report` 独立运行链上 Gas benchmark。
- `go run ./cmd/audit-exp ...` 重新生成 raw log，并在可用时解析 Foundry per-test gas。
- `chapter5_all_figures.py` 会读取 raw log；如果 raw log 缺失或缺少当前版本必需字段，会自动运行 Go 实验入口补齐，然后生成结构化数据和图片。

## 只重新生成可视化

如果 `experiment/logs/raw_experiment_log.json` 已经存在：

```bash
python3 chapter5/visualization/chapter5_all_figures.py
```

该脚本仍会校验 raw log 中的协议 trace、检测参数、Monte Carlo/Gamma sweep 和 Gas 数据。若 raw log 版本过旧，会自动重新运行 `go run ./cmd/audit-exp`。

## 输出文件

- Raw log：`chapter5/experiment/logs/raw_experiment_log.json`
- 结构化数据：`chapter5/visualization/chapter5_experiment_data.json`
- 生成图片：
  - `chapter5/visualization/正确图片输出/fig24_feasibility_ab.png`
  - `chapter5/visualization/正确图片输出/fig25_joint_feasibility.png`
  - `chapter5/visualization/正确图片输出/fig26_ranck_detection.png`
  - `chapter5/visualization/正确图片输出/fig27_senck_passthrough.png`
  - `chapter5/visualization/正确图片输出/fig28_monte_carlo_and_gate.png`
  - `chapter5/visualization/正确图片输出/fig29_overhead.png`
  - `chapter5/visualization/正确图片输出/fig30_gas_comparison.png`
  - `chapter5/visualization/正确图片输出/fig31_tradeoff.png`

## 数据链路

```text
experiment/audit/protocol.go
  -> experiment/cmd/audit-exp/main.go
  -> experiment/logs/raw_experiment_log.json
  -> visualization/chapter5_all_figures.py
  -> visualization/chapter5_experiment_data.json
  -> visualization/正确图片输出/*.png
```

raw log 的主要来源：

- `ranck_traces`：Go 实现的 `TrackInit`、`UpdateAlpha`、`Trigger`、`SimulateRanCk` 和 `ContAudit`。
- `senck_traces`：Go 实现的 instrumented local EVM、`AfterOpcodeHook`、`Gamma`、`SentinelEvent`、`DigestEvents`、`AudRoot` 和 `SimulateSenCk`。
- `detection_parameters`：图 26/27 使用的参数扫描范围，来自表 9 范围和 5.5 节检测公式。
- `monte_carlo`：RanCk/SenCk Bernoulli 仿真和 Gamma 命中率 sweep，固定随机种子或确定性 EVM trace。
- `feasibility`：C1/C2 公式扫描和联合可行域。
- `overhead_traces`：opcode hook 开销、快照加载和局部重放模型。
- `gas_trace`：Foundry 函数级 Gas report、本地解析状态和测量来源说明。

快速检查数据：

```bash
jq '.protocol_coverage[] | {requirement, source, status}' chapter5/experiment/logs/raw_experiment_log.json
jq '.senck_traces[0].trace_sample[0] | {op_name, hook_source, runtime_accesses}' chapter5/experiment/logs/raw_experiment_log.json
jq '.monte_carlo.gamma_hit_sweep[] | {M_c, L, theory_rho, observed_rho, segments}' chapter5/experiment/logs/raw_experiment_log.json
jq '.gas_trace | {foundry_status, foundry_parsed, measurement_provenance}' chapter5/experiment/logs/raw_experiment_log.json
```

## 论文图表对应关系

| 论文图表 | 输出图片/数据 | 数据字段 | 生成/测量代码 |
| --- | --- | --- | --- |
| 表 9 参数配置 | `chapter5_experiment_data.json` | `table9_parameters`、`metadata.params` | `audit.DefaultParams`、`audit.Table9Parameters` |
| 图 24 归一化边界 | `fig24_feasibility_ab.png` | `feasibility.C1_online_deviation`、`C2_diligence_deviation` | `feasibilityEvidence` |
| 图 25 联合可行域 | `fig25_joint_feasibility.png` | `feasibility.joint_feasible_region` | `feasibilityEvidence` |
| 图 26 RanCk 检测 | `fig26_ranck_detection.png` | `detection.ranck_heartbeat`、`ranck_combined_pass` | `detection_parameters`、`ranck_detect`、`ranck_combined_pass` |
| 图 27 SenCk 检测 | `fig27_senck_passthrough.png` | `detection.senck_lazy_pass` | `detection_parameters`、`senck_pass` |
| 图 28 效果模拟仿真 | `fig28_monte_carlo_and_gate.png` | `monte_carlo.ranck`、`senck`、`gamma_hit_sweep`、`senck_traces[].trigger_stats` | `MonteCarloRanCk`、`MonteCarloSenCk`、`GammaHitSweep`、`TriggerStats` |
| 图 29 旁路审计开销 | `fig29_overhead.png` | `overhead` | `GenerateSegment`、`OverheadTraces` |
| 表 10 链上开销 | `gas_trace.operations` | `gas.operations`、`gas.measurement_provenance`、`gas.foundry_parsed` | `ValidatorAudit.sol`、`ValidatorAudit.t.sol`、`GasTraceFromFoundry` |
| 图 30 Gas 对比 | `fig30_gas_comparison.png` | `gas.operations`、`gas.monthly_by_scheme` | `GasTraceFromFoundry`、`fig30` |
| 图 31 心跳权衡 | `fig31_tradeoff.png` | `gas.pi_h_sweep` | `GasTraceFromFoundry`、`fig31` |
| 表 11 综合对比 | `comparison_table_11` | `comparison_table_11` | `table11` |

## 注意事项

- `out/`、`cache/`、`__pycache__/`、生成 PDF 和其他可重建产物不提交。
- Foundry per-test gas 可能包含测试 harness/setup 效应；raw log 会保存本地实测值和测量来源，图 30-31 使用本地 Foundry 实测值。
- 重新生成 raw log 后，应重新运行 `chapter5_all_figures.py`，使结构化数据和图片保持一致。
