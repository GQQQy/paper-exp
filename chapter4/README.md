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

第四章结果由三类可复现实验产物汇总形成：

- CTWR、捕获概率、女巫拆分、收益和非比例权重对比来自 `visualization/chapter4_all_figures.py` 中的公式 trace 与固定随机种子候选样本，输出到 `figures.fig14` 至 `figures.fig21`。
- 图 22 的链上 Gas 来自 `forge test --gas-report` 的函数级 benchmark，输出到 `foundry_gas_runs` 与 `figures.fig22.single_call`。
- 图 22 的电路约束和 witness 状态来自 Circom 编译产物、`snarkjs r1cs info` 与 `node scripts/verify_circuits.js`，输出到 `zk_circuits` 与 `figures.fig22.circuit_stats`。

验收时应确认 `experiment/logs/raw_experiment_log.json` 与 `visualization/chapter4_experiment_data.json` 中的 `protocol_trace`、`zk_circuits`、`foundry_gas_runs` 完全一致，图 14-22 的 PNG 再由结构化实验产物渲染生成。

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

## 代码与论文结果定位

论文第四章 4.5 节的表 5、表 6 和图 14-22 可按下表定位到具体代码。第四章的主入口是 `visualization/chapter4_all_figures.py`：它会重新采集 Foundry Gas、检查/编译 Circom 电路、运行 witness 检查、汇总 R1CS 约束与 CTWR 公式 trace，生成 `experiment/logs/raw_experiment_log.json` 和 `visualization/chapter4_experiment_data.json`，再渲染所有 PNG。

| 论文结果 | 负责生成或测量的代码 | JSON 数据字段 | 最终输出 |
| --- | --- | --- | --- |
| 表 5 默认参数 | `chapter4_all_figures.py` 中的 `PARAMS`、`lognormal_stakes()`、`make_context()` | `metadata.params`、`context_summary` | README 默认参数表；`visualization/chapter4_experiment_data.json` |
| 表 6 加权选举方案对比 | `fig14()` 中的 `linear_wr`、`linear_wor`、`uniform_wor`、`ctwr_single`、`ctwr_opt` 五组曲线；`capture_uniform_wor()`、`capture_weighted_wor()`、`optimal_ctwr_split()` 给出对应机制 | `figures.fig14.series` | README 图 14 说明；`visualization/正确图片输出/fig14.png` |
| 图 14 CTWR 捕获概率 | `fig14()` 扫描 `rho` 与注册成本 `c_reg`；`render_fig14()` 出图 | `figures.fig14` | `visualization/正确图片输出/fig14.png` |
| 图 15 委员会规模影响 | `fig15()` 固定 `rho=0.35` 计算 `rho_beacon`、`rho_eff` 与 `N=10..60` 的捕获概率；`render_fig15()` 出图 | `figures.fig15` | `visualization/正确图片输出/fig15.png` |
| 图 16 拆分身份影响 | `split_case()` 计算给定 `rho` 下不同 `k` 的 Linear-WR、Uniform-WoR、CTWR 捕获概率；`fig16_17()` 生成 `rho=0.20/0.35/0.50` 三组；`render_fig16()` 出图 | `figures.fig16_17` | `visualization/正确图片输出/fig16.png` |
| 图 17 CTWR 有效权重 | `fig16_17()` 复用 `rho=0.35` 的 `rho_eff` 和 `w_adv_eff` trace；`render_fig17()` 出图 | `figures.fig16_17["0.35"]` | `visualization/正确图片输出/fig17.png` |
| 图 18 女巫身份与收益 | `incentive_context()` 固定诚实方/对手预算；`fig18()` 计算 CTWR 单位收益、对手净利润、`k*` 与盈亏平衡点；`render_fig18()` 出图 | `figures.fig18` | `visualization/正确图片输出/fig18.png` |
| 图 19 最优女巫策略 | `best_fixed_budget_split()` 扫描注册成本和截断上限；`fig19()` 生成 `k_by_s_cap`、`revenue_by_s_cap` 和热力图；`render_fig19()` 出图 | `figures.fig19` | `visualization/正确图片输出/fig19.png` |
| 图 20 诚实验证者期望收益 | `honest_revenue()`、`fig20()` 扫描女巫身份数、激励池 `R_total` 和运营成本 `c_op`；`render_fig20()` 出图 | `figures.fig20` | `visualization/正确图片输出/fig20.png` |
| 图 21 非比例权重规则对比 | `rho_eff_rule()`、`fig21()` 对比 Sqrt-WoR 与 CTWR 的有效占比、捕获概率和收益；`render_fig21()` 出图 | `figures.fig21` | `visualization/正确图片输出/fig21.png` |
| 图 22 匿名质押开销 | `experiment/test/AnoStBenchmarks.t.sol` 触发 7 个 Gas 测试；`run_foundry_gas()` 解析测试 Gas；`experiment/circuits/*.circom`、`scripts/verify_circuits.js`、`circuit_stats()` 给出电路约束和 witness 状态；`fig22()`、`render_fig22()` 出图 | `foundry_gas_runs`、`zk_circuits`、`figures.fig22` | `visualization/正确图片输出/fig22.png` |

## 电路验证结果

`node scripts/verify_circuits.js` 会为三个电路生成样例输入、witness，并运行 `snarkjs wtns check`。当前结果如下。

