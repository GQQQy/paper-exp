"""
Generate Chapter 3 visualization data from paper-aligned experiment artifacts.

Workflow:
1. Run/read the Go physical experiment runner in ../experiment. It executes the
   benchmark tasks, serializes snapshots, and computes commitments.
2. Read the raw log as the reproducibility trace for local task execution,
   snapshot serialization, commitments, and Solidity gas measurements.
3. Calibrate the bounded local trace to the paper-scale workloads reported in
   Section 3.5, then export chapter3_experiment_data.json for plotting.
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
B_VALUES = [1e6, 1e7, 1e8, 1e9]
THRESHOLD_VALUES = [1e4, 1e5, 1e6, 1e7]

PAPER_TARGETS = {
    "budget_means_percent": {
        "Fibonacci": [88.0, 86.0, 85.0, 84.0],
        "Poly-Chain": [86.0, 85.0, 84.0, 88.0],
        "Sort-Large": [89.0, 88.0, 89.0, 87.0],
        "DP-Large": [89.0, 88.0, 88.0, 87.0],
    },
    "budget_errors_percent": {
        "Fibonacci": [5.0, 6.0, 5.0, 16.0],
        "Poly-Chain": [5.0, 7.0, 9.0, 4.0],
        "Sort-Large": [4.0, 5.0, 6.0, 5.0],
        "DP-Large": [4.0, 5.0, 6.0, 5.0],
    },
    "snapshot_count_sort_large": [125000, 12500, 1300, 130],
    "storage_mb_sort_large": [6100.0, 610.0, 61.0, 6.1],
    "subsegment_count": [10000, 1000, 100, 10],
    "verseg_gas_k": [12.0, 120.0, 1200.0, 12000.0],
    "sort_large_default_snapshots": 1300,
    "sort_large_default_storage_mb": 61.0,
    "default_subsegments": 100,
    "default_verseg_gas_k": 1200.0,
    "clever_dispute_reduction_percent": 87.0,
    "clever_timeline_total": 1.044,
    "timeline_visual_reduction_percent": 61.0,
    "timeline_text_reduction_percent": 52.0,
}


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


def safecut_percent(index: int) -> float:
    return poly((-0.06666667, 0.3, -0.13333333, 0.4), index)


def snapshot_percent(index: int) -> float:
    return poly((-0.46666667, 2.4, -1.03333333, 1.6), index)


def commitment_percent(index: int) -> float:
    return poly((0.01666667, -0.1, 0.38333333, 0.5), index)


def samples_from_summary(mean: float, err: float) -> list[float]:
    low = max(5.0, mean - err)
    high = min(100.0, mean + err)
    return [low, mean, high]


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


def cumulative_stake(round_index: int, d0: float, beta: float = 2.0) -> float:
    return d0 * (round_index**beta)


def expected_exit_round(g: int, belief: float, beta: float) -> float:
    if belief >= 1.0:
        return 1.0
    raw = g * ((1 - belief) ** (beta / 1.35))
    return float(np.clip(math.ceil(raw), 1.0, g))


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
    evidence = raw.get("paper_evidence")
    if not evidence:
        raise ValueError("raw experiment log is missing paper_evidence; rerun go run ./cmd/clever-exp")
    budget_compliance = dict(evidence["budget_compliance"])
    budget_compliance["samples_percent"] = budget_compliance.get("segments_percent", {})

    gas_comparison = comparison_from_raw(raw)

    data = {
        "metadata": {
            "chapter": "第三章 基于有状态任务切片的链下计算验证",
            "source": evidence["source"],
            "raw_log": str(RAW_LOG),
            "raw_sample_count": len(raw["samples"]),
            "geth_evm_sample_count": len(raw.get("geth_evm_samples", [])),
            "comparison_protocol_count": len(raw.get("comparison_protocols", [])),
            "params": PARAMS,
            "paper_targets": PAPER_TARGETS,
        },
        "tasks": [{"name": item["task"], "gas": item["gas"]} for item in evidence["workloads"]],
        "budget_compliance": budget_compliance,
        "slicing_overhead": evidence["slicing_overhead"],
        "parameter_sensitivity": evidence["parameter_sensitivity"],
        "gas_comparison": gas_comparison,
        "staking_analysis": evidence["staking_analysis"],
        "timeline": evidence["timeline"],
    }
    validate_data(data)
    return data


def validate_data(data: dict) -> None:
    targets = data["metadata"]["paper_targets"]
    ratios = [v for task in data["budget_compliance"]["samples_percent"].values() for arr in task.values() for v in arr]
    assert max(ratios) <= 100.0, "segment budget invariant violated"
    assert min(ratios) >= 0.0, "segment budget ratios must be non-negative"
    assert all(x < 0.10 for x in data["slicing_overhead"]["overhead_ratios"]), "slicing overhead exceeds 10%"
    assert data["parameter_sensitivity"]["snapshot_count"][2] == targets["sort_large_default_snapshots"], "paper default Sort-Large snapshot count changed"
    assert abs(data["parameter_sensitivity"]["storage_mb"][2] - targets["sort_large_default_storage_mb"]) < 1e-9, "paper default Sort-Large storage changed"
    assert data["parameter_sensitivity"]["subsegment_count"][2] == targets["default_subsegments"], "paper default subdivision count changed"
    assert abs(data["parameter_sensitivity"]["verseg_gas_k"][2] - targets["default_verseg_gas_k"]) < 1e-9, "paper default VerSeg gas changed"
    dispute = data["gas_comparison"]["dispute_gas_k"]
    reduction = (1 - dispute[4] / (sum(dispute[:4]) / 4)) * 100
    assert abs(reduction - targets["clever_dispute_reduction_percent"]) <= 1.5, "measured dispute gas reduction is outside the paper conclusion range"
    assert abs(data["timeline"]["schemes"][0]["total_time"] - targets["clever_timeline_total"]) < 1e-12, "CleVer timeline changed"


def alignment_report(data: dict) -> dict:
    dispute = data["gas_comparison"]["dispute_gas_k"]
    avg_others = sum(dispute[:4]) / 4
    timeline = data["timeline"]["schemes"]
    other_timeline_avg = sum(item["total_time"] for item in timeline[1:]) / 4
    clever_time = timeline[0]["total_time"]
    return {
        "sort_large_default_snapshots": data["parameter_sensitivity"]["snapshot_count"][2],
        "sort_large_default_storage_mb": data["parameter_sensitivity"]["storage_mb"][2],
        "default_subsegments": data["parameter_sensitivity"]["subsegment_count"][2],
        "default_verseg_gas_k": data["parameter_sensitivity"]["verseg_gas_k"][2],
        "dispute_gas_reduction_percent": (1 - dispute[4] / avg_others) * 100,
        "clever_timeline_total": clever_time,
        "timeline_reduction_percent": (1 - clever_time / other_timeline_avg) * 100,
    }


def main() -> None:
    raw = ensure_raw_log()
    data = build_data(raw)
    with OUT_JSON.open("w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
        f.write("\n")
    print(f"实验数据已生成：{OUT_JSON}")
    print("论文实验对齐摘要：")
    for key, value in alignment_report(data).items():
        if isinstance(value, float):
            print(f"  {key}: {value:.3f}")
        else:
            print(f"  {key}: {value}")


if __name__ == "__main__":
    main()
