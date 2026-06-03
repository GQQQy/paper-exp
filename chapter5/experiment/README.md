# 第五章验证者工作审计实验工程

本目录用于运行第五章“面向链下计算的验证者工作审计”实验，覆盖 RanCk 随机连续性抽查、SenCk 哨兵抽样勤勉检测、链下旁路审计开销、链上 Gas 基准与论文图 24-31/表 9-11 的数据生成。

## 实验内容

- RanCk：生成验证者私有 seed/nonce、提交 `TC_0`、维护 `alpha_t = H(alpha_{t-1} || bh_t || PRF(seed_i,t) || tau_t)`、按 `Trigger(t,i)=H(bh_t||tid||i) mod M` 触发心跳，并对端点和抽样点执行连续性审计。
- SenCk：在确定性 EVM 风格执行步中采集 `rw_t`、`val_t`、`pc_t/op_t`，计算 `Gamma(r,tid,k,t,rw_t)`，生成哨兵事件和段摘要 `dig_k`，再从快照段局部重放提交 `SentReport`。
- 行为模型：覆盖诚实在线、完全离线、间歇在线、补算失败、凭证链不一致、惰性猜摘要、执行方污染 `ComAud` 等偏离。
- Foundry：提供最小审计合约和 gas benchmark 测试，覆盖 `TrackInit`、`HBRespond`、`ContAudit`、`SentReport`、`Dispute`、`SentProve` 和 PoD baseline。
- 可视化：从 `logs/raw_experiment_log.json` 生成结构化数据、图 24-31 和 `paper_alignment_report.md`。

## 目录说明

- `audit/protocol.go`：RanCk/SenCk 协议实现、仿真、Monte Carlo、旁路开销和 Gas 校准数据结构。
- `audit/protocol_test.go`：RanCk、SenCk 和完整审计流单元/集成测试。
- `cmd/audit-exp/main.go`：实验入口，生成 `logs/raw_experiment_log.json`。
- `src/ValidatorAudit.sol`：链上审计 benchmark 合约。
- `test/ValidatorAudit.t.sol`：Foundry gas 测试。
- `logs/raw_experiment_log.json`：实验原始日志。

## 环境配置

需要：

```bash
/usr/local/go/bin/go version
/Users/gqy/.foundry/bin/forge --version
python3 -c "import matplotlib, numpy; print('python deps ok')"
```

如果本机 PATH 已包含 `go` 与 `forge`，也可以直接使用短命令。

## 运行测试

```bash
cd chapter5/experiment
mkdir -p /private/tmp/chapter5-gocache
GOCACHE=/private/tmp/chapter5-gocache /usr/local/go/bin/go test ./...
/Users/gqy/.foundry/bin/forge test --gas-report
```

## 生成 raw log

```bash
cd chapter5/experiment
GOCACHE=/private/tmp/chapter5-gocache /usr/local/go/bin/go run ./cmd/audit-exp --out logs/raw_experiment_log.json
```

实验入口会解析 Foundry per-test gas。由于本地测试合约的 per-test gas 会包含 harness/setup 效应，raw log 同时记录 Foundry 实测值、论文表 10 目标值和 target/measured 校准因子；图 30-31 使用论文表 10 目标值生成，并在报告中说明校准来源。

## 生成可视化与报告

在仓库根目录运行：

```bash
python3 chapter5/visualization/chapter5_all_figures.py
```

输出：

- `chapter5/visualization/chapter5_experiment_data.json`
- `chapter5/visualization/正确图片输出/fig24_feasibility_ab.png`
- `chapter5/visualization/正确图片输出/fig25_joint_feasibility.png`
- `chapter5/visualization/正确图片输出/fig26_ranck_detection.png`
- `chapter5/visualization/正确图片输出/fig27_senck_passthrough.png`
- `chapter5/visualization/正确图片输出/fig28_monte_carlo_and_gate.png`
- `chapter5/visualization/正确图片输出/fig29_overhead.png`
- `chapter5/visualization/正确图片输出/fig30_gas_comparison.png`
- `chapter5/visualization/正确图片输出/fig31_tradeoff.png`
- `chapter5/visualization/paper_alignment_report.md`

## 注意事项

- `out/`、`cache/`、`__pycache__/` 和其他可重建产物不提交。
- 如果重新生成图表，应先重新运行实验入口或让可视化脚本自动补齐 raw log。
- 所有图表数据均来自协议实现、公式扫描、Monte Carlo trace、Foundry 输出或报告中记录的表 10 校准参数。
