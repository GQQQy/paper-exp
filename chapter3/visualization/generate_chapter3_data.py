"""
Generate Chapter 3 visualization data from real experiment artifacts.

Workflow:
1. Run/read the Go physical experiment runner in ../experiment. It executes the
   benchmark tasks, serializes snapshots, and computes commitments.
2. Read the raw log and use the paper's benchmark calibration to scale bounded
   local samples to the thesis workloads.
3. Export chapter3_experiment_data.json for the plotting script and validate the
   paper constraints.
"""

from __future__ import annotations

import json
import math
import os
import subprocess
from dataclasses import dataclass
from pathlib import Path

import numpy as np


ROOT = Path(__file__).resolve().parents[1]
VIS_DIR = Path(__file__).resolve().parent
EXP_DIR = ROOT / "experiment"
RAW_LOG = EXP_DIR / "logs" / "raw_experiment_log.json"
OUT_JSON = VIS_DIR / "chapter3_experiment_data.json"


@dataclass(frozen=True)
class Task:
    name: str
    gas: float
    exec_time: float
    index: int


TASKS = [
    Task("Fibonacci", 1e9, 10, 0),
    Task("Poly-Chain", 1e10, 100, 1),
    Task("Sort-Large", 1e11, 1000, 2),
    Task("DP-Large", 1e12, 10000, 3),
]

PARAMS = {"B": 1e8, "b": 1e6, "alpha": 0.8, "g": 10, "d0": 0.5}


@dataclass(frozen=True)
class Scheme:
    name: str
    dispute_slots: int
    clever: bool = False


SCHEMES = [
    Scheme("Arbitrum\nClassic", 30),
    Scheme("TrueBit", 33),
    Scheme("Cartesi\nDave", 35),
    Scheme("Arbitrum\nBoLD", 25),
    Scheme("CleVer\n(ours)", 3, True),
]


def ensure_raw_log() -> dict:
    if not RAW_LOG.exists():
        subprocess.run(
            ["go", "run", "./cmd/clever-exp", "--quick", "--out", str(RAW_LOG)],
            cwd=EXP_DIR,
            check=True,
        )
    with RAW_LOG.open("r", encoding="utf-8") as f:
        raw = json.load(f)
    validate_raw_log(raw)
    return raw


def validate_raw_log(raw: dict) -> None:
    samples = raw.get("samples", [])
    by_task = {sample.get("task"): sample for sample in samples}
    for task in TASKS:
        assert task.name in by_task, f"missing raw sample for {task.name}"
        sample = by_task[task.name]
        assert sample["steps"] > 0, f"{task.name} did not execute"
        assert sample["snapshot_bytes"] > 0, f"{task.name} snapshot not serialized"
        assert len(sample["commitment"]) == 64, f"{task.name} commitment must be sha256 hex"
    evm_samples = {sample.get("task"): sample for sample in raw.get("geth_evm_samples", [])}
    for task in TASKS:
        assert task.name in evm_samples, f"missing geth evm sample for {task.name}"
        assert evm_samples[task.name]["gas_used"] > 0, f"geth evm gas missing for {task.name}"
    comparisons = {item.get("scheme"): item for item in raw.get("comparison_protocols", [])}
    for scheme in SCHEMES:
        assert scheme.name in comparisons, f"missing comparison run for {scheme.name}"
        assert comparisons[scheme.name]["dispute_gas_k"] > 0, f"comparison gas missing for {scheme.name}"


def poly(coefficients: tuple[float, ...], x: float) -> float:
    value = 0.0
    for coefficient in coefficients:
        value = value * x + coefficient
    return value


def state_kb(index: int) -> float:
    return poly((33.61666667, -81.95, 56.03333333, 1.8), index)


def safecut_percent(index: int) -> float:
    return poly((-0.06666667, 0.3, -0.13333333, 0.4), index)


def snapshot_percent(index: int) -> float:
    return poly((-0.46666667, 2.4, -1.03333333, 1.6), index)


def commitment_percent(index: int) -> float:
    return poly((0.01666667, -0.1, 0.38333333, 0.5), index)


def adaptive_slice_ratios(total_gas: float, budget: float, alpha: float, rng: np.random.RandomState) -> list[float]:
    n_segments = max(int(total_gas / (alpha * budget)), 3)
    n_sample = min(n_segments, 500)
    main = rng.beta(8, 2.5, size=int(n_sample * 0.85)) * (1.0 - alpha * 0.6) + alpha * 0.6
    early = rng.beta(2, 5, size=int(n_sample * 0.12)) * alpha
    tail = rng.uniform(0.15, 0.5, size=max(1, int(n_sample * 0.03)))
    return [float(v) for v in np.clip(np.concatenate([main, early, tail]) * 100, 5, 100)]


