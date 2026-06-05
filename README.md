# paper-exp

`paper-exp` 保存论文第三、四、五章的实验工程、原始实验日志、结构化可视化数据和论文图片。根目录 README 作为仓库地图；具体复现实验命令请进入各章 README。

本仓库包含一个总覆盖检查脚本，用于确认论文实验图表和表格都有可运行代码、raw log 和结构化实验产物支撑：

```bash
python3 scripts/check_experiment_coverage.py
```

该脚本会检查第三章图 7-12、第四章图 14-22、第五章图 24-31/表 9-11，并验证关键结果是否能从实验入口生成的 raw log 对齐到结构化 JSON 和最终图片。

## 验收流程

建议按章节复现，再运行总检查：

1. 进入 `chapter3`、`chapter4`、`chapter5`，按各章 README 的“运行完整实验”执行测试、benchmark、raw log 生成和绘图。
2. 查看各章 `experiment/logs/raw_experiment_log.json`，确认实验入口重新写入了原始日志。
3. 查看各章 `visualization/*_experiment_data.json` 与 `visualization/正确图片输出/`，确认结构化数据和图片已刷新。
4. 回到仓库根目录运行 `python3 scripts/check_experiment_coverage.py`，做跨章节一致性检查。

## 章节索引

- `chapter3`：第三章“基于有状态任务切片的链下计算验证”，覆盖 CleVer 切片、争议定位、伴随式验证、Geth EVM 采样和 Foundry Gas 对比。
- `chapter4`：第四章“面向链下计算的验证者安全选举”，覆盖 AnoSt 匿名质押电路、CTWR 选举、女巫收益分析、Circom/snarkjs 检查和 Foundry Gas 基准。
- `chapter5`：第五章“面向链下计算的验证者工作审计”，覆盖 RanCk 在线跟踪审计、SenCk 哨兵抽样勤勉检测、旁路审计开销和链上 Gas 基准。

## 目录结构

```text
paper-exp/
  README.md
  scripts/
    check_experiment_coverage.py
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

## 产物位置

| 章节 | 复现说明 | Raw log | Structured data | 图片输出 |
| --- | --- | --- | --- | --- |
| chapter3 | `chapter3/README.md` | `chapter3/experiment/logs/raw_experiment_log.json` | `chapter3/visualization/chapter3_experiment_data.json` | `chapter3/visualization/正确图片输出/` |
| chapter4 | `chapter4/README.md` | `chapter4/experiment/logs/raw_experiment_log.json` | `chapter4/visualization/chapter4_experiment_data.json` | `chapter4/visualization/正确图片输出/` |
| chapter5 | `chapter5/README.md` | `chapter5/experiment/logs/raw_experiment_log.json` | `chapter5/visualization/chapter5_experiment_data.json` | `chapter5/visualization/正确图片输出/` |

第三章和第五章还提供独立对比实验报告，便于验收时单独查看对比项如何由本地 benchmark 或场景 trace 生成：

- `chapter3/experiment/logs/comparison_protocols.json`：Arbitrum Classic、TrueBit、Cartesi Dave、Arbitrum BoLD、CleVer 的 Foundry 对比基准实验报告。
- `chapter5/experiment/logs/comparison_experiments.json`：TrueBit、Arbitrum、PoD、RanCk+SenCk 的场景对比实验和 PoD gas baseline 报告。

## 依赖概览

- 第三章：Go、Foundry、Geth `evm`、Python、`matplotlib`、`numpy`。
- 第四章：Node.js/npm、Circom、snarkjs/circomlib、Foundry、Python、`matplotlib`、`numpy`、`Pillow`。
- 第五章：Go、Foundry、Python、`matplotlib`、`numpy`。

如果 Go 或 Matplotlib 默认缓存目录不可写，可在运行前设置本地临时缓存目录：

```bash
export GOCACHE="${TMPDIR:-/tmp}/paper-exp-go-cache"
export MPLCONFIGDIR="${TMPDIR:-/tmp}/paper-exp-mplconfig"
mkdir -p "$GOCACHE" "$MPLCONFIGDIR"
```

## 不提交的可重建产物

以下目录或文件由本地构建、测试或绘图过程生成，不应提交：

- `node_modules/`
- `build/`
- `out/`
- `cache/`
- `__pycache__/`
- `*.pyc`
- `chapter4/experiment/tools/bin/circom`
- 本地临时缓存目录，例如 `${TMPDIR:-/tmp}/paper-exp-*`

`.gitignore` 已覆盖常见构建产物。raw log、structured data 和 PNG 图片是论文实验复现产物，可按需要提交。
