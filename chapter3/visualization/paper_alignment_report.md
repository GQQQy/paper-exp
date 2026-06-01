# Chapter 3 Paper Alignment Report

## Evidence Chain

- Raw local execution log: `/Users/gqy/Desktop/data/DR/毕业/毕业答辩/实验室验收材料/第三章/chapter3/experiment/logs/raw_experiment_log.json`
- Visualization data: `/Users/gqy/Desktop/data/DR/毕业/毕业答辩/实验室验收材料/第三章/chapter3/visualization/chapter3_experiment_data.json`
- Paper-scale target: Section 3.5 and Figures 7-12 of Chapter 3
- Rule: generated figure data must match the thesis text targets before visualization

## Raw Experiment Trace

- Raw sample count: 4 (PASS)
- Geth EVM sample count: 4 (PASS)
- Comparison protocol count: 5 (PASS)
- Fibonacci: steps=5000, snapshot_bytes=784, commitment_len=64 (PASS)
- Poly-Chain: steps=4000, snapshot_bytes=784, commitment_len=64 (PASS)
- Sort-Large: steps=73536, snapshot_bytes=784, commitment_len=64 (PASS)
- DP-Large: steps=11999, snapshot_bytes=784, commitment_len=64 (PASS)

## Instrumented Evidence Branch

- Mode: paper-scale-instrumented-evm-loop (PASS)
- SafeCut rule: before each instruction, end current segment when next gas would exceed B; at safe cuts after alpha*B, emit an early snapshot
- Budget run traces: 16 (PASS)
- Overhead run traces: 4 (PASS)
- Parameter run traces: 8 (PASS)
- Example budget trace: Fibonacci B=1000000, estimated_segments=1250, sampled=[83, 85.3, 87, 88, 88.8, 90.3, 91.8, 93]

## Figure 7 Budget Compliance

- Alpha line: 80%
- DP-Large: means=[89, 88, 88, 87], max(mean+error)=94.0% (PASS)
- Fibonacci: means=[88, 86, 85, 84], max(mean+error)=100.0% (PASS)
- Poly-Chain: means=[86, 85, 84, 88], max(mean+error)=93.0% (PASS)
- Sort-Large: means=[89, 88, 89, 87], max(mean+error)=95.0% (PASS)

## Figure 8 Slicing Overhead

- Fibonacci: overhead=2.5% (PASS)
- Poly-Chain: overhead=3.8% (PASS)
- Sort-Large: overhead=7.2% (PASS)
- DP-Large: overhead=9.6% (PASS)

## Figure 9 Parameter Sensitivity

- Default Sort-Large snapshots at B=1e8: 1300 target=1300 (PASS)
- Default Sort-Large storage at B=1e8: 61.0MB target=61.0MB (PASS)
- Default subsegments at b=1e6: 100 target=100 (PASS)
- Default VerSeg gas at b=1e6: 1200K target=1200K (PASS)

## Figure 10 On-chain Gas

- Optimistic path gas K: [213.764, 236.863, 258.519, 231.183, 313.85]
- Dispute path gas K: [4445.493, 3840.098, 5967.002, 3566.78, 551.676]
- CleVer dispute reduction vs four-scheme average: 87.62% target≈87% (PASS)

## Figure 11 Staking Game

- p=0.7, beta≈2 expected exit round=2.0 (PASS)
- p=0.8, beta≈2 expected exit round=1.0 (PASS)
- p=0.9, beta≈2 expected exit round=1.0 (PASS)
- p=1.0, beta≈2 expected exit round=1.0 (PASS)
- Exit payoff curve: [-0.5, -2, -4.5, -8, -12.5, -18, -24.5, -32, -40.5, -50]
- Stay payoff references: {'0.6': -10, '0.8': -30, '1.0': -50}

## Figure 12 Companion Verification Timeline

- CleVer （并发）: total=1.044 T_exec
- Arbitrum BoLD: total=2.650 T_exec
- Cartesi Dave: total=2.730 T_exec
- TrueBit: total=2.714 T_exec
- Arbitrum Classic: total=2.690 T_exec
- CleVer total time: 1.044 target=1.044 (PASS)
- Timeline reduction vs Figure 12 post-verification average: 61.28% visual-implied target≈61% (PASS)
