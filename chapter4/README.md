# 第四章实验复现说明

本目录对应论文第四章“面向链下计算的验证者安全选举”。实验覆盖论文 4.5 节与图 14-22，数据来自 Circom/snarkjs 电路验证、Foundry Gas 基准、CTWR 解析公式和固定随机种子的 Monte Carlo 仿真。

## 实验内容

- 抗女巫攻击：比较 Linear-WR、Linear-WoR、Uniform-WoR、CTWR single 和 CTWR opt 的委员会完全捕获概率。
- 激励可持续性：通过解析扫描和 Monte Carlo 仿真分析女巫身份数量、注册成本、截断上限、激励池和运营成本对诚实验证者收益的影响。
- 匿名质押开销：编译并验证 AnoSt 的三个 Circom 电路，运行 witness/snarkjs 检查，使用 Foundry 测量公开/匿名接口和 CTWR 选举的链上 Gas。

默认参数：

| 参数 | 值 |
| --- | --- |
| 合格候选验证者数量 `|V_qual|` | `1000` |
| 最低质押门槛 `D_min` | `32 ETH` |
| 截断上限 `S_cap` | `256 ETH` |
| 票据粒度 `Delta` | `1 ETH` |
| 委员会规模 `N` | `20` |
| 捕获安全预算 `delta_cap` | `1e-6` |
| 注册成本 `c_reg` | `0.5 ETH` |
| 单轮运营成本 `c_op` | `0.01 ETH` |
| 单轮激励池 `R_total` | `1 ETH` |
| 诚实验证者质押分布 | 对数正态分布，均值 64 ETH，标准差 32 ETH |
| Monte Carlo | 每个 `k` 1000 次，固定随机种子 |

## 目录结构

- `第四章面向链下计算的验证者安全选举.pdf`：第四章论文正文。
- `experiment/circuits/`：AnoSt 的三个 Circom 电路。
- `experiment/scripts/verify_circuits.js`：生成样例输入和 witness，运行 `snarkjs wtns check`，校验公开输出。
- `experiment/scripts/groth16_smoke.js`：可选 Groth16 本地证明/验证 smoke test。
- `experiment/src/AnoStBenchmarks.sol`：公开准入、匿名准入和 CTWR 选举的 Solidity Gas 基准。
- `experiment/test/AnoStBenchmarks.t.sol`：Foundry Gas 测试入口。
- `experiment/logs/raw_experiment_log.json`：实验原始日志，由主可视化脚本生成。
- `visualization/chapter4_all_figures.py`：数据生成和绘图入口。
- `visualization/chapter4_experiment_data.json`：结构化实验数据。
- `visualization/正确图片输出/`：图 14-22 的 PNG 输出。

## 环境依赖

需要以下工具：

- Node.js 18 或更高版本。
- npm。
- Circom 2.1.8。
- Foundry，包括 `forge`。
- Python 3.10 或兼容版本。
- Python 包：`matplotlib`、`numpy`、`Pillow`。

检查命令：

```bash
node --version
npm --version
circom --version
forge --version
python3 --version
python3 -c "import matplotlib, numpy, PIL; print('python deps ok')"
```

如果没有全局 `circom`，可把 Circom 2.1.8 二进制放到 `experiment/tools/bin/circom`，或设置环境变量：

```bash
export CIRCOM=/path/to/circom
$CIRCOM --version
```

`visualization/chapter4_all_figures.py` 会按以下顺序寻找编译器：

1. `$CIRCOM`
2. 全局 `circom`
3. `experiment/tools/bin/circom`

安装 Node/snarkjs/circomlib 依赖：

```bash
cd experiment
npm install
node -e "console.log(require('./node_modules/snarkjs/package.json').version)"
cd ..
```

安装 Python 依赖：

```bash
python3 -m pip install matplotlib numpy Pillow
```

## 编译 Circom 电路

主脚本会自动检查并编译缺失或过期的电路产物。也可以手动编译：

