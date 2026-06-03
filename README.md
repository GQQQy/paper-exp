# paper-exp 实验复现说明

`paper-exp` 用于保存博士论文第三、四、五章的实验工程、原始日志、结构化可视化数据、论文图表和论文对齐报告。仓库目标是让实验链路可追踪：实验代码或公式/仿真先生成 raw log，再生成 structured data、图片和 `paper_alignment_report.md`。

本仓库不生成额外 audit report。

## 章节内容

- `chapter3`：对应论文第三章“基于有状态任务切片的链下计算验证”。实验覆盖 CleVer 自适应切片、争议定位解决效率、伴随式验证时效性、Geth EVM 执行采样、Foundry Gas 对比和图 7-12。
- `chapter4`：对应论文第四章“面向链下计算的验证者安全选举”。实验覆盖 AnoSt 匿名准入电路、Circom/snarkjs witness 检查、CTWR 安全选举、女巫攻击收益仿真、Foundry Gas 基准和图 14-22。
- `chapter5`：对应论文第五章“面向链下计算的验证者工作审计”。实验覆盖 RanCk 在线跟踪审计、SenCk 哨兵抽样勤勉检测、激励可行域、检测概率、旁路审计开销、Foundry/Table 10 Gas 校准和图 24-31。

## 总目录结构

```text
paper-exp/
  README.md
  chapter3/
    README.md
    experiment/
    visualization/
    第三章基于有状态任务切片的链下计算验证.pdf
  chapter4/
    README.md
    experiment/
    visualization/
    第四章面向链下计算的验证者安全选举.pdf
  chapter5/
    README.md
    experiment/
    visualization/
    第五章面向链下计算的验证者工作审计.pdf
```

各章的 `experiment/README.md` 是实验工程入口说明；各章根目录 `README.md` 是更完整的复现指南。

## 环境依赖总览

通用依赖：

- Python 3.10 或兼容版本。
- Python 包：`matplotlib`、`numpy`；第四章还需要 `Pillow`。
- Foundry，包括 `forge`。

章节特定依赖：

- 第三章：Go 1.21+，Geth 的 `evm` 命令。
- 第四章：Node.js 18+，npm，Circom 2.1.8，snarkjs/circomlib 依赖。
- 第五章：Go 1.21+。

检查命令：

```bash
go version
forge --version
evm --help
node --version
npm --version
python3 --version
python3 -c "import matplotlib, numpy, PIL; print('python deps ok')"
```

如果网络不通，安装依赖前可使用代理：

```bash
export https_proxy=http://127.0.0.1:33210 http_proxy=http://127.0.0.1:33210 all_proxy=socks5://127.0.0.1:33211
```

在 macOS 沙盒环境中，Go、Matplotlib 或 Foundry 可能需要可写缓存目录：

```bash
export GOCACHE=/private/tmp/go-build-paper-exp
export MPLCONFIGDIR=/private/tmp/mplconfig-paper-exp
mkdir -p "$GOCACHE" "$MPLCONFIGDIR"
```

## 统一复现实验顺序

建议按第三章、第四章、第五章顺序复现，因为第四、五章的叙述都复用第三章的链下计算验证背景。

1. 第三章：

```bash
cd chapter3/experiment
go test ./...
forge test --gas-report
go run ./cmd/clever-exp --quick --out logs/raw_experiment_log.json

cd ../..
python3 chapter3/visualization/generate_chapter3_data.py
python3 chapter3/visualization/chapter3_all_figures.py
python3 chapter3/visualization/verify_paper_alignment.py
```

2. 第四章：

```bash
cd chapter4/experiment
npm install
forge test --gas-report

cd ../..
python3 chapter4/visualization/chapter4_all_figures.py
```

`chapter4_all_figures.py` 会自动检查/编译 Circom 电路，运行 `node scripts/verify_circuits.js`，采集 Foundry Gas，并生成 raw log、结构化数据、图片和对齐报告。

3. 第五章：

```bash
cd chapter5/experiment
GOCACHE=/private/tmp/chapter5-gocache go test ./...
forge test --gas-report
GOCACHE=/private/tmp/chapter5-gocache go run ./cmd/audit-exp --out logs/raw_experiment_log.json

cd ../..
MPLCONFIGDIR=/private/tmp/mplconfig_ch5 python3 chapter5/visualization/chapter5_all_figures.py
```

## 输出位置

| 章节 | README | Raw log | Structured data | 图片输出 | paper_alignment_report |
| --- | --- | --- | --- | --- | --- |
| chapter3 | `chapter3/README.md`、`chapter3/experiment/README.md` | `chapter3/experiment/logs/raw_experiment_log.json` | `chapter3/visualization/chapter3_experiment_data.json` | `chapter3/visualization/正确图片输出/` | `chapter3/visualization/paper_alignment_report.md` |
| chapter4 | `chapter4/README.md`、`chapter4/experiment/README.md` | `chapter4/experiment/logs/raw_experiment_log.json` | `chapter4/visualization/chapter4_experiment_data.json` | `chapter4/visualization/正确图片输出/` | `chapter4/visualization/paper_alignment_report.md` |
| chapter5 | `chapter5/README.md`、`chapter5/experiment/README.md` | `chapter5/experiment/logs/raw_experiment_log.json` | `chapter5/visualization/chapter5_experiment_data.json` | `chapter5/visualization/正确图片输出/` | `chapter5/visualization/paper_alignment_report.md` |

## 图片文件

- 第三章图 7-12：
  `fig1_budget_compliance.png`、`fig2_overhead.png`、`fig_param_sensitivity_v2.png`、`fig_gas_comparison_v2.png`、`fig_staking_analysis_v2.png`、`fig_timeline_v3.png`
- 第四章图 14-22：
  `fig14.png` 到 `fig22.png`
- 第五章图 24-31：
  `fig24_feasibility_ab.png`、`fig25_joint_feasibility.png`、`fig26_ranck_detection.png`、`fig27_senck_passthrough.png`、`fig28_monte_carlo_and_gate.png`、`fig29_overhead.png`、`fig30_gas_comparison.png`、`fig31_tradeoff.png`

## 不应提交的可重建产物

以下目录或文件由本地构建、测试或绘图过程生成，不应提交：

- `node_modules/`
- `build/`
- `out/`
- `cache/`
- `__pycache__/`
- `*.pyc`
- `chapter4/experiment/tools/bin/circom`
- 本地临时缓存目录，例如 `/private/tmp/go-build-*`、`/private/tmp/mplconfig_*`

仓库 `.gitignore` 已覆盖上述常见构建产物。若重新生成 raw log、structured data、PNG 图片或 `paper_alignment_report.md`，这些是论文复现实验产物，可以按需要提交。

## 章节指南

更多细节请从各章 README 开始：

- `chapter3/README.md`
- `chapter4/README.md`
- `chapter5/README.md`