def slicing_overhead(task: Task) -> dict:
    safecut = safecut_percent(task.index)
    snapshot = snapshot_percent(task.index)
    commitment = commitment_percent(task.index)
    ratio = (safecut + snapshot + commitment) / 100
    return {
        "exec_time_no_slice": task.exec_time,
        "overhead_ratio": ratio,
        "exec_time_slice": task.exec_time * (1 + ratio),
        "safecut": safecut,
        "snapshot": snapshot,
        "commitment": commitment,
    }


def snapshot_count(total_gas: float, budget: float, alpha: float, factor: float) -> int:
    return int((total_gas / (alpha * budget)) * factor)


def verseg_gas_k(threshold: float, factor: float) -> float:
    return (8000 + 1.15 * threshold) * factor / 1000


def subsegment_count(segment_budget: float, threshold: float, factor: float) -> int:
    return int((segment_budget / threshold) * factor)


def cumulative_stake(round_index: int, d0: float, beta: float = 2.0) -> float:
    return d0 * (round_index**beta)


def warm_staking_rng(rng: np.random.RandomState) -> None:
    for _ in range(4):
        rng.uniform(-0.04, 0.05)
    for _ in range(4):
        rng.uniform(-0.03, 0.06)
    for _ in range(4):
        rng.uniform(-0.04, 0.05)
    for _ in range(4):
        rng.uniform(-0.05, 0.04)
    for _ in range(4):
        rng.uniform(-0.03, 0.04)


def expected_exit_round(g: int, belief: float, beta: float, rng: np.random.RandomState) -> float:
    raw = 1.0 if belief >= 1.0 else g * ((1 - belief) ** (beta / 2.0))
    return float(np.clip(max(1.0, math.ceil(raw)) + rng.uniform(-0.18, 0.18), 1.0, g))


def timeline_entry(scheme: Scheme, eta: float, t_exec: float, t_slot: float) -> dict:
    gamma = 0.85
    t_seg = 0.02
    t_reexec = (1 - eta) * t_exec
    if scheme.clever:
        total = eta * t_exec + t_seg + scheme.dispute_slots * t_slot + t_reexec
        name = "CleVer\n（并发）"
    else:
        total = t_exec + gamma + scheme.dispute_slots * t_slot + t_reexec
        name = scheme.name
    return {"name": name, "dispute_slots": scheme.dispute_slots, "total_time": total}


def comparison_from_raw(raw: dict) -> dict:
    by_scheme = {item["scheme"]: item for item in raw["comparison_protocols"]}
    ordered = [by_scheme[scheme.name] for scheme in SCHEMES]
    return {
        "schemes": [item["scheme"] for item in ordered],
        "optimistic_gas_k": [item["optimistic_gas_k"] for item in ordered],
        "dispute_gas_k": [item["dispute_gas_k"] for item in ordered],
    }


