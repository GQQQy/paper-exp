#!/usr/bin/env python3
"""Generate Chapter 5 experiment data and figures.

The script consumes chapter5/experiment/logs/raw_experiment_log.json. If the
raw log is missing, it runs the Go experiment entrypoint first. All plot arrays
are derived from the raw protocol traces, formulas recorded in the raw log,
Monte Carlo outputs, and Foundry gas measurements. Thesis table values are not
used as fallback plot data.
"""

from __future__ import annotations

import json
import math
import os
import subprocess
from pathlib import Path

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt
import numpy as np
from matplotlib import font_manager
from matplotlib.colors import ListedColormap
from matplotlib.patches import Patch
from matplotlib.ticker import FuncFormatter, LogLocator, NullFormatter


ROOT = Path(__file__).resolve().parents[1]
EXP_DIR = ROOT / "experiment"
VIS_DIR = ROOT / "visualization"
RAW_LOG = EXP_DIR / "logs" / "raw_experiment_log.json"
OUT_JSON = VIS_DIR / "chapter5_experiment_data.json"
OUT_DIR = VIS_DIR / "正确图片输出"

GO = "/usr/local/go/bin/go"

COLORS = {
    "light_blue": "#90C9E7",
    "cyan": "#219EBC",
    "dark_blue": "#136783",
    "navy": "#02304A",
    "gold": "#FEB705",
    "orange": "#FF9E02",
    "dark_orange": "#FA8600",
    "green": "#78C6A3",
    "gray": "#777777",
    "pale": "#f2f2f2",
}


def configure_matplotlib() -> None:
    available = [f.name for f in font_manager.fontManager.ttflist]
    preferred = [
        "Microsoft YaHei",
        "SimHei",
        "SimSun",
        "WenQuanYi Micro Hei",
        "Noto Sans CJK SC",
        "STHeiti",
        "PingFang SC",
        "Songti SC",
        "Heiti TC",
        "AR PL UMing CN",
        "DejaVu Sans",
    ]
    usable = [f for f in preferred if f in available] or ["DejaVu Sans"]
    plt.rcParams.update(
        {
            "font.size": 10.5,
            "axes.labelsize": 10.5,
            "axes.titlesize": 11.0,
            "legend.fontsize": 8.5,
            "figure.dpi": 600,
            "savefig.dpi": 600,
            "font.family": "sans-serif",
            "font.sans-serif": usable,
            "mathtext.fontset": "dejavusans",
            "axes.unicode_minus": False,
            "axes.grid": True,
            "grid.alpha": 0.25,
            "axes.edgecolor": "#333333",
            "axes.linewidth": 0.8,
        }
    )


def regenerate_raw_log() -> None:
    env = os.environ.copy()
    env["GOCACHE"] = "/private/tmp/chapter5-gocache"
    Path(env["GOCACHE"]).mkdir(parents=True, exist_ok=True)
    subprocess.run([GO, "run", "./cmd/audit-exp", "--out", str(RAW_LOG)], cwd=EXP_DIR, env=env, check=True)


def ensure_raw_log() -> dict:
    if not RAW_LOG.exists():
        regenerate_raw_log()
    raw = json.loads(RAW_LOG.read_text(encoding="utf-8"))
    gas = raw.get("gas_trace", {})
    provenance = gas.get("measurement_provenance", {})
    if (
        "detection_parameters" not in raw
        or "gamma_hit_sweep" not in raw.get("monte_carlo", {})
        or not provenance.get("no_paper_table_fallback")
    ):
        regenerate_raw_log()
        raw = json.loads(RAW_LOG.read_text(encoding="utf-8"))
    validate_raw(raw)
    return raw


