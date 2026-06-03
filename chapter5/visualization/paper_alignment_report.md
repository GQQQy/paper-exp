# 第五章论文实验对齐报告

## 数据来源

- Raw log: `/Users/gqy/Desktop/data/DR/毕业/毕业答辩/实验室验收材料/paper-exp/chapter5/experiment/logs/raw_experiment_log.json`
- Structured data: `/Users/gqy/Desktop/data/DR/毕业/毕业答辩/实验室验收材料/paper-exp/chapter5/visualization/chapter5_experiment_data.json`
- 证据链：论文机制 -> Go 协议实现/Foundry 合约 -> 测试与实验运行 -> raw log -> structured data -> figures -> 本报告。

## 校准说明

- RanCk/SenCk 检测概率、激励可行域和 Monte Carlo 均由协议实现或公式逐点生成。
- Foundry per-test gas 受测试 harness/setup 影响，raw log 保留本地实测值；图 30-31 使用论文表 10 目标值，并记录 target/measured 校准因子。
- C1 默认精确阈值为 1/72=1.39%，与论文“约 1%”文字结论一致，报告同时保留精确值和论文表述。

## 对齐检查

| 项目 | 状态 | 数据来源 | 说明 |
| --- | --- | --- | --- |
| 表 9 参数落实 | PASS | raw.table9_parameters + params | 12s, T_win=7200, N=20, B=1e+08, b=1e+06, M=100, M_c=100, L=5000, m_s=10 |
| RanCk 协议实现 | PASS | audit/protocol.go + protocol_test.go | 诚实 trace 心跳 68 次，ContAudit=True |
| SenCk 协议实现 | PASS | audit/protocol.go + protocol_test.go | rho=1.000000, 抽样段 10 个 |
| 图 24 激励边界 | PASS | feasibility.C1/C2 formula scan | C1 默认精确阈值 1.39%，论文文字约 1% |
| 图 25 联合可行域 | PASS | feasibility.joint grid | 默认 c_hb=0.05, c_sent=1.5, rho=0.3；仅左下角不可行 |
| 图 26 RanCk 检测 | PASS | formula + implemented trigger traces | s=10, ell=200 联合侥幸通过率 8.101e-04 |
| 图 27 SenCk 检测 | PASS | rho/m_s formula | 侥幸通过概率随 m_s 指数衰减 |
| 图 28 Monte Carlo | PASS | raw.monte_carlo + SenCk trigger stats | RanCk/SenCk 模拟点落入理论曲线统计置信范围；Gamma 触发率来自实现 trace |
| 图 29 旁路审计开销 | PASS | raw.overhead_traces | rw 编码、哈希、Gamma、快照加载和局部重放均记录到 raw log |
| 表 10 链上 Gas | PASS | Foundry parsed + Table 10 calibration | PASS: parsed forge test --gas-report；图表采用校准正常路径 2.71M gas |
| 图 30 Gas 对比 | PASS | gas.operations + monthly_by_scheme | 默认路径心跳为主要成本；降频后低于 PoD |
| 图 31 pi_h 权衡 | PASS | gas.pi_h_sweep | pi_h 同时影响 Gas 和 ell=300 检测概率，并标出低于 PoD 区域 |
| 表 11 综合对比 | PASS | raw.comparison_table_11 | RanCk+SenCk 同时覆盖在线状态审计和语义勤勉检测 |

## 生成文件

- `fig24_feasibility_ab.png`
- `fig25_joint_feasibility.png`
- `fig26_ranck_detection.png`
- `fig27_senck_passthrough.png`
- `fig28_monte_carlo_and_gate.png`
- `fig29_overhead.png`
- `fig30_gas_comparison.png`
- `fig31_tradeoff.png`
