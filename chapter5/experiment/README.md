# 第五章验证者工作审计实验工程

本目录用于运行第五章“面向链下计算的验证者工作审计”实验，覆盖 RanCk 随机连续性抽查、SenCk 哨兵抽样勤勉检测、链下旁路审计开销、链上 Gas 基准与论文图 24-31/表 9-11 的数据生成。

表 10、图 30 和图 31 的 Gas 数据由 Foundry gas report 生成，并在 raw log 的 `measurement_provenance.measurement_status=checked` 中记录函数级 Gas、PoD 月度派生公式、PoD watchtower trace 和 `podMineBounty` benchmark。根目录的 `scripts/check_experiment_coverage.py` 会检查这些结果是否由可运行实验产物支撑。

## 实验内容

- RanCk：生成验证者私有 seed/nonce、提交 `TC_0`、维护 `alpha_t = H(alpha_{t-1} || bh_t || PRF(seed_i,t) || tau_t)`、按 `Trigger(t,i)=H(bh_t||tid||i) mod M` 触发心跳，并对端点和抽样点执行连续性审计。
- SenCk：在 instrumented EVM-style opcode interpreter 的每个 opcode 执行后，通过 `AfterOpcodeHook` 采集 `pc_t/op_t`、memory/storage/stack 访问，编码运行时 `rw_t` 与 `val_t`，计算 `Gamma(r,tid,k,t,rw_t)`，生成哨兵事件和段摘要 `dig_k`，再从快照段局部重放提交 `SentReport`。
- 行为模型：覆盖诚实在线、完全离线、间歇在线、补算失败、凭证链不一致、惰性猜摘要、执行方污染 `ComAud` 等偏离。
- Foundry：提供最小审计合约和 gas benchmark 测试，覆盖 `TrackInit`、`HBRespond`、`ContAudit`、`SentReport`、`Dispute`、`SentProve` 和 PoD MineBounty。
- 对比实验：`comparisons/model.go` 生成 TrueBit、Arbitrum、PoD、RanCk+SenCk 的能力场景对比，并用 Foundry `testPoDMineBountyGas`/`podMineBounty` 计算 PoD 月度基线。
- 可视化：校验 `logs/raw_experiment_log.json` 中的协议 trace、Monte Carlo、PoD baseline 和 Foundry Gas，生成结构化数据和图 24-31。

关键结果均由可复现实验产物形成：

- RanCk/SenCk 协议行为来自 Go trace 与 `protocol_test.go`，对应 `ranck_traces`、`senck_traces`、`protocol_coverage`。
- 图 24-28 来自 C1/C2 公式 trace、检测概率公式、固定随机种子 Monte Carlo trace 和 Gamma hit sweep。
- 图 29 来自 instrumented EVM opcode hook、rw 编码、哈希、Gamma、快照加载与局部重放 trace。
- 表 10、图 30-31 来自 Foundry gas report、`podMineBounty` benchmark、PoD watchtower epoch trace 和月度 Gas 派生公式。
- 表 11 来自 `comparison_experiments` 场景模型、PoD trace 与 RanCk/SenCk 协议 trace。

## 对比复现边界

`comparison_experiments` 和 `gas_trace.pod_baseline` 记录的是本地场景模型和 benchmark 的复现边界。TrueBit/Arbitrum 行由本地场景模型生成能力结果，PoD 行由 watchtower epoch trace 和 `podMineBounty` Foundry Gas 生成，RanCk+SenCk 行由本目录的协议 trace 和合约 Gas 生成。

| 方案 | 本地复现实验 | 验收查看字段 |
| --- | --- | --- |
| TrueBit | solver/verifier 场景模型 | `comparison_experiments[].scenarios`、`protocol_flow` |
| Arbitrum | assertion/challenge 场景模型 | `comparison_experiments[].scenarios`、`protocol_flow` |
| PoD | deterministic watchtower epoch trace + `podMineBounty` benchmark | `gas_trace.pod_experiment_trace`、`gas_trace.pod_baseline` |
| RanCk+SenCk | RanCk/SenCk Go trace + `ValidatorAudit.sol` benchmark | `ranck_traces`、`senck_traces`、`gas_trace.operations` |

## 目录说明

- `audit/protocol.go`：RanCk/SenCk 协议实现、instrumented EVM opcode hook、Monte Carlo、旁路开销和 Foundry Gas 数据结构。
- `audit/protocol_test.go`：RanCk、SenCk 和完整审计流单元/集成测试。
- `cmd/audit-exp/main.go`：实验入口，生成 `logs/raw_experiment_log.json`。
- `cmd/comparison-exp/main.go`：单独导出表 11 能力对比和 PoD MineBounty 报告。
- `comparisons/model.go`：TrueBit、Arbitrum、PoD、RanCk+SenCk 的场景对比实验配置。
- `src/ValidatorAudit.sol`：链上审计 benchmark 合约。
- `test/ValidatorAudit.t.sol`：Foundry gas 测试。
- `logs/raw_experiment_log.json`：实验原始日志。

## 环境配置

需要 Go、Foundry、Python，以及 Python 包 `matplotlib`、`numpy`。

```bash
go version
forge --version
python3 -c "import matplotlib, numpy; print('python deps ok')"
```