def validate_raw(raw: dict) -> None:
    assert raw["params"]["block_time_seconds"] == 12
    assert raw["params"]["B"] == 1e8
    assert raw["params"]["b"] == 1e6
    assert raw["ranck_traces"], "missing RanCk traces"
    assert raw["senck_traces"], "missing SenCk traces"
    assert raw["detection_parameters"], "missing detection parameter provenance"
    assert raw["monte_carlo"]["gamma_hit_sweep"], "missing Gamma hit sweep"
    assert raw["ranck_traces"][0]["cont_audit"]["passed"], "honest RanCk failed"
    assert raw["senck_traces"][0]["sent_report"]["passed"], "honest SenCk failed"
    assert any(t["cont_audit"]["detected"] for t in raw["ranck_traces"][1:]), "RanCk deviations not detected"
    assert any(t["sent_report"]["detected"] for t in raw["senck_traces"][1:]), "SenCk deviations not detected"
    assert raw["gas_trace"]["default_per_validator_gas"] > 0
    assert raw["gas_trace"]["measurement_provenance"]["no_paper_table_fallback"], "gas trace uses paper table fallback"


def ranck_detect(pi_h: float, ell: np.ndarray | float) -> np.ndarray | float:
    return 1.0 - (1.0 - pi_h) ** ell


def ranck_combined_pass(pi_h: float, ell: np.ndarray, n: int, s: int) -> np.ndarray:
    p_hb = (1.0 - pi_h) ** ell
    p_cont = np.maximum(0.0, 1.0 - ell / n) ** s
    return np.maximum(1e-14, p_hb * p_cont)


def senck_pass(r: float, ms: np.ndarray | int) -> np.ndarray | float:
    return (1.0 - r) ** ms


def build_data(raw: dict) -> dict:
    params = raw["params"]
    det_params = raw["detection_parameters"]
    ell = np.arange(1, det_params["offline_length_max_blocks"] + 1)
    ms_values = list(range(det_params["senck_m_s_min"], det_params["senck_m_s_max"] + 1))
    detection = {
        "ranck_heartbeat": {
            "ell": ell.tolist(),
            "series": {f"M={m}": ranck_detect(1 / m, ell).tolist() for m in det_params["heartbeat_M_values"]},
        },
        "ranck_combined_pass": {
            "ell": ell.tolist(),
            "pi_h": det_params["combined_pi_h"],
            "n": det_params["combined_audit_window_blocks"],
            "series": {
                f"s={s}": ranck_combined_pass(
                    det_params["combined_pi_h"],
                    ell,
                    det_params["combined_audit_window_blocks"],
                    s,
                ).tolist()
                for s in det_params["continuity_sample_sizes"]
            },
        },
        "senck_lazy_pass": {
            "m_s": ms_values,
            "series": {f"rho={r:g}": [senck_pass(r, m) for m in ms_values] for r in det_params["senck_rho_values"]},
        },
    }
    data = {
        "metadata": {
            "chapter": raw["metadata"]["chapter"],
            "raw_log": str(RAW_LOG),
            "params": params,
            "source": raw["metadata"]["source"],
        },
        "table9_parameters": raw["table9_parameters"],
        "protocol_coverage": raw["protocol_coverage"],
        "ranck_traces": raw["ranck_traces"],
        "senck_traces": raw["senck_traces"],
        "detection_parameters": det_params,
        "feasibility": raw["feasibility"],
        "detection": detection,
        "monte_carlo": raw["monte_carlo"],
        "overhead": raw["overhead_traces"],
        "gas": raw["gas_trace"],
        "comparison_table_11": raw["comparison_table_11"],
    }
    OUT_JSON.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8")
    return data


def bottom_title(ax, title: str) -> None:
    ax.text(0.5, -0.24, title, transform=ax.transAxes, ha="center", va="top", fontsize=12, fontweight="bold")


def log_axis(ax) -> None:
    def fmt(value, _pos):
        if value <= 0:
            return ""
        if abs(value - 1) < 1e-12:
            return "1"
        exp = int(round(math.log10(value)))
        return f"$10^{{{exp}}}$"

    ax.yaxis.set_major_locator(LogLocator(base=10))
    ax.yaxis.set_major_formatter(FuncFormatter(fmt))
    ax.yaxis.set_minor_formatter(NullFormatter())


def save(fig, name: str) -> None:
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    fig.savefig(OUT_DIR / f"{name}.png", bbox_inches="tight")
    plt.close(fig)


