# 第四章 AnoSt 实验工程

本目录用于运行第四章“面向链下计算的验证者安全选举”实验，覆盖匿名质押电路验证、链上 Gas 基准、CTWR 安全选举仿真和图 14-22 的可视化数据生成。

## 实验内容

实验流程包含以下部分：

- 匿名准入电路：编译并验证 `AnostRegister`、`AnostStake`、`AnostCredential` 三个 Circom 电路，检查 Poseidon 承诺、nullifier、余额约束和 Merkle 路径约束。
- Groth16 本地 smoke test：基于本地 Powers of Tau 参数生成证明、验证证明，用于确认电路证明流程可跑通。
- 链上 Gas 测量：调用 Foundry `forge test --gas-report`，统计公开注册/质押/声明与匿名注册/质押/凭证展示/CTWR 选举的 Gas 开销。
- 论文参数实验：按照第四章 4.5 节参数生成 CTWR 捕获概率、委员会规模、女巫拆分、诚实收益、参数敏感性和工程开销相关 trace。

## 目录说明

- `circuits/anost_register.circom`：匿名注册电路，输出注册承诺、存款承诺和注册 nullifier。
- `circuits/anost_stake.circom`：匿名质押电路，验证存款承诺并输出质押承诺、找零承诺和质押 nullifier。
- `circuits/anost_credential.circom`：凭证展示电路，验证注册 Merkle 路径并输出凭证 nullifier 和质押承诺。
- `scripts/verify_circuits.js`：生成样例 witness，运行 `snarkjs wtns check`，并校验 witness 中的关键公开输出。
- `scripts/groth16_smoke.js`：对三个电路执行本地 Groth16 setup/prove/verify smoke test。
- `src/AnoStBenchmarks.sol`：第四章匿名质押与 CTWR 选举的 Solidity 基准合约。
- `test/AnoStBenchmarks.t.sol`：Foundry Gas 测试入口。
- `logs/raw_experiment_log.json`：实验原始日志，与可视化数据保持一致。
- `../visualization/chapter4_all_figures.py`：第四章数据生成、绘图和论文对齐报告入口。
- `../visualization/chapter4_experiment_data.json`：图 14-22 使用的结构化实验数据。
- `../visualization/正确图片输出/`：图 14-22 的 PNG 输出。
- `../visualization/paper_alignment_report.md`：第四章论文实验分析对齐报告。

## 环境配置

需要安装以下工具：

- Node.js 18 或更高版本。
- Circom 2.1.8，用于编译 `.circom` 电路。
- Foundry，包括 `forge`。
- Python 3.10 或更高版本。
- Python 包：`matplotlib`、`numpy`、`Pillow`。

检查命令：

```bash
node --version
circom --version
forge --version
python3 --version
python3 -c "import matplotlib, numpy, PIL; print('python deps ok')"
```

安装 Node 依赖：

```bash
cd chapter4/experiment
npm install
node -e "console.log(require('./node_modules/snarkjs/package.json').version)"
```

如果缺少 Python 包：

```bash
python3 -m pip install matplotlib numpy Pillow
```

如果本机没有全局 `circom`，也可以把 `circom` 二进制放到 `chapter4/experiment/tools/bin/circom`，后续命令把 `circom` 替换为 `./tools/bin/circom` 即可。

## 编译电路

在 `chapter4/experiment` 目录下运行：

```bash
mkdir -p build/circuits
circom circuits/anost_register.circom --r1cs --wasm --sym -o build/circuits
circom circuits/anost_stake.circom --r1cs --wasm --sym -o build/circuits
circom circuits/anost_credential.circom --r1cs --wasm --sym -o build/circuits
```

编译完成后应生成：

- `build/circuits/anost_register.r1cs`
- `build/circuits/anost_register_js/anost_register.wasm`
- `build/circuits/anost_stake.r1cs`
- `build/circuits/anost_stake_js/anost_stake.wasm`
- `build/circuits/anost_credential.r1cs`
- `build/circuits/anost_credential_js/anost_credential.wasm`

## 验证电路 witness

