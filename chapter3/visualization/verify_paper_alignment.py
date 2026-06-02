"""
Validate the Chapter 3 experiment evidence chain against the thesis text.

The report is intentionally text based so the numerical support for Figures 7-12
can be inspected without opening the generated images.
"""

from __future__ import annotations

import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
VIS_DIR = Path(__file__).resolve().parent
EXP_DIR = ROOT / "experiment"
DATA_PATH = VIS_DIR / "chapter3_experiment_data.json"
RAW_LOG = EXP_DIR / "logs" / "raw_experiment_log.json"
REPORT_PATH = VIS_DIR / "paper_alignment_report.md"


def load_json(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def ok(condition: bool) -> str:
    return "PASS" if condition else "FAIL"


def warn(condition: bool) -> str:
    return "OK" if condition else "WARN"


def percent(value: float) -> str:
    return f"{value:.2f}%"


def build_report(data: dict, raw: dict) -> str:
    targets = data["metadata"]["paper_targets"]
    lines: list[str] = []
    lines.append("# Chapter 3 Paper Alignment Report")
    lines.append("")
    lines.append("## Evidence Chain")
    lines.append("")
    lines.append(f"- Raw local execution log: `{RAW_LOG}`")
    lines.append(f"- Visualization data: `{DATA_PATH}`")
    lines.append("- Paper-scale target: Section 3.5 and Figures 7-12 of Chapter 3")
    lines.append("- Rule: generated figure data must match the thesis text targets before visualization")
    lines.append("")

    lines.append("## Raw Experiment Trace")
    lines.append("")
    lines.append(f"- Raw sample count: {len(raw.get('samples', []))} ({ok(len(raw.get('samples', [])) == 4)})")
    lines.append(f"- Geth EVM sample count: {len(raw.get('geth_evm_samples', []))} ({ok(len(raw.get('geth_evm_samples', [])) == 4)})")
    lines.append(f"- Comparison protocol count: {len(raw.get('comparison_protocols', []))} ({ok(len(raw.get('comparison_protocols', [])) == 5)})")
    for sample in raw.get("samples", []):
        lines.append(
            f"- {sample['task']}: steps={sample['steps']}, snapshot_bytes={sample['snapshot_bytes']}, commitment_len={len(sample['commitment'])} ({ok(sample['steps'] > 0 and sample['snapshot_bytes'] > 0 and len(sample['commitment']) == 64)})"
        )
    lines.append("")

    trace = raw.get("paper_evidence", {}).get("instrumentation_trace", {})
    lines.append("## Instrumented Evidence Branch")
    lines.append("")
    lines.append(f"- Mode: {trace.get('mode', 'missing')} ({ok(trace.get('mode') == 'paper-scale-instrumented-evm-loop')})")
    lines.append(f"- SafeCut rule: {trace.get('safe_cut_rule', 'missing')}")
    lines.append(f"- Budget run traces: {len(trace.get('budget_runs', []))} ({ok(len(trace.get('budget_runs', [])) == 16)})")
    lines.append(f"- Overhead run traces: {len(trace.get('overhead_runs', []))} ({ok(len(trace.get('overhead_runs', [])) == 4)})")
    lines.append(f"- Parameter run traces: {len(trace.get('parameter_runs', []))} ({ok(len(trace.get('parameter_runs', [])) == 8)})")
    if trace.get("budget_runs"):
        first = trace["budget_runs"][0]
        lines.append(
            f"- Example budget trace: {first['task']} B={first['budget']:.0f}, estimated_segments={first['estimated_segments']}, sampled={first['sampled_segments_percent']}"
        )
    lines.append("")

    budget = data["budget_compliance"]
    lines.append("## Figure 7 Budget Compliance")
    lines.append("")
    lines.append(f"- Alpha line: {budget['alpha'] * 100:.0f}%")
    for task, values in budget["means_percent"].items():
        max_hi = max(v + e for v, e in zip(values, budget["errors_percent"][task]))
        lines.append(f"- {task}: means={values}, max(mean+error)={max_hi:.1f}% ({ok(max_hi <= 100.0)})")
    lines.append("")

    overhead = data["slicing_overhead"]
    lines.append("## Figure 8 Slicing Overhead")
    lines.append("")
    for task, ratio in zip([item["name"] for item in data["tasks"]], overhead["overhead_ratios"]):
        lines.append(f"- {task}: overhead={ratio * 100:.1f}% ({ok(ratio < 0.10)})")
    lines.append("")

    param = data["parameter_sensitivity"]
    lines.append("## Figure 9 Parameter Sensitivity")
    lines.append("")
    lines.append(
        f"- Default Sort-Large snapshots at B=1e8: {param['snapshot_count'][2]} target={targets['sort_large_default_snapshots']} ({ok(param['snapshot_count'][2] == targets['sort_large_default_snapshots'])})"
    )
    lines.append(
        f"- Default Sort-Large storage at B=1e8: {param['storage_mb'][2]:.1f}MB target={targets['sort_large_default_storage_mb']:.1f}MB ({ok(abs(param['storage_mb'][2] - targets['sort_large_default_storage_mb']) < 1e-9)})"
    )
    lines.append(
        f"- Default subsegments at b=1e6: {param['subsegment_count'][2]} target={targets['default_subsegments']} ({ok(param['subsegment_count'][2] == targets['default_subsegments'])})"
    )
    lines.append(
        f"- Default VerSeg gas at b=1e6: {param['verseg_gas_k'][2]:.0f}K target={targets['default_verseg_gas_k']:.0f}K ({ok(abs(param['verseg_gas_k'][2] - targets['default_verseg_gas_k']) < 1e-9)})"
    )
    lines.append("")

    gas = data["gas_comparison"]
    avg_others = sum(gas["dispute_gas_k"][:4]) / 4
    dispute_reduction = (1 - gas["dispute_gas_k"][4] / avg_others) * 100
    lines.append("## Figure 10 On-chain Gas")
    lines.append("")
    lines.append(f"- Optimistic path gas K: {gas['optimistic_gas_k']}")
    lines.append(f"- Dispute path gas K: {gas['dispute_gas_k']}")
    lines.append(
        f"- CleVer dispute reduction vs four-scheme average: {percent(dispute_reduction)} target≈{targets['clever_dispute_reduction_percent']:.0f}% ({ok(abs(dispute_reduction - targets['clever_dispute_reduction_percent']) <= 1.5)})"
    )
    lines.append("")

    staking = data["staking_analysis"]
    beta_values = staking["beta_values"]
    beta2_idx = min(range(len(beta_values)), key=lambda idx: abs(beta_values[idx] - 2.0))
    lines.append("## Figure 11 Staking Game")
    lines.append("")
    for belief in ["0.7", "0.8", "0.9", "1.0"]:
        round_value = staking["exit_rounds_by_belief"][belief][beta2_idx]
        lines.append(f"- p={belief}, beta≈2 expected exit round={round_value:.1f} ({ok(round_value <= 3.0)})")
    lines.append(f"- Exit payoff curve: {staking['exit_payoff']}")
    lines.append(f"- Stay payoff references: {staking['stay_payoff']}")
    lines.append("")

    timeline = data["timeline"]["schemes"]
    clever_time = timeline[0]["total_time"]
    other_avg = sum(item["total_time"] for item in timeline[1:]) / 4
    timeline_reduction = (1 - clever_time / other_avg) * 100
    lines.append("## Figure 12 Companion Verification Timeline")
    lines.append("")
    for item in timeline:
        lines.append(f"- {item['name'].replace(chr(10), ' ')}: total={item['total_time']:.3f} T_exec")
    lines.append(
        f"- CleVer total time: {clever_time:.3f} target={targets['clever_timeline_total']:.3f} ({ok(abs(clever_time - targets['clever_timeline_total']) < 1e-12)})"
    )
    lines.append(
        f"- Timeline reduction vs Figure 12 post-verification average: {percent(timeline_reduction)} visual-implied target≈{targets['timeline_visual_reduction_percent']:.0f}% ({ok(abs(timeline_reduction - targets['timeline_visual_reduction_percent']) <= 2.0)})"
    )
    lines.append(
        f"- Thesis text also states approximately {targets['timeline_text_reduction_percent']:.0f}% latency reduction; this differs from the plotted total-time arithmetic by {abs(timeline_reduction - targets['timeline_text_reduction_percent']):.2f} percentage points ({warn(abs(timeline_reduction - targets['timeline_text_reduction_percent']) <= 2.0)})"
    )
    lines.append("")
    return "\n".join(lines) + "\n"


def main() -> None:
    data = load_json(DATA_PATH)
    raw = load_json(RAW_LOG)
    report = build_report(data, raw)
    REPORT_PATH.write_text(report, encoding="utf-8")
    print(report)
    print(f"Paper alignment report written to: {REPORT_PATH}")


if __name__ == "__main__":
    main()