如果 Go 或 Matplotlib 缓存目录不可写：

```bash
export GOCACHE="${TMPDIR:-/tmp}/chapter5-gocache"
export MPLCONFIGDIR="${TMPDIR:-/tmp}/mplconfig-ch5"
mkdir -p "$GOCACHE" "$MPLCONFIGDIR"
```

## 运行测试

```bash
go test ./...
forge test --gas-report
```

## 生成 raw log

```bash
go run ./cmd/audit-exp --out logs/raw_experiment_log.json
go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_experiments.json
```

实验入口会解析 Foundry gas report 的函数级平均 Gas。raw log 会记录 Foundry 实测值、测量来源、PoD watchtower trace 和月度 Gas 派生公式；图 30-31 基于这些函数级 Gas 形成。如果 Foundry 不可用，Gas 数据会被标记为不完整。
`cmd/comparison-exp` 可单独输出 TrueBit、Arbitrum、PoD、RanCk+SenCk 的对比场景，以及 PoD MineBounty 的 `podMineBounty` 函数、单 epoch Gas 和月度 Gas。

## 生成可视化

回到 `chapter5` 目录运行：

```bash
cd ..
python3 visualization/chapter5_all_figures.py
```

输出：

- `visualization/chapter5_experiment_data.json`
- `experiment/logs/comparison_experiments.json`
- `visualization/正确图片输出/fig24_feasibility_ab.png`
- `visualization/正确图片输出/fig25_joint_feasibility.png`
- `visualization/正确图片输出/fig26_ranck_detection.png`
- `visualization/正确图片输出/fig27_senck_passthrough.png`
- `visualization/正确图片输出/fig28_monte_carlo_and_gate.png`
- `visualization/正确图片输出/fig29_overhead.png`
- `visualization/正确图片输出/fig30_gas_comparison.png`
- `visualization/正确图片输出/fig31_tradeoff.png`

命令与产物对应关系：

| 命令 | 实验内容 | 产物 | 支撑论文结果 |
| --- | --- | --- | --- |
| `go test ./...` | RanCk/SenCk 协议行为、PoD trace、表 11 对比模型和 Gas 派生逻辑测试 | 测试输出 | 表 9-11、图 24-31 的协议实现可信度 |
| `forge test --gas-report` | `TrackInit`、`HBRespond`、`ContAudit`、`SentReport`、`Dispute`、`SentProve`、`podMineBounty` 函数级 Gas | Foundry gas report | 表 10、图 30、图 31 |
| `go run ./cmd/audit-exp --out logs/raw_experiment_log.json` | RanCk/SenCk trace、Monte Carlo、可行域、旁路开销、Gas trace、表 11 场景对比 | `logs/raw_experiment_log.json` | 表 9-11、图 24-31 |
| `go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_experiments.json` | standalone 能力对比和 PoD MineBounty baseline | `logs/comparison_experiments.json` | 表 11、图 30/31 PoD 对照 |
| `python3 visualization/chapter5_all_figures.py` | raw log 校验、结构化实验产物生成、图表渲染 | `../visualization/chapter5_experiment_data.json`、`../visualization/正确图片输出/*.png` | 图 24-31、表 9-11 |

结构化产物与图表对应关系：

| 字段 | 证据来源 | 输出 |
| --- | --- | --- |
| `table9_parameters` | `audit.DefaultParams()` 与 `audit.Table9Parameters()` | 表 9 |
| `feasibility` | C1/C2 罚没边界和联合可行域公式 trace | 图 24、图 25 |
| `detection` | RanCk 心跳/连续性检测公式、SenCk lazy pass 公式 | 图 26、图 27 |
| `monte_carlo` | 固定随机种子 Monte Carlo trace 与 Gamma hit sweep | 图 28 |
| `overhead` | instrumented EVM opcode hook、rw 编码、哈希、Gamma、快照加载、局部重放 trace | 图 29 |
| `gas` | Foundry gas report、PoD watchtower trace、月度 Gas 公式 | 表 10、图 30、图 31 |
| `comparison_experiments`、`comparison_table_11` | TrueBit、Arbitrum、PoD、RanCk+SenCk 场景模型 | 表 11 |

## 注意事项

- `out/`、`cache/`、`__pycache__/`、生成 PDF 和其他可重建产物不提交。
- 如果重新生成图表，应先重新运行实验入口或让可视化脚本自动补齐 raw log。
- 所有图表数据均来自 RanCk/SenCk 协议实现、SenCk instrumented EVM opcode hook、公式扫描、Monte Carlo trace、对比实验模型或 Foundry 输出。

## 验收流程

1. 运行 `go test ./...` 和 `forge test --gas-report`。
2. 运行 `go run ./cmd/audit-exp --out logs/raw_experiment_log.json`，重新生成协议 trace、Monte Carlo、Gas trace 和表 11 场景对比。
3. 运行 `go run ./cmd/comparison-exp --raw logs/raw_experiment_log.json --out logs/comparison_experiments.json`，导出对比报告。
4. 回到 `chapter5` 运行 `python3 visualization/chapter5_all_figures.py`。
5. 回到仓库根目录运行 `python3 scripts/check_experiment_coverage.py`。