def fig24(data: dict) -> None:
    c1 = data["feasibility"]["C1_online_deviation"]
    c2 = data["feasibility"]["C2_diligence_deviation"]
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(12, 4.8))
    for t_win, color in [(3600, COLORS["light_blue"]), (7200, COLORS["cyan"]), (14400, COLORS["dark_orange"])]:
        pts = [p for p in c1 if p["T_win"] == t_win]
        ax1.plot([p["pi_h"] for p in pts], [p["min_c_hb_over_c_track_plus_c_m"] for p in pts], lw=2, color=color, label=fr"$T_{{win}}={t_win}$")
    ax1.fill_between(
        [p["pi_h"] for p in c1 if p["T_win"] == 7200],
        [p["min_c_hb_over_c_track_plus_c_m"] for p in c1 if p["T_win"] == 7200],
        0.12,
        color=COLORS["cyan"],
        alpha=0.10,
    )
    ax1.set_xlabel(r"心跳触发概率 $\pi_h$")
    ax1.set_ylabel(r"$c_{hb}/(c_{track}+c_m)$ 最小值")
    ax1.set_xlim(0.002, 0.025)
    ax1.set_ylim(0, 0.12)
    ax1.legend(loc="upper right")
    ax1.annotate("可行区域", xy=(0.018, 0.02), color=COLORS["dark_blue"], ha="center")
    ax1.annotate("不可行区域", xy=(0.005, 0.09), color=COLORS["gray"], ha="center")
    bottom_title(ax1, "（a）条件 C1 的可行性边界")

    for r, color in [(0.2, COLORS["light_blue"]), (0.4, COLORS["cyan"]), (0.6, COLORS["gold"]), (0.8, COLORS["dark_orange"])]:
        pts = [p for p in c2 if abs(p["rho"] - r) < 1e-12]
        ax2.plot([p["m_s"] for p in pts], [p["min_c_sent_over_c_m"] for p in pts], "o-", lw=1.8, ms=4, color=color, label=fr"$\rho={r}$")
    ax2.axhline(1.0, color=COLORS["navy"], ls="--", lw=1, alpha=0.6)
    ax2.set_xlabel(r"抽样段数量 $m_s$")
    ax2.set_ylabel(r"$c_{sent}/c_m$ 最小值")
    ax2.set_xlim(0.5, 20.5)
    ax2.set_ylim(0.9, 8)
    ax2.set_yscale("log")
    ax2.legend(loc="upper right")
    log_axis(ax2)
    bottom_title(ax2, "（b）条件 C2 的可行性边界")
    fig.tight_layout(rect=[0, 0.08, 1, 1])
    save(fig, "fig24_feasibility_ab")


def fig25(data: dict) -> None:
    joint = data["feasibility"]["joint_feasible_region"]
    pi_values = sorted({round(p["pi_h"], 12) for p in joint})
    ms_values = sorted({p["m_s"] for p in joint})
    idx_pi = {v: i for i, v in enumerate(pi_values)}
    idx_ms = {v: i for i, v in enumerate(ms_values)}
    grid = np.zeros((len(ms_values), len(pi_values)), dtype=int)
    for p in joint:
        value = 0
        if p["C1_satisfied"] and p["C2_satisfied"]:
            value = 3
        elif p["C1_satisfied"]:
            value = 1
        elif p["C2_satisfied"]:
            value = 2
        grid[idx_ms[p["m_s"]], idx_pi[round(p["pi_h"], 12)]] = value
    fig, ax = plt.subplots(figsize=(7.2, 5.4))
    cmap = ListedColormap([COLORS["pale"], COLORS["light_blue"], "#FCE8B2", "#BFE7D3"])
    ax.grid(False)
    ax.pcolormesh(pi_values, ms_values, grid, cmap=cmap, shading="nearest")
    ax.set_xlabel(r"心跳触发概率 $\pi_h$")
    ax.set_ylabel(r"抽样段数量 $m_s$")
    ax.set_xlim(0.001, 0.025)
    ax.set_ylim(0.5, 20.5)
    ax.legend(
        handles=[
            Patch(facecolor="#BFE7D3", edgecolor="gray", label="同时满足 C1 与 C2"),
            Patch(facecolor=COLORS["light_blue"], edgecolor="gray", label="仅满足 C1"),
            Patch(facecolor="#FCE8B2", edgecolor="gray", label="仅满足 C2"),
            Patch(facecolor=COLORS["pale"], edgecolor="gray", label="均不满足"),
        ],
        loc="lower right",
    )
    ax.set_title("约束条件联合可行域")
    fig.tight_layout()
    save(fig, "fig25_joint_feasibility")


