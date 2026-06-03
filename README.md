# paper-exp

`paper-exp` 保存论文第三、四、五章实验工程、实验日志、结构化可视化数据、论文图片和论文对齐报告。根目录 README 只作为仓库地图；具体复现实验命令、依赖安装和注意事项请进入各章 README。

本仓库不生成额外 audit report；各章的 `paper_alignment_report.md` 仅用于说明论文图表与实验数据链路是否对齐。

## 章节索引

- `chapter3`：第三章“基于有状态任务切片的链下计算验证”，覆盖 CleVer 切片、争议定位、伴随式验证、Geth EVM 采样和 Foundry Gas 对比。
- `chapter4`：第四章“面向链下计算的验证者安全选举”，覆盖 AnoSt 匿名质押电路、CTWR 选举、女巫收益分析、Circom/snarkjs 检查和 Foundry Gas 基准。
- `chapter5`：第五章“面向链下计算的验证者工作审计”，覆盖 RanCk 在线跟踪审计、SenCk 哨兵抽样勤勉检测、旁路审计开销和链上 Gas 基准。

## 目录结构

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

## 产物位置

| 章节 | 复现说明 | Raw log | Structured data | 图片输出 | 对齐报告 |
| --- | --- | --- | --- | --- | --- |
| chapter3 | `chapter3/README.md` | `chapter3/experiment/logs/raw_experiment_log.json` | `chapter3/visualization/chapter3_experiment_data.json` | `chapter3/visualization/正确图片输出/` | `chapter3/visualization/paper_alignment_report.md` |
| chapter4 | `chapter4/README.md` | `chapter4/experiment/logs/raw_experiment_log.json` | `chapter4/visualization/chapter4_experiment_data.json` | `chapter4/visualization/正确图片输出/` | `chapter4/visualization/paper_alignment_report.md` |
| chapter5 | `chapter5/README.md` | `chapter5/experiment/logs/raw_experiment_log.json` | `chapter5/visualization/chapter5_experiment_data.json` | `chapter5/visualization/正确图片输出/` | `chapter5/visualization/paper_alignment_report.md` |

## 依赖概览

- 第三章：Go、Foundry、Geth `evm`、Python、`matplotlib`、`numpy`。
- 第四章：Node.js/npm、Circom、snarkjs/circomlib、Foundry、Python、`matplotlib`、`numpy`、`Pillow`。
- 第五章：Go、Foundry、Python、`matplotlib`、`numpy`。

如果网络不通，安装依赖前可按仓库约定使用代理：

```bash
export https_proxy=http://127.0.0.1:33210 http_proxy=http://127.0.0.1:33210 all_proxy=socks5://127.0.0.1:33211
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
- 本地临时缓存目录，例如 `/private/tmp/go-build-*`、`/private/tmp/mplconfig_*`

`.gitignore` 已覆盖常见构建产物。raw log、structured data、PNG 图片和 `paper_alignment_report.md` 是论文复现实验产物，可按需要提交。