在 `chapter4/experiment` 目录下运行：

```bash
node scripts/verify_circuits.js
```

脚本会生成样例输入、witness 和检查报告：

- `build/witness/anost_register.input.json`
- `build/witness/anost_stake.input.json`
- `build/witness/anost_credential.input.json`
- `build/witness/*.wtns`
- `build/witness/*.witness.json`
- `build/witness/circuit_verification_summary.json`

期望三个电路的 `status` 均为 `PASS`。

## 运行 Groth16 smoke test

该步骤用于本地证明流程自检，不是生产可信设置。先生成本地 Powers of Tau 参数：

```bash
mkdir -p build/zk
npx snarkjs powersoftau new bn128 13 build/zk/pot13_0000.ptau -v
npx snarkjs powersoftau contribute build/zk/pot13_0000.ptau build/zk/pot13_0001.ptau --name="chapter4-local-smoke" -e="chapter4 deterministic local smoke entropy"
npx snarkjs powersoftau prepare phase2 build/zk/pot13_0001.ptau build/zk/pot13_final.ptau -v
```

然后运行：

```bash
node scripts/groth16_smoke.js
```

输出文件位于 `build/zk/`，包括每个电路的 zkey、verification key、proof、public signals 和 `groth16_smoke_summary.json`。

## 运行链上 Gas 基准

在 `chapter4/experiment` 目录下运行：

```bash
forge test --gas-report
```

期望看到以下测试项通过：

```text
testPublicReg()
testPublicStake()
testCandidateDeclare()
testAnonyReg()
testAnonyStake()
testPresentCred()
testElectCTWR()
```

`../visualization/chapter4_all_figures.py` 会解析这些测试的 per-test Gas 行，用于生成图 22。

## 生成可视化结果

在仓库根目录运行：

```bash
python3 chapter4/visualization/chapter4_all_figures.py
```

该脚本会自动执行：

- `forge test --gas-report`
- `node chapter4/experiment/scripts/verify_circuits.js`
- `snarkjs r1cs info`
- CTWR 解析计算和 Monte Carlo 选举仿真
- 图 14-22 绘制
- 论文对齐报告生成

输出文件：

- `chapter4/experiment/logs/raw_experiment_log.json`
- `chapter4/visualization/chapter4_experiment_data.json`
- `chapter4/visualization/正确图片输出/fig14.png` 到 `fig22.png`
- `chapter4/visualization/paper_alignment_report.md`

## 结果检查

生成可视化后，重点检查：

- 图 14：CTWR opt、CTWR single、Uniform-WoR、Linear-WR、Linear-WoR 的捕获概率曲线均由解析公式逐点生成。
- 图 15：委员会规模增大时 CTWR 捕获概率应显著低于 Uniform-WoR，默认 `N=40` 的差距应超过 100 倍。
- 图 16-17：对手拆分数量 `k` 与 CTWR 有效权重、有效占比对应一致。
- 图 18：女巫身份数量对诚实验证者收益和对手净利润的影响来自 1000 次/每 `k` 的 Monte Carlo 仿真。
- 图 19-21：注册成本、截断上限、激励池、运营成本和非比例权重规则均由脚本参数扫描得到。
- 图 22：Gas 数据来自 Foundry，R1CS 约束数和 witness 状态来自 snarkjs。

`../visualization/paper_alignment_report.md` 中关键项应为 `PASS`；Groth16 证明时间如果未纳入图表，会以 `SKIP` 标注为长耗时本地 smoke test。

## 注意事项

- `build/`、`node_modules/`、`out/`、`cache/` 和本地 `tools/bin/circom` 都是可重建产物，不提交到 git。
- 重新生成数据前，请先完成 Node 依赖安装和电路编译，否则 witness 检查会因为缺少 `.wasm` 或 `.r1cs` 失败。
- `forge test --gas-report` 的 Gas 通常稳定，运行时间会受机器环境影响。
- `groth16_smoke.js` 会生成较大的本地证明文件，运行时间明显长于 witness 检查；日常复现实验图时通常只需要运行 `chapter4_all_figures.py`。