| 电路 | 对应接口 | 测试内容 | R1CS 约束数 | Witness 状态 |
| --- | --- | --- | ---: | --- |
| `anost_register.circom` | `AnonyReg` | 检查注册承诺、存款承诺、注册 nullifier 与公开输入一致 | `1088` | `PASS` |
| `anost_stake.circom` | `AnonyStake` | 检查存款承诺、质押承诺、找零承诺、余额约束和质押 nullifier | `1328` | `PASS` |
| `anost_credential.circom` | `PresentCred` | 检查注册 Merkle 路径、凭证 nullifier 和质押承诺 | `5817` | `PASS` |

`experiment/scripts/groth16_smoke.js` 是可选 Groth16 本地证明/验证流程。

## 链上 Gas 测试结果

`forge test --gas-report` 会运行 7 个测试，覆盖公开基线、匿名接口和 CTWR 选举。

| Foundry 测试 | 测试内容 | 当前 Gas |
| --- | --- | ---: |
| `testPublicReg` | 公开注册，把 identity commitment 写入 registry | `70931` |
| `testPublicStake` | 公开质押最小基线，计算 stake commitment hash | `5657` |
| `testCandidateDeclare` | 候选声明最小基线，检查最低质押并生成候选承诺 | `5666` |
| `testAnonyReg` | 匿名注册，检查 nullifier、执行 Groth16-like pairing 验证负载并写 registry | `196706` |
| `testAnonyStake` | 匿名质押，检查 nullifier、执行证明验证负载并生成质押/找零承诺 | `152449` |
| `testPresentCred` | 凭证展示，检查 Merkle path、nullifier 和证明校验 | `212746` |
| `testElectCTWR` | 对 64 个候选按截断权重无放回选出 20 人 | `622254` |

图 22 的非匿名项是最小公开基线，不包含完整公开质押系统可能需要的所有状态写入和校验，因此它主要用于说明当前合约 benchmark 下的相对开销。

## 目录结构

- `第四章面向链下计算的验证者安全选举.pdf`：第四章论文正文。
- `experiment/circuits/`：AnoSt 的三个 Circom 电路。
- `experiment/scripts/verify_circuits.js`：生成样例输入和 witness，运行 `snarkjs wtns check`，校验公开输出。
- `experiment/scripts/groth16_smoke.js`：可选 Groth16 本地证明/验证流程。
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
- `chapter4_all_figures.py`：重新执行 Foundry Gas、检查/编译电路、运行 witness 检查、汇总 R1CS 约束数、计算 CTWR 捕获概率与固定预算激励曲线，并生成图 14-22。

命令与产物对应关系：

| 命令 | 实验内容 | 生成或刷新产物 | 支撑论文结果 |
| --- | --- | --- | --- |
| `node scripts/verify_circuits.js` | AnoSt 注册、质押、凭证展示三个电路的 witness 生成与 `snarkjs wtns check` | `experiment/build/witness/*.wtns`、`experiment/build/witness/circuit_verification_summary.json` | 图 22 电路 witness 状态、匿名质押正确性 |
| `forge test --gas-report` | 公开基线、匿名接口、CTWR 选举链上 Gas | Foundry gas report，主脚本汇总为 `foundry_gas_runs` | 图 22 Gas、链上 Gas 测试结果表 |
| `python3 visualization/chapter4_all_figures.py` | CTWR 解析公式、女巫收益、非比例权重、Gas、电路约束和 witness 状态全链路汇总 | `experiment/logs/raw_experiment_log.json`、`visualization/chapter4_experiment_data.json`、`visualization/正确图片输出/fig14.png` 至 `fig22.png` | 图 14-22、表 5、表 6 |

## 可选 Groth16 Smoke Test

该步骤用于确认本地 proving/verifying 流程可跑通。

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

## 图表产物映射

| 结构化字段 | 实验依据 | 输出图片/结果 |
| --- | --- | --- |
| `figures.fig14` | CTWR、Linear-WR、Linear-WoR、Uniform-WoR 捕获概率公式 trace | 图 14 `fig14.png` |
| `figures.fig15` | 固定 `rho=0.35` 的委员会规模扫描 | 图 15 `fig15.png` |
| `figures.fig16_17` | 固定预算拆分身份扫描与 CTWR 有效权重 trace | 图 16 `fig16.png`、图 17 `fig17.png` |
| `figures.fig18` | 固定对手预算下的女巫身份收益函数 | 图 18 `fig18.png` |
| `figures.fig19` | 注册成本和截断上限的整数拆分扫描 | 图 19 `fig19.png` |
| `figures.fig20` | 激励池与运营成本参数扫描 | 图 20 `fig20.png` |
| `figures.fig21` | CTWR 与 Sqrt-WoR 的非比例权重对比 trace | 图 21 `fig21.png` |
| `foundry_gas_runs`、`zk_circuits`、`figures.fig22` | Foundry gas report、Circom witness check、R1CS 约束统计 | 图 22 `fig22.png` |

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

## 验收流程

1. 在 `chapter4/experiment` 运行 `node scripts/verify_circuits.js`，确认三个 Circom 电路 witness 检查均为 `PASS`。
2. 运行 `forge test --gas-report`，确认公开基线、匿名接口和 CTWR 选举的 Gas 可重新测量。
3. 回到 `chapter4` 运行 `python3 visualization/chapter4_all_figures.py`，重新生成 raw log、结构化 JSON 和图 14-22。
4. 用“快速检查数据”中的 `jq` 命令查看电路约束、Foundry Gas 和图 14/18/22 的结构化字段。
5. 回到仓库根目录运行 `python3 scripts/check_experiment_coverage.py`，检查第四章产物是否与 raw log、结构化数据、图片一致。