def fig26(data: dict) -> None:
    d = data["detection"]
    ell = np.array(d["ranck_heartbeat"]["ell"])
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(12, 4.6))
    colors = [COLORS["light_blue"], COLORS["cyan"], COLORS["gold"], COLORS["dark_orange"]]
    for (label, y), color in zip(d["ranck_heartbeat"]["series"].items(), colors):
        ax1.plot(ell, y, lw=1.8, color=color, label=label)
    ax1.axhline(0.95, color=COLORS["navy"], ls=":", lw=1)
    ax1.axhline(0.99, color=COLORS["navy"], ls=":", lw=1)
    ax1.set_xlabel(r"离线持续时长 $\ell$（区块）")
    ax1.set_ylabel("检测概率")
    ax1.set_xlim(0, 500)
    ax1.set_ylim(0, 1.05)
    ax1.legend(loc="lower right")
    bottom_title(ax1, "（a）心跳缺失检出概率")

    for (label, y), color in zip(d["ranck_combined_pass"]["series"].items(), [COLORS["cyan"], COLORS["gold"], COLORS["dark_orange"]]):
        ax2.semilogy(ell, y, lw=1.8, color=color, label=label)
    ax2.axhline(0.01, color=COLORS["navy"], ls=":", lw=1)
    ax2.axhline(0.001, color=COLORS["navy"], ls=":", lw=1)
    ax2.set_xlabel(r"离线持续时长 $\ell$（区块）")
    ax2.set_ylabel("侥幸通过概率")
    ax2.set_xlim(0, 500)
    ax2.set_ylim(1e-12, 1)
    ax2.legend(loc="upper right")
    log_axis(ax2)
    bottom_title(ax2, r"（b）心跳 + 连续性抽样联合通过概率")
    fig.tight_layout(rect=[0, 0.08, 1, 1])
    save(fig, "fig26_ranck_detection")


def fig27(data: dict) -> None:
    d = data["detection"]["senck_lazy_pass"]
    ms = np.array(d["m_s"])
    fig, ax = plt.subplots(figsize=(7.2, 4.8))
    for (label, y), color in zip(d["series"].items(), [COLORS["light_blue"], COLORS["cyan"], COLORS["gold"], COLORS["orange"], COLORS["dark_orange"]]):
        ax.semilogy(ms, y, "o-", lw=1.6, ms=4, color=color, label=label)
    ax.axhline(0.01, color=COLORS["navy"], ls=":", lw=1)
    ax.set_xlabel(r"抽样段数量 $m_s$")
    ax.set_ylabel("惰性策略侥幸通过概率")
    ax.set_xlim(0.5, 20.5)
    ax.set_ylim(1e-10, 1)
    ax.legend(loc="upper right")
    log_axis(ax)
    ax.set_title("SenCk 检测概率分析")
    fig.tight_layout()
    save(fig, "fig27_senck_passthrough")