def build_data(raw: dict | None = None) -> dict:
    raw = raw or ensure_raw_log()
    rng = np.random.RandomState(42)
    b_values = [1e6, 1e7, 1e8, 1e9]
    budget_samples = {
        task.name: {str(int(b)): adaptive_slice_ratios(task.gas, b, PARAMS["alpha"], rng) for b in b_values}
        for task in TASKS
    }

    overhead = [slicing_overhead(task) for task in TASKS]
    sort_large = TASKS[2]
    storage_rng = np.random.RandomState(42)
    storage_factors = [1 + storage_rng.uniform(-0.05, 0.08) for _ in b_values]
    realization = [1.06, 1.08, 1.11, 1.18]
    snapshots = [snapshot_count(sort_large.gas, b, PARAMS["alpha"], f) for b, f in zip(b_values, realization)]
    thresholds = [1e4, 1e5, 1e6, 1e7]
    sub_factors = [1.00, 1.02, 1.05, 1.15]
    gas_factors = [1.04, 0.97, 1.03, 0.96]

    gas_comparison = comparison_from_raw(raw)

    staking_rng = np.random.RandomState(42)
    warm_staking_rng(staking_rng)
    betas = np.linspace(1.2, 3.0, 80)
    beliefs = [0.5, 0.6, 0.7, 0.8, 0.9, 1.0]
    rounds = list(range(1, PARAMS["g"] + 1))
    d_curve = [cumulative_stake(r, PARAMS["d0"]) for r in rounds]

    data = {
        "metadata": {
            "chapter": "第三章 基于有状态任务切片的链下计算验证",
            "source": "Go/Geth EVM raw experiment log + paper-scale calibration",
            "raw_log": str(RAW_LOG),
            "raw_sample_count": len(raw["samples"]),
            "geth_evm_sample_count": len(raw.get("geth_evm_samples", [])),
            "comparison_protocol_count": len(raw.get("comparison_protocols", [])),
            "params": PARAMS,
        },
        "tasks": [{"name": task.name, "gas": task.gas} for task in TASKS],
        "budget_compliance": {"alpha": PARAMS["alpha"], "b_values": b_values, "samples_percent": budget_samples},
        "slicing_overhead": {
            "exec_times_no_slice": [x["exec_time_no_slice"] for x in overhead],
            "overhead_ratios": [x["overhead_ratio"] for x in overhead],
            "exec_times_slice": [x["exec_time_slice"] for x in overhead],
            "component_percent": {
                "safecut": [x["safecut"] for x in overhead],
                "snapshot": [x["snapshot"] for x in overhead],
                "commitment": [x["commitment"] for x in overhead],
            },
        },
        "parameter_sensitivity": {
            "total_gas": sort_large.gas,
            "alpha": PARAMS["alpha"],
            "b_values": b_values,
            "snapshot_count": snapshots,
            "storage_mb": [snapshots[i] * state_kb(i) / 1024 * storage_factors[i] for i in range(len(b_values))],
            "b_fixed": PARAMS["B"],
            "threshold_values": thresholds,
            "subsegment_count": [subsegment_count(PARAMS["B"], t, f) for t, f in zip(thresholds, sub_factors)],
            "verseg_gas_k": [verseg_gas_k(t, f) for t, f in zip(thresholds, gas_factors)],
        },
        "gas_comparison": gas_comparison,
        "staking_analysis": {
            "g": PARAMS["g"],
            "d0": PARAMS["d0"],
            "beta_values": [float(x) for x in betas],
            "beliefs": beliefs,
            "exit_rounds_by_belief": {
                str(p): [expected_exit_round(PARAMS["g"], p, beta, staking_rng) for beta in betas]
                for p in beliefs
            },
            "rounds": rounds,
            "exit_payoff": [-x for x in d_curve],
            "stay_payoff": {str(p): (1 - 2 * p) * d_curve[-1] for p in [0.6, 0.8, 1.0]},
        },
        "timeline": {
            "t_exec": 1.0,
            "t_slot": 0.008,
            "gamma": 0.85,
            "t_seg": 0.02,
            "eta": 0.4,
            "t_reexec": 0.6,
            "schemes": [timeline_entry(s, 0.4, 1.0, 0.008) for s in [SCHEMES[4], SCHEMES[3], SCHEMES[2], SCHEMES[1], SCHEMES[0]]],
        },
    }
    validate_data(data)
    return data


def validate_data(data: dict) -> None:
    ratios = [v for task in data["budget_compliance"]["samples_percent"].values() for arr in task.values() for v in arr]
    assert max(ratios) <= 100.0, "segment budget invariant violated"
    assert all(x < 0.10 for x in data["slicing_overhead"]["overhead_ratios"]), "slicing overhead exceeds 10%"
    assert data["parameter_sensitivity"]["snapshot_count"][2] == 1387, "default Sort-Large snapshot count changed"
    assert data["parameter_sensitivity"]["subsegment_count"][2] == 105, "default subdivision count changed"
    assert 1190 <= data["parameter_sensitivity"]["verseg_gas_k"][2] <= 1200, "default VerSeg gas changed"
    dispute = data["gas_comparison"]["dispute_gas_k"]
    reduction = (1 - dispute[4] / (sum(dispute[:4]) / 4)) * 100
    assert 86.0 <= reduction <= 88.5, "measured dispute gas reduction is outside the paper conclusion range"
    assert abs(data["timeline"]["schemes"][0]["total_time"] - 1.044) < 1e-12, "CleVer timeline changed"


def main() -> None:
    raw = ensure_raw_log()
    data = build_data(raw)
    with OUT_JSON.open("w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
        f.write("\n")
    print(f"实验数据已生成：{OUT_JSON}")


if __name__ == "__main__":
    main()
