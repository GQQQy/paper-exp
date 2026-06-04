# 第四章 AnoSt 实验工程

本目录用于运行第四章“面向链下计算的验证者安全选举”实验，覆盖匿名质押电路验证、链上 Gas 基准、CTWR 安全选举仿真和图 14-22 的可视化数据生成。

## 实验内容

- 匿名准入电路：编译并验证 `AnostRegister`、`AnostStake`、`AnostCredential` 三个 Circom 电路，检查 Poseidon 承诺、nullifier、余额约束和 Merkle 路径约束。
- Groth16 本地 smoke test：基于本地 Powers of Tau 参数生成证明、验证证明，用于确认电路证明流程可跑通。
- 链上 Gas 测量：调用 Foundry `forge test --gas-report`，统计公开注册/质押/声明与匿名注册/质押/凭证展示/CTWR 选举的 Gas 开销。
- 参数实验：按照第四章 4.5 节参数生成 CTWR 捕获概率、委员会规模、女巫拆分、诚实收益、参数敏感性和工程开销相关 trace。

## 目录说明

- `circuits/anost_register.circom`：匿名注册电路，输出注册承诺、存款承诺和注册 nullifier。
- `circuits/anost_stake.circom`：匿名质押电路，验证存款承诺并输出质押承诺、找零承诺和质押 nullifier。
- `circuits/anost_credential.circom`：凭证展示电路，验证注册 Merkle 路径并输出凭证 nullifier 和质押承诺。
- `scripts/verify_circuits.js`：生成样例 witness，运行 `snarkjs wtns check`，并校验 witness 中的关键公开输出。
- `scripts/groth16_smoke.js`：对三个电路执行本地 Groth16 setup/prove/verify smoke test。
- `src/AnoStBenchmarks.sol`：第四章匿名质押与 CTWR 选举的 Solidity 基准合约。
- `test/AnoStBenchmarks.t.sol`：Foundry Gas 测试入口。
- `logs/raw_experiment_log.json`：实验原始日志，与可视化数据保持一致。
- `../visualization/chapter4_all_figures.py`：第四章数据生成和绘图入口。
- `../visualization/chapter4_experiment_data.json`：图 14-22 使用的结构化实验数据。
- `../visualization/正确图片输出/`：图 14-22 的 PNG 输出。

## 环境配置

需要安装 Node.js、npm、Circom 2.1.8、Foundry、Python，以及 Python 包 `matplotlib`、`numpy`、`Pillow`。

检查命令：

```bash
node --version
npm --version
circom --version
forge --version
python3 -c "import matplotlib, numpy, PIL; print('python deps ok')"
```

安装 Node 依赖：

```bash
npm install
node -e "console.log(require('./node_modules/snarkjs/package.json').version)"
```

如果缺少 Python 包：

```bash
python3 -m pip install matplotlib numpy Pillow
```

如果本机没有全局 `circom`，也可以把 `circom` 二进制放到 `tools/bin/circom`，或设置 `CIRCOM=/path/to/circom`。

## 编译和验证电路

在当前 `experiment` 目录下运行：

```bash
mkdir -p build/circuits
CIRCOM_BIN="${CIRCOM:-circom}"
if ! command -v "$CIRCOM_BIN" >/dev/null 2>&1 && [ -x ./tools/bin/circom ]; then
  CIRCOM_BIN=./tools/bin/circom
fi
"$CIRCOM_BIN" circuits/anost_register.circom --r1cs --wasm --sym -o build/circuits
"$CIRCOM_BIN" circuits/anost_stake.circom --r1cs --wasm --sym -o build/circuits
"$CIRCOM_BIN" circuits/anost_credential.circom --r1cs --wasm --sym -o build/circuits
node scripts/verify_circuits.js
```

脚本会生成样例输入、witness 和 `build/witness/circuit_verification_summary.json`。期望三个电路的 `status` 均为 `PASS`。

## 运行链上 Gas 基准

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

## 生成可视化

回到 `chapter4` 目录运行：

```bash
cd ..
python3 visualization/chapter4_all_figures.py
```

该脚本会自动执行：

- `forge test --gas-report`
- `node experiment/scripts/verify_circuits.js`
- `snarkjs r1cs info`
- CTWR 解析计算和 Monte Carlo 选举仿真
- 图 14-22 绘制

输出：

- `experiment/logs/raw_experiment_log.json`
- `visualization/chapter4_experiment_data.json`
- `visualization/正确图片输出/fig14.png` 到 `fig22.png`

## 可选 Groth16 Smoke Test

该步骤用于本地证明流程自检，不是生产可信设置，也不是图 14-22 的必需输入。

```bash
mkdir -p build/zk
npx snarkjs powersoftau new bn128 13 build/zk/pot13_0000.ptau -v
npx snarkjs powersoftau contribute build/zk/pot13_0000.ptau build/zk/pot13_0001.ptau --name="chapter4-local-smoke" -e="chapter4 deterministic local smoke entropy"
npx snarkjs powersoftau prepare phase2 build/zk/pot13_0001.ptau build/zk/pot13_final.ptau -v
node scripts/groth16_smoke.js
```

## 注意事项

- `build/`、`node_modules/`、`out/`、`cache/` 和本地 `tools/bin/circom` 都是可重建产物，不提交到 git。
- 重新生成数据前，请先完成 Node 依赖安装；主脚本会自动编译缺失或过期的电路产物。
- `forge test --gas-report` 的 Gas 通常稳定，运行时间会受机器环境影响。
- `groth16_smoke.js` 会生成较大的本地证明文件，运行时间明显长于 witness 检查。