```bash
cd experiment
mkdir -p build/circuits
CIRCOM_BIN="${CIRCOM:-circom}"
if ! command -v "$CIRCOM_BIN" >/dev/null 2>&1 && [ -x ./tools/bin/circom ]; then
  CIRCOM_BIN=./tools/bin/circom
fi
"$CIRCOM_BIN" circuits/anost_register.circom --r1cs --wasm --sym -o build/circuits
"$CIRCOM_BIN" circuits/anost_stake.circom --r1cs --wasm --sym -o build/circuits
"$CIRCOM_BIN" circuits/anost_credential.circom --r1cs --wasm --sym -o build/circuits
```

## 运行完整实验

从 `chapter4` 目录运行：

```bash
cd experiment
mkdir -p build/circuits
CIRCOM_BIN="${CIRCOM:-circom}"
if ! command -v "$CIRCOM_BIN" >/dev/null 2>&1 && [ -x ./tools/bin/circom ]; then
  CIRCOM_BIN=./tools/bin/circom
fi
"$CIRCOM_BIN" circuits/anost_register.circom --r1cs --wasm --sym -o build/circuits
"$CIRCOM_BIN" circuits/anost_stake.circom --r1cs --wasm --sym -o build/circuits
"$CIRCOM_BIN" circuits/anost_credential.circom --r1cs --wasm --sym -o build/circuits
node scripts/verify_circuits.js
forge test --gas-report

cd ..
python3 visualization/chapter4_all_figures.py
```

说明：

- 前三条 `"$CIRCOM_BIN" ...` 命令编译三个 AnoSt 电路，生成 `build/circuits/*.r1cs` 和 witness 生成器。
- `node scripts/verify_circuits.js` 使用编译后的 wasm 生成 witness，运行 `snarkjs wtns check`，并校验 Poseidon 承诺、nullifier、余额约束和 Merkle 路径相关公开输出。
- `forge test --gas-report` 运行 7 个 Gas 基准测试：`testPublicReg`、`testPublicStake`、`testCandidateDeclare`、`testAnonyReg`、`testAnonyStake`、`testPresentCred`、`testElectCTWR`。
- `chapter4_all_figures.py` 会重新执行 Foundry Gas、自动编译/检查 Circom 电路、运行 witness 检查、读取 R1CS 约束数、计算 CTWR 解析曲线、执行 Monte Carlo 仿真、生成图 14-22，并写出 raw log 和结构化数据。

如遇 Matplotlib 配置目录不可写，可使用临时目录：

```bash
export MPLCONFIGDIR="${TMPDIR:-/tmp}/mplconfig-ch4"
mkdir -p "$MPLCONFIGDIR"
python3 visualization/chapter4_all_figures.py
```

## 可选 Groth16 Smoke Test

Groth16 smoke test 用于确认本地 proving/verifying 流程可跑通，不是生产可信设置，也不是图 14-22 的必需输入。

```bash
cd experiment
mkdir -p build/zk
npx snarkjs powersoftau new bn128 13 build/zk/pot13_0000.ptau -v
npx snarkjs powersoftau contribute build/zk/pot13_0000.ptau build/zk/pot13_0001.ptau --name="chapter4-local-smoke" -e="chapter4 deterministic local smoke entropy"
npx snarkjs powersoftau prepare phase2 build/zk/pot13_0001.ptau build/zk/pot13_final.ptau -v
node scripts/groth16_smoke.js
```

输出位于 `experiment/build/zk/`。

## 只重新生成可视化

如果依赖已安装，且想从当前实验源码和 ignored 构建产物刷新全部图表：

```bash
python3 visualization/chapter4_all_figures.py
```

该命令仍会重新采集 Foundry Gas、重新运行 witness 检查和 R1CS info，并重新计算公式/仿真数据。

## 输出文件

- Raw log：`experiment/logs/raw_experiment_log.json`
- 结构化数据：`visualization/chapter4_experiment_data.json`
- 生成图片：
  - `visualization/正确图片输出/fig14.png`
  - `visualization/正确图片输出/fig15.png`
  - `visualization/正确图片输出/fig16.png`
  - `visualization/正确图片输出/fig17.png`
  - `visualization/正确图片输出/fig18.png`
  - `visualization/正确图片输出/fig19.png`
  - `visualization/正确图片输出/fig20.png`
  - `visualization/正确图片输出/fig21.png`
  - `visualization/正确图片输出/fig22.png`

