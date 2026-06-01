# 第三章 CleVer 实验工程

本目录用于复现实验数据来源，而不是直接硬编码可视化数据。

- `src/BenchmarkTasks.sol`：四类基准任务合约，对应 Fibonacci、Poly-Chain、Sort-Large、DP-Large。
- `src/CleVerVerifier.sol`：最小裁决单元 `VerSeg` 的链上重放裁决接口。
- `test/Benchmarks.t.sol`：Foundry Gas 测试入口。
- `cmd/clever-exp/main.go`：真实实验采集程序，实际执行四类任务的小规模样本，输出快照序列化大小、状态承诺、Geth `evm --bench run` 字节码执行结果，以及 Arbitrum/TrueBit/Cartesi/BoLD/CleVer 横向争议协议对比日志。

推荐命令：

```bash
forge test --gas-report
go run ./cmd/clever-exp --quick --out logs/raw_experiment_log.json
python3 ../第三章可视化/generate_chapter3_data.py
python3 ../第三章可视化/chapter3_all_figures.py
```

长任务使用日志模式，不强制跑完 10 分钟/小时级实验：

```bash
go run ./cmd/clever-exp --full --max-seconds 5 --out logs/raw_experiment_log.json
```
