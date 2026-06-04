# 第四章实验复现说明

本目录对应论文第四章“面向链下计算的验证者安全选举”。实验覆盖论文 4.5 节与图 14-22，目标是复现 AnoSt 匿名质押电路、CTWR 安全选举、女巫收益分析、非比例权重对比和工程开销测试。

实验链路为：

```text
experiment/circuits/*.circom
  -> experiment/build/circuits/*.r1cs + *_js/*.wasm
  -> experiment/scripts/verify_circuits.js
  -> experiment/build/witness/*.wtns + circuit_verification_summary.json
  -> visualization/chapter4_all_figures.py
  -> experiment/logs/raw_experiment_log.json
  -> visualization/chapter4_experiment_data.json
  -> visualization/正确图片输出/*.png
```

## 实验总览

默认参数如下。

| 参数 | 值 | 用途 |
| --- | --- | --- |
| 合格候选验证者数量 `|V_qual|` | `1000` | 选举候选池规模 |
| 最低质押门槛 `D_min` | `32 ETH` | 单个候选身份进入选举的最低质押 |
| 截断上限 `S_cap` | `256 ETH` | CTWR 对单身份有效权重的截断上限 |
| 票据粒度 `Delta` | `1 ETH` | 权重离散化粒度 |
| 委员会规模 `N` | `20` | 每轮选出的验证者数量 |
| 捕获安全预算 `delta_cap` | `1e-6` | 捕获概率安全目标 |
| 注册成本 `c_reg` | `0.5 ETH` | 身份拆分攻击的注册成本参数 |
| 单轮运营成本 `c_op` | `0.01 ETH` | 诚实验证者参与成本 |
| 单轮激励池 `R_total` | `1 ETH` | 默认收益分析激励池 |
| 诚实验证者质押分布 | 对数正态分布，均值 `64 ETH`，标准差 `32 ETH` | 捕获概率与 CTWR 权重计算输入 |
| 固定预算激励计算 | 固定对手总预算 `S_A`，扫描女巫身份数 `k`、注册成本 `c_reg` 与截断上限 `S_cap` | 女巫数量、收益下界与最优分裂策略 |

当前结构化数据中的诚实候选样本为 `500` 个，质押均值约 `64.681 ETH`，总质押约 `32340.431 ETH`，beacon 身份数量为 `803`。

## 测试内容与结果

| 测试/图表 | 测试对象 | 数据来源 | 当前关键结果 | 输出 |
| --- | --- | --- | --- | --- |
| 图 14a 捕获概率随权益比变化 | Linear-WR、Linear-WoR、Uniform-WoR、CTWR single、CTWR opt 的委员会完全捕获概率 | `figures.fig14.series`，解析公式逐点计算 | 在 `rho=0.35,N=20` 时，Uniform-WoR 捕获概率 `1.078e-08`，CTWR opt 为 `9.116e-11`，约低两个数量级 | `visualization/正确图片输出/fig14.png` |
| 图 14b 注册成本影响 | 不同 `c_reg` 下 CTWR 最优拆分后的捕获概率 | `figures.fig14.by_c_reg` | `c_reg` 越大，对手可行拆分数量越少，中低权益比区间安全性提升更明显 | `visualization/正确图片输出/fig14.png` |
| 图 15 委员会规模影响 | 固定 `rho=0.35` 时捕获概率随委员会规模 `N` 的衰减 | `figures.fig15` | `rho_beacon=0.403860`，`rho_eff=0.349556`；CTWR 的有效对手占比低于 Uniform-WoR 的 beacon 占比 | `visualization/正确图片输出/fig15.png` |
| 图 16 拆分身份影响 | 对手总权益固定时，拆成 `k` 个身份后各方案捕获概率变化 | `figures.fig16_17` | `rho=0.20/0.35/0.50` 下 CTWR 的最优拆分点分别为 `k*=248/535/787`，捕获概率峰值分别为 `4.338e-15/4.912e-10/7.459e-07` | `visualization/正确图片输出/fig16.png` |
| 图 17 CTWR 有效权重 | `rho=0.35` 时 CTWR 有效权重和有效对手占比随 `k` 变化 | `figures.fig16_17["0.35"]` | 有效对手占比保持在原始 `rho=0.35` 附近或以下，截断机制限制单身份权重集中 | `visualization/正确图片输出/fig17.png` |
| 图 18 女巫身份与收益 | 女巫身份数量对诚实验证者单位收益和对手净利润的影响 | `figures.fig18`，固定对手预算下的 CTWR 收益函数逐点计算 | `S_A/S_cap=8`；`c_reg=2.0` 时 CTWR 对手利润在 `k=8` 达峰，盈亏平衡点约 `k=14` | `visualization/正确图片输出/fig18.png` |
| 图 19 最优女巫策略 | 注册成本、截断上限对最优女巫身份数和诚实收益下界的影响 | `figures.fig19`，固定预算整数拆分扫描 | 注册成本越高，最优拆分身份数越低；收益下界呈整数 `k*` 引起的阶梯跳变 | `visualization/正确图片输出/fig19.png` |
| 图 20 诚实验证者期望收益 | 激励池 `R_total`、运营成本 `c_op` 对诚实验证者净收益的影响 | `figures.fig20` | 当 `R_total=5,10,15,20,30` 且 `c_op=0.005` 时收益均为正；当 `R_total=10` 且 `c_op>=0.02` 时收益转负 | `visualization/正确图片输出/fig20.png` |
| 图 21 非比例权重规则对比 | CTWR 与 Sqrt-WoR 的捕获概率、有效占比、诚实收益对比 | `figures.fig21` | Sqrt-WoR 能削弱大额质押线性优势，但捕获概率和有效对手占比整体高于 CTWR | `visualization/正确图片输出/fig21.png` |
| 图 22 匿名质押开销 | AnoSt 匿名注册、匿名质押、凭证展示、CTWR 选举的链上 Gas 与电路约束 | `foundry_gas_runs`、`zk_circuits`、`figures.fig22` | 单次 Gas：AnonyReg `196706`、AnonyStake `152449`、PresentCred `212746`、ElectCTWR `622254`；`N=100` 时匿名周期总 Gas 为 `56190.1K`，非匿名最小基线为 `8225.4K` | `visualization/正确图片输出/fig22.png` |

