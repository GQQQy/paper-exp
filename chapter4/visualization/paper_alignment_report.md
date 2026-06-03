# Chapter 4 Paper Alignment Report

## Evidence Chain

- Raw local execution log: `experiment/logs/raw_experiment_log.json`
- Visualization data: `visualization/chapter4_experiment_data.json`
- Paper-scale target: PDF Section 4.5 and Figures 14-22 of Chapter 4
- Rule: analytic curves use the formulas specified in the PDF; Monte Carlo figures execute the stated 1000-run simulation; engineering overhead uses Foundry/snarkjs outputs only.

## Coverage Matrix

| Item | Paper Target | Method Stated In PDF | Data Source Used Here | Status | Detail |
| --- | --- | --- | --- | --- | --- |
| 表 5 | 默认参数与质押分布 | PDF 4.5.1 | 脚本参数 + 对数正态样本 | PASS | seed=153456, honest_total=32340.43 ETH, beacon_ids=803 |
| 表 6 | 五种选举方案差异 | PDF 4.5.2 | 绘图曲线标签与计算分支 | PASS | Linear-WR/Linear-WoR/Uniform-WoR/CTWR single/CTWR opt 均覆盖 |
| 图 14 | CTWR 捕获概率分析 | 定理 4.4 解析公式 | 逐点解析计算 | PASS | rho=0.35: rho_b=0.4039, rho_eff=0.3496 |
| 图 15 | 委员会规模影响 | 定理 4.4 解析公式 | 逐 N 解析计算 | PASS | N=40 Uniform/CTWR=3.23e+02 |
| 图 16 | 拆分身份捕获概率 | 解析扫描 | rho=0.20/0.35/0.50 逐 k 扫描 | PASS | B=8085/17414/32340 ETH |
| 图 17 | CTWR 有效权重 | 解析扫描 | 复用图 16 rho=0.35 逐 k 结果 | PASS | k*=535 |
| 图 18 | 女巫身份与期望收益 | PDF 4.5.3 蒙特卡洛 | 1000 次/每 k 选举仿真 | PASS | Monte Carlo CTWR election simulation, k*=8, seed_base=2026 |
| 图 19 | 最优女巫策略 | 整数最优拆分扫描 | 由同一质押样本和 rho=0.35 推导 | PASS | c_reg 与 S_cap 网格逐点求整数最优 k |
| 图 20 | 诚实验证者期望收益 | 经济参数敏感性 | 固定 500x64 ETH，R_total/c_op 参数扫描 | PASS | R_total 与 c_op 曲线均由公式逐点计算 |
| 图 21 | 非比例权重规则对比 | 解析对照 | Sqrt-WoR 与 CTWR 逐 rho/逐 k 计算 | PASS | 权重函数差异显式计算 |
| 图 22 | 匿名质押开销 | 本地工程实验 | forge gas + snarkjs R1CS/witness | PASS | gas ratio=2.77x, cycle ratio=6.83x, PresentCred=PASS, AnonyReg=PASS, AnonyStake=PASS |
| 图 22 附加 | 链下 Groth16 证明时间 | snarkjs groth16 smoke | 脚本已修，完整证明本轮按长耗时跳过 | SKIP | 不使用伪证明时间；图中展示真实 gas 与 R1CS 规模 |

## Engineering Evidence

- Foundry gas: AnonyReg/PublicReg=2.77x, n=200 AnoSt/nonanon=6.83x.
- R1CS constraints: {'PresentCred': 5817, 'AnonyReg': 1088, 'AnonyStake': 1328}.
- Witness checks: PresentCred=PASS, AnonyReg=PASS, AnonyStake=PASS.