def fig28(data: dict) -> None:
    fig, axes = plt.subplots(2, 2, figsize=(12, 8.4))
    ax1, ax2, ax3, ax4 = axes.flatten()
    for key, pts in data["monte_carlo"]["ranck"].items():
        x = np.array([p["parameter"] for p in pts])
        theory = np.array([p["theory"] for p in pts])
        sim = np.array([p["simulated"] for p in pts])
        ci = np.array([p["ci95"] for p in pts])
        color = {"pi_h_0.005": COLORS["light_blue"], "pi_h_0.010": COLORS["cyan"], "pi_h_0.020": COLORS["dark_orange"]}[key]
        ax1.plot(x, theory, color=color, lw=1.2, alpha=0.45)
        ax1.errorbar(x, sim, yerr=ci, fmt="o", ms=4, capsize=2, color=color, markerfacecolor="white", label=key.replace("_", "="))
    ax1.set_xlabel(r"离线持续时长 $\ell$（区块）")
    ax1.set_ylabel("检测概率")
    ax1.set_ylim(0, 1.05)
    ax1.legend(fontsize=7.8)
    bottom_title(ax1, "（a）RanCk 理论与 Monte Carlo")

    for key, pts in data["monte_carlo"]["senck"].items():
        x = np.array([p["parameter"] for p in pts])
        theory = np.array([p["theory"] for p in pts])
        sim = np.maximum(np.array([p["simulated"] for p in pts]), 1e-7)
        color = {"rho_0.2": COLORS["light_blue"], "rho_0.4": COLORS["cyan"], "rho_0.6": COLORS["gold"], "rho_0.8": COLORS["dark_orange"]}[key]
        ax2.semilogy(x, theory, color=color, lw=1.2, alpha=0.45)
        ax2.semilogy(x, sim, "o", ms=4, color=color, markerfacecolor="white", label=key.replace("_", "="))
    ax2.set_xlabel(r"抽样段数量 $m_s$")
    ax2.set_ylabel("侥幸通过概率")
    ax2.set_ylim(1e-7, 1)
    ax2.legend(fontsize=7.8, ncol=2)
    log_axis(ax2)
    bottom_title(ax2, "（b）SenCk 理论与 Monte Carlo")

    stats = data["senck_traces"][0]["trigger_stats"]
    ax3.bar([0, 1], [stats["theory_rate"], stats["observed_rate"]], width=0.55, color=[COLORS["cyan"], COLORS["orange"]], edgecolor="white")
    ax3.set_xticks([0, 1])
    ax3.set_xticklabels(["理论 $1/M_c$", "实现 trace"])
    ax3.set_ylabel("单步触发率")
    ax3.set_title(f"steps={stats['step_count']:,}, triggers={stats['trigger_count']:,}")
    bottom_title(ax3, "（c）Gamma 门控触发率")

    gamma = data["monte_carlo"]["gamma_hit_sweep"]
    labels = [f"({p['M_c']},{p['L']})" for p in gamma]
    theory = [p["theory_rho"] for p in gamma]
    observed = [p["observed_rho"] for p in gamma]
    x = np.arange(len(gamma))
    ax4.plot(x, theory, "o-", color=COLORS["dark_blue"], lw=1.8, label="理论")
    ax4.plot(x, observed, "s", color=COLORS["dark_orange"], ms=5, markerfacecolor="white", label="仿真")
    ax4.set_xticks(x)
    ax4.set_xticklabels(labels)
    ax4.set_ylabel(r"单段命中率 $\rho$")
    ax4.set_ylim(0, 1.05)
    ax4.legend()
    bottom_title(ax4, "（d）段命中率验证")
    fig.tight_layout(rect=[0, 0.06, 1, 1])
    save(fig, "fig28_monte_carlo_and_gate")