## 数据链路

```text
experiment/circuits/*.circom
  -> Circom build/circuits/*.r1cs + *_js/*.wasm
  -> experiment/scripts/verify_circuits.js
  -> build/witness/*.wtns + circuit_verification_summary.json
  -> visualization/chapter4_all_figures.py
  -> experiment/logs/raw_experiment_log.json
  -> visualization/chapter4_experiment_data.json
  -> visualization/正确图片输出/*.png
```

主数据来源：

- `zk_circuits`：来自 Circom 源码、`snarkjs r1cs info` 和 witness 检查结果。
- `foundry_gas_runs`：来自 `forge test --gas-report` 的 per-test Gas 行。
- `figures.fig14` 到 `figures.fig17`：来自 CTWR、均匀无放回、线性加权等公式逐点计算。
- `figures.fig18`：来自固定随机种子的 Monte Carlo 选举仿真，每个 `k` 运行 1000 次。
- `figures.fig19` 到 `figures.fig21`：来自注册成本、截断上限、激励池、运营成本和非比例权重规则的参数扫描。
- `figures.fig22`：来自 Foundry Gas、R1CS 约束数和 witness 状态。

快速检查数据：

```bash
jq '.zk_circuits[] | {name, constraint_count, witness_status}' visualization/chapter4_experiment_data.json
jq '.foundry_gas_runs[] | {function, gas}' visualization/chapter4_experiment_data.json
jq '.figures.fig18 | {method, runs_per_k, seed_base, k_star_profit}' visualization/chapter4_experiment_data.json
jq '.metadata.tooling.compiled_circuits' visualization/chapter4_experiment_data.json
```

## 论文图表对应关系

| 论文图表 | 输出图片 | 数据字段 | 生成/测量代码 |
| --- | --- | --- | --- |
| 图 14 CTWR 捕获概率分析 | `fig14.png` | `figures.fig14` | `fig14`、`optimal_ctwr_split`、`capture_uniform_wor`、`capture_weighted_wor` |
| 图 15 委员会规模影响 | `fig15.png` | `figures.fig15` | `fig15` |
| 图 16 捕获概率随拆分身份变化 | `fig16.png` | `figures.fig16_17` | `split_case`、`fig16_17`、`render_fig16` |
| 图 17 CTWR 有效权重分析 | `fig17.png` | `figures.fig16_17["0.35"]` | `split_case`、`render_fig17` |
| 图 18 女巫身份与期望收益 | `fig18.png` | `figures.fig18` | `simulate_group_round`、`attack_opportunity_probability`、`fig18` |
| 图 19 最优女巫策略分析 | `fig19.png` | `figures.fig19` | `best_integer_split`、`fig19` |
| 图 20 诚实验证者期望收益 | `fig20.png` | `figures.fig20` | `honest_revenue`、`fig20` |
| 图 21 非比例权重规则对比 | `fig21.png` | `figures.fig21` | `rho_eff_rule`、`fig21` |
| 图 22 匿名质押开销分析 | `fig22.png` | `figures.fig22`、`zk_circuits`、`foundry_gas_runs` | `AnoStBenchmarks.sol`、`AnoStBenchmarks.t.sol`、`verify_circuits.js`、`snarkjs r1cs info`、`fig22` |

## 注意事项

- `chapter4_all_figures.py` 会删除并重建 `visualization/正确图片输出/` 下的 PNG 文件。
- `build/`、`node_modules/`、`out/`、`cache/` 和本地 `tools/bin/circom` 都是可重建产物，不提交到 git。
- 如果缺少 `snarkjs`，在 `experiment` 下运行 `npm install`。
- 如果 `forge test --gas-report` 输出格式变化，脚本可能解析不到 per-test Gas 行，需要同步更新 `run_foundry_gas()` 的解析逻辑。