## 电路验证结果

`node scripts/verify_circuits.js` 会为三个电路生成样例输入、witness，并运行 `snarkjs wtns check`。当前结果如下。

| 电路 | 对应接口 | 测试内容 | R1CS 约束数 | Witness 状态 |
| --- | --- | --- | ---: | --- |
| `anost_register.circom` | `AnonyReg` | 检查注册承诺、存款承诺、注册 nullifier 与公开输入一致 | `1088` | `PASS` |
| `anost_stake.circom` | `AnonyStake` | 检查存款承诺、质押承诺、找零承诺、余额约束和质押 nullifier | `1328` | `PASS` |
| `anost_credential.circom` | `PresentCred` | 检查注册 Merkle 路径、凭证 nullifier 和质押承诺 | `5817` | `PASS` |

`experiment/scripts/groth16_smoke.js` 是可选证明/验证流程 smoke test，不是图 14-22 的必要输入。

## 链上 Gas 测试结果

`forge test --gas-report` 会运行 7 个测试，覆盖公开基线、匿名接口和 CTWR 选举。

| Foundry 测试 | 测试内容 | 当前 Gas |
| --- | --- | ---: |
| `testPublicReg` | 公开注册，把 identity commitment 写入 registry | `70931` |
| `testPublicStake` | 公开质押最小基线，计算 stake commitment hash | `5657` |
| `testCandidateDeclare` | 候选声明最小基线，检查最低质押并生成候选承诺 | `5666` |
| `testAnonyReg` | 匿名注册，检查 nullifier、模拟 Groth16 pairing 开销并写 registry | `196706` |
| `testAnonyStake` | 匿名质押，检查 nullifier、模拟证明校验并生成质押/找零承诺 | `152449` |
| `testPresentCred` | 凭证展示，检查 Merkle path、nullifier 和证明校验 | `212746` |
| `testElectCTWR` | 对 64 个候选按截断权重无放回选出 20 人 | `622254` |

图 22 的非匿名项是最小公开基线，不包含完整公开质押系统可能需要的所有状态写入和校验，因此它主要用于说明当前合约 benchmark 下的相对开销。

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
cd ..
```

## 运行完整实验

从 `chapter4` 目录运行：

```bash
cd experiment
node scripts/verify_circuits.js
forge test --gas-report

cd ..
python3 visualization/chapter4_all_figures.py
```

每条命令的作用：

- `node scripts/verify_circuits.js`：生成 witness，执行 `snarkjs wtns check`，校验三类匿名质押电路的公开输出。
- `forge test --gas-report`：测量公开注册/质押/声明、匿名注册/质押/凭证展示和 CTWR 选举 Gas。
- `chapter4_all_figures.py`：重新执行 Foundry Gas、检查/编译电路、运行 witness 检查、读取 R1CS 约束数、计算 CTWR 捕获概率与固定预算激励曲线，并生成图 14-22。

## 可选 Groth16 Smoke Test

该步骤用于确认本地 proving/verifying 流程可跑通，不是生产可信设置，也不是图 14-22 的必需输入。

```bash
cd experiment
mkdir -p build/zk
npx snarkjs powersoftau new bn128 13 build/zk/pot13_0000.ptau -v
npx snarkjs powersoftau contribute build/zk/pot13_0000.ptau build/zk/pot13_0001.ptau --name="chapter4-local-smoke" -e="chapter4 deterministic local smoke entropy"
npx snarkjs powersoftau prepare phase2 build/zk/pot13_0001.ptau build/zk/pot13_final.ptau -v
node scripts/groth16_smoke.js
cd ..
```

## 只重新生成可视化

如果依赖已安装，且想从当前实验源码和构建产物刷新全部图表：

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

## 快速检查数据

```bash
jq '.zk_circuits[] | {name, constraint_count, witness_status}' visualization/chapter4_experiment_data.json
jq '.foundry_gas_runs[] | {function, gas}' visualization/chapter4_experiment_data.json
jq '.figures.fig14.trace_rho_035' visualization/chapter4_experiment_data.json
jq '.figures.fig18 | {method, k_threshold_s_adv_over_s_cap, k_star_profit, k_break_even}' visualization/chapter4_experiment_data.json
jq '.figures.fig22.single_call' visualization/chapter4_experiment_data.json
```

## 注意事项

- `chapter4_all_figures.py` 会删除并重建 `visualization/正确图片输出/` 下的 PNG 文件。
- `build/`、`node_modules/`、`out/`、`cache/` 和本地 `tools/bin/circom` 都是可重建产物，不提交到 git。
- 如果缺少 `snarkjs`，在 `experiment` 下运行 `npm install`。
- 如果 `forge test --gas-report` 输出格式变化，脚本可能解析不到 per-test Gas 行，需要同步更新 `run_foundry_gas()` 的解析逻辑。