def fig29(data: dict) -> None:
    overhead = data["overhead"]
    tasks = [item["workload"]["name"] for item in overhead]
    x = np.arange(len(tasks))
    slicing = np.array([item["workload"]["chapter3_slicing_overhead_percent"] for item in overhead])
    audit_pct = np.array([item["default_report"]["audit_overhead_percent"] for item in overhead])
    enc = np.array([item["default_report"]["encoding_us"] for item in overhead])
    hsh = np.array([item["default_report"]["hash_us"] for item in overhead])
    gamma = np.array([item["default_report"]["gamma_us"] for item in overhead])
    exec_step = np.array([item["workload"]["exec_step_us"] for item in overhead])
    fig, axes = plt.subplots(2, 2, figsize=(12, 8.2))
    ax1, ax2, ax3, ax4 = axes.flatten()
    ax1.bar(x, slicing, 0.55, color=COLORS["cyan"], edgecolor="white", label="第3章切片开销")
    ax1.bar(x, audit_pct, 0.55, bottom=slicing, color=COLORS["orange"], edgecolor="white", label="第5章审计开销")
    ax1.axhline(10, color=COLORS["navy"], ls="--", lw=1)
    ax1.set_xticks(x)
    ax1.set_xticklabels(tasks, fontsize=8.5)
    ax1.set_ylabel("开销比例（%）")
    ax1.legend()
    bottom_title(ax1, "（a）切片与审计叠加开销")

    ax2.bar(x, enc, 0.55, color=COLORS["light_blue"], edgecolor="white", label="rw_t 编码")
    ax2.bar(x, hsh, 0.55, bottom=enc, color=COLORS["gold"], edgecolor="white", label="哈希")
    ax2.bar(x, gamma, 0.55, bottom=enc + hsh, color=COLORS["dark_orange"], edgecolor="white", label="Gamma")
    ax2r = ax2.twinx()
    ax2r.plot(x, exec_step, "D--", color=COLORS["dark_blue"], lw=1.5, ms=5)
    ax2r.set_ylabel("单步执行时间（us）", color=COLORS["dark_blue"])
    ax2r.tick_params(axis="y", labelcolor=COLORS["dark_blue"])
    ax2.set_xticks(x)
    ax2.set_xticklabels(tasks, fontsize=8.5)
    ax2.set_ylabel("单步审计开销（us）")
    ax2.legend(loc="upper left", fontsize=8)
    bottom_title(ax2, "（b）单步审计开销分解")

    colors = [COLORS["light_blue"], COLORS["cyan"], COLORS["gold"], COLORS["dark_orange"]]
    for item, color in zip(overhead, colors):
        ax3.plot([p["L"] for p in item["replay_by_length"]], [p["replay_ms"] for p in item["replay_by_length"]], "o-", lw=1.8, ms=4, color=color, label=item["workload"]["name"])
    ax3.axhline(1000, color=COLORS["navy"], ls="--", lw=1)
    ax3.set_xlabel(r"重放长度 $L$")
    ax3.set_ylabel("重放时间（ms）")
    ax3.set_xlim(0, 10500)
    ax3.legend(fontsize=8)
    bottom_title(ax3, "（c）快照加载与局部重放")

    snap = np.array([item["default_report"]["snapshot_ms"] for item in overhead])
    total = np.array([next(p["replay_ms"] for p in item["replay_by_length"] if p["L"] == 5000) for item in overhead])
    replay = total - snap
    ax4.barh(x, snap, 0.55, color=COLORS["cyan"], edgecolor="white", label="快照加载")
    ax4.barh(x, replay, 0.55, left=snap, color=COLORS["orange"], edgecolor="white", label="执行+审计")
    ax4.set_yticks(x)
    ax4.set_yticklabels(tasks, fontsize=8.5)
    ax4.invert_yaxis()
    ax4.set_xlabel("时间（ms）")
    ax4.legend()
    bottom_title(ax4, "（d）L=5000 重放耗时分解")
    fig.tight_layout(rect=[0, 0.06, 1, 1])
    save(fig, "fig29_overhead")


def fig30(data: dict) -> None:
    gas = data["gas"]
    ops = gas["operations"][:4]
    labels = [op["name"].replace(" ", "\n") + f"\n({op['default_uses']}次)" for op in ops]
    vals = np.array([op["gas"] * op["default_uses"] for op in ops])
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(12, 4.8))
    bars = ax1.bar(range(len(vals)), vals / 1000, color=[COLORS["light_blue"], COLORS["cyan"], COLORS["gold"], COLORS["dark_orange"]], edgecolor="white")
    for i, bar in enumerate(bars):
        pct = vals[i] / vals.sum() * 100
        ax1.text(bar.get_x() + bar.get_width() / 2, bar.get_height() + 30, f"{vals[i]/1000:.0f}K\n({pct:.0f}%)", ha="center", fontsize=8.5, color=COLORS["navy"], fontweight="bold")
    ax1.set_xticks(range(len(vals)))
    ax1.set_xticklabels(labels, fontsize=8)
    ax1.set_ylabel("Gas（K）")
    ax1.set_ylim(0, max(vals / 1000) * 1.20)
    ax1.text(0.98, 0.95, f"总计：{gas['default_per_validator_gas']/1e6:.2f}M gas/验证者/窗口", transform=ax1.transAxes, ha="right", va="top", fontsize=9)
    bottom_title(ax1, "（a）单验证者正常路径 Gas 构成")

    monthly = gas["monthly_by_scheme"]
    schemes = ["本方案\n$\\pi_h=0.01$", "本方案\n$\\pi_h=0.005$", "PoD\n$\\theta=0.9$"]
    values = [monthly["ours_pi_h_0.01"], monthly["ours_pi_h_0.005"], monthly["pod_theta_0.9"]]
    ax2.bar(range(3), np.array(values) / 1e9, color=[COLORS["cyan"], COLORS["gold"], COLORS["dark_orange"]], edgecolor="white", width=0.55)
    for i, value in enumerate(values):
        ax2.text(i, value / 1e9 + 0.03, f"{value/1e9:.2f}B", ha="center", color=COLORS["navy"], fontweight="bold")
    ax2.axhline(values[2] / 1e9, color=COLORS["dark_orange"], ls="--", lw=1.2)
    ax2.set_xticks(range(3))
    ax2.set_xticklabels(schemes)
    ax2.set_ylabel("系统月度 Gas（十亿）")
    ax2.set_ylim(0, max(np.array(values) / 1e9) * 1.20)
    bottom_title(ax2, "（b）系统月度 Gas 对比（N=20）")
    fig.tight_layout(rect=[0, 0.08, 1, 1])
    save(fig, "fig30_gas_comparison")


def fig31(data: dict) -> None:
    gas = data["gas"]
    sweep = gas["pi_h_sweep"]
    pi = np.array([p["pi_h"] for p in sweep])
    monthly = np.array([p["monthly_gas"] for p in sweep]) / 1e9
    detect = np.array([p["detection_at_300_blocks"] for p in sweep])
    pod = gas["monthly_by_scheme"]["pod_theta_0.9"] / 1e9
    fig, ax = plt.subplots(figsize=(9, 5))
    ax.plot(pi, monthly, color=COLORS["cyan"], lw=2.4, label="本方案系统月度 Gas")
    ax.axhline(pod, color=COLORS["dark_orange"], ls="--", lw=2, label=f"PoD 基线（{pod:.2f}B gas）")
    mask = monthly <= pod
    ax.fill_between(pi, monthly, pod, where=mask, color=COLORS["gold"], alpha=0.18)
    ax.set_xlabel(r"心跳触发概率 $\pi_h$")
    ax.set_ylabel("系统月度 Gas（十亿）", color=COLORS["cyan"])
    ax.tick_params(axis="y", labelcolor=COLORS["cyan"])
    ax.set_ylim(0, max(monthly.max(), pod) * 1.12)
    for ph, label in [(0.01, "默认 M=100"), (0.005, "降频 M=200")]:
        idx = int(np.argmin(np.abs(pi - ph)))
        ax.plot(pi[idx], monthly[idx], "o", color=COLORS["cyan"], ms=6)
        ax.annotate(label, xy=(pi[idx], monthly[idx]), xytext=(pi[idx] + 0.0015, monthly[idx] + 0.12), arrowprops=dict(arrowstyle="->", color=COLORS["cyan"]), fontsize=8.5)
    ax2 = ax.twinx()
    ax2.plot(pi, detect, color=COLORS["dark_orange"], lw=2, label=r"$\ell=300$ 检测概率")
    ax2.set_ylabel(r"检测概率（$\ell=300$）", color=COLORS["dark_orange"])
    ax2.tick_params(axis="y", labelcolor=COLORS["dark_orange"])
    ax2.set_ylim(0, 1.05)
    lines = [ax.lines[0], ax.lines[1], ax2.lines[0]]
    ax.legend(lines, [line.get_label() for line in lines], loc="center right")
    ax.set_title("心跳触发概率权衡分析")
    fig.tight_layout()
    save(fig, "fig31_tradeoff")


def main() -> None:
    configure_matplotlib()
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    raw = ensure_raw_log()
    data = build_data(raw)
    fig24(data)
    fig25(data)
    fig26(data)
    fig27(data)
    fig28(data)
    fig29(data)
    fig30(data)
    fig31(data)
    print(f"structured data: {OUT_JSON}")
    print(f"figures: {OUT_DIR}")


if __name__ == "__main__":
    main()
