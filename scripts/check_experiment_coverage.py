#!/usr/bin/env python3
"""Check that thesis experiment artifacts are covered by runnable code paths."""

from __future__ import annotations

import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def load_json(path: Path) -> dict:
    if not path.exists():
        raise AssertionError(f"missing artifact: {path.relative_to(ROOT)}")
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def close(a: float, b: float, tol: float = 1e-9) -> bool:
    return abs(a - b) <= tol


def require_close(a: float, b: float, message: str, tol: float = 1e-9) -> None:
    assert close(float(a), float(b), tol), f"{message}: got {a}, want {b}"


def check_chapter3() -> list[str]:
    raw = load_json(ROOT / "chapter3/experiment/logs/raw_experiment_log.json")
    data = load_json(ROOT / "chapter3/visualization/chapter3_experiment_data.json")
    standalone = load_json(ROOT / "chapter3/experiment/logs/comparison_protocols.json")
    messages: list[str] = []

    assert len(raw.get("samples", [])) == 4, "chapter3 raw log must contain four runnable task samples"
    assert len(raw.get("geth_evm_samples", [])) == 4, "chapter3 raw log must contain four geth evm samples"
    for sample in raw["samples"]:
        assert sample["steps"] > 0, f"chapter3 {sample['task']} did not execute"
        assert sample["snapshot_bytes"] > 0, f"chapter3 {sample['task']} snapshot missing"
        assert len(sample["commitment"]) == 64, f"chapter3 {sample['task']} commitment is not sha256 hex"
    for sample in raw["geth_evm_samples"]:
        assert sample["gas_used"] > 0, f"chapter3 {sample['task']} geth evm gas missing"

    required_schemes = {
        "Arbitrum\nClassic": "arbitrumClassicPath",
        "TrueBit": "truebitPath",
        "Cartesi\nDave": "cartesiDavePath",
        "Arbitrum\nBoLD": "boldPath",
        "CleVer\n(ours)": "cleverPath",
    }
    by_scheme = {item["scheme"]: item for item in raw.get("comparison_protocols", [])}
    foundry_by_function = {item["function"]: item["gas"] for item in raw.get("foundry_gas_runs", [])}
    for scheme, function in required_schemes.items():
        item = by_scheme.get(scheme)
        assert item, f"chapter3 missing comparison protocol: {scheme}"
        benchmark = item.get("dispute_benchmark", {})
        assert benchmark.get("function") == function, f"chapter3 {scheme} dispute benchmark not wired to {function}"
        assert foundry_by_function.get(function) == benchmark.get("gas"), f"chapter3 {scheme} benchmark gas not parsed from Foundry output"
        assert close(float(benchmark.get("gas_k", 0)), float(item["dispute_gas_k"])), f"chapter3 {scheme} gas not derived from benchmark"
        assert benchmark.get("gas", 0) > 0, f"chapter3 {scheme} dispute benchmark gas missing"
        optimistic = item.get("optimistic_benchmark", {})
        assert foundry_by_function.get(optimistic.get("function")) == optimistic.get("gas"), f"chapter3 {scheme} optimistic gas not parsed from Foundry output"
        assert item.get("reproduce"), f"chapter3 {scheme} missing reproduction command"
        assert item.get("validation_claim"), f"chapter3 {scheme} missing validation claim"
        assert item.get("source_kind") in {"solidity_protocol_benchmark", "project_solidity_benchmark"}, f"chapter3 {scheme} source kind mismatch"
        assert item.get("measurement_method"), f"chapter3 {scheme} missing measurement method"
        assert "official source binding" not in item.get("notes", ""), f"chapter3 {scheme} still uses source-binding wording"
        assert "official protocol flow" not in item.get("measurement_method", ""), f"chapter3 {scheme} still describes source material as the measurement"
        assert len(item.get("dispute_flow", [])) >= 4, f"chapter3 {scheme} missing submit/challenge/localize/adjudicate flow"

    gas_comparison = data.get("gas_comparison", {})
    assert len(gas_comparison.get("schemes", [])) == 5, "chapter3 figure 10 data missing comparison schemes"
    assert len(gas_comparison.get("protocols", [])) == 5, "chapter3 structured data missing protocol provenance"
    assert len(standalone.get("protocols", [])) == 5, "chapter3 standalone comparison report missing protocols"
    standalone_by_scheme = {item["scheme"]: item for item in standalone["protocols"]}
    for scheme, raw_item in by_scheme.items():
        report_item = standalone_by_scheme[scheme]
        assert report_item["dispute_benchmark"]["gas"] == raw_item["dispute_benchmark"]["gas"], f"chapter3 {scheme} standalone report gas mismatch"
        assert report_item["optimistic_benchmark"]["gas"] == raw_item["optimistic_benchmark"]["gas"], f"chapter3 {scheme} standalone optimistic gas mismatch"
    avg_others = sum(by_scheme[name]["dispute_gas_k"] for name in list(required_schemes)[:4]) / 4
    reduction = (1 - by_scheme["CleVer\n(ours)"]["dispute_gas_k"] / avg_others) * 100
    require_close(avg_others, 4445.867, "chapter3 figure 10 comparison average")
    require_close(round(reduction, 2), 87.58, "chapter3 figure 10 reduction")

    timeline = raw["paper_evidence"]["timeline"]["schemes"]
    assert len(timeline) == 5, "chapter3 timeline missing comparison schemes"
    for item in timeline:
        assert item.get("formula"), f"chapter3 timeline formula missing for {item['name']}"
        assert float(item["total_time"]) > 0, f"chapter3 timeline not positive for {item['name']}"
    clever_total = next(item["total_time"] for item in timeline if item["name"] == "CleVer\n（并发）")
    other_totals = [item["total_time"] for item in timeline if item["name"] != "CleVer\n（并发）"]
    assert all(clever_total < value for value in other_totals), "chapter3 timeline ordering does not show CleVer lower than post-execution schemes"
    require_close(clever_total, 1.044, "chapter3 figure 12 CleVer timeline")

    evidence = raw["paper_evidence"]
    overhead = [round(value * 100, 1) for value in evidence["slicing_overhead"]["overhead_ratios"]]
    assert overhead == [2.6, 2.7, 3.5, 9.0], f"chapter3 figure 8 overhead mismatch: {overhead}"
    sensitivity = evidence["parameter_sensitivity"]
    assert sensitivity["snapshot_count"] == [130000, 13000, 1300, 130], "chapter3 figure 9 snapshot count mismatch"
    assert sensitivity["storage_mb"] == [6100, 610, 61, 6.1], "chapter3 figure 9 storage mismatch"
    assert sensitivity["subsegment_count"] == [10000, 1000, 100, 10], "chapter3 figure 9 subsegment count mismatch"
    assert [round(v) for v in sensitivity["verseg_gas_k"]] == [12, 120, 1200, 12000], "chapter3 figure 9 VerSeg gas mismatch"
    budget_ratios = [
        value
        for task in evidence["budget_compliance"]["segments_percent"].values()
        for samples in task.values()
        for value in samples
    ]
    assert 0 <= min(budget_ratios) and max(budget_ratios) <= 100, "chapter3 figure 7 budget ratios out of range"

    messages.append("chapter3: figures 7-12 covered by runnable logs, local protocol benchmarks, and benchmark-derived gas")
    return messages


def check_chapter4() -> list[str]:
    raw = load_json(ROOT / "chapter4/experiment/logs/raw_experiment_log.json")
    data = load_json(ROOT / "chapter4/visualization/chapter4_experiment_data.json")
    for key in ["protocol_trace", "zk_circuits", "foundry_gas_runs"]:
        assert key in raw, f"chapter4 raw log missing {key}"
        assert key in data, f"chapter4 structured data missing {key}"
        assert data[key] == raw[key], f"chapter4 structured {key} is not copied from raw log"
    figures = data.get("figures", {})
    for key in ["fig14", "fig15", "fig16_17", "fig18", "fig19", "fig20", "fig21", "fig22"]:
        assert key in figures, f"chapter4 structured figure data missing {key}"

    foundry = {item["function"]: item["gas"] for item in raw["foundry_gas_runs"]}
    for name in ["testAnonyReg", "testAnonyStake", "testPresentCred", "testElectCTWR", "testPublicReg", "testPublicStake", "testCandidateDeclare"]:
        assert foundry.get(name, 0) > 0, f"chapter4 missing Foundry gas for {name}"
    single = figures["fig22"]["single_call"]
    require_close(single["anon_gas_k"]["AnonyReg"], foundry["testAnonyReg"] / 1000, "chapter4 fig22 AnonyReg gas")
    require_close(single["anon_gas_k"]["AnonyStake"], foundry["testAnonyStake"] / 1000, "chapter4 fig22 AnonyStake gas")
    require_close(single["anon_gas_k"]["PresentCred"], foundry["testPresentCred"] / 1000, "chapter4 fig22 PresentCred gas")
    require_close(single["nonanon_gas_k"]["PublicReg"], foundry["testPublicReg"] / 1000, "chapter4 fig22 PublicReg gas")
    require_close(single["nonanon_gas_k"]["PublicStake"], foundry["testPublicStake"] / 1000, "chapter4 fig22 PublicStake gas")
    require_close(single["nonanon_gas_k"]["CandDecl"], foundry["testCandidateDeclare"] / 1000, "chapter4 fig22 CandidateDeclare gas")
    n_values = figures["fig22"]["n"]
    index_100 = n_values.index(100)
    require_close(figures["fig22"]["cycle_gas_k"]["total_anon"][index_100], 56190.1, "chapter4 fig22 N=100 anonymous cycle gas", tol=1e-6)
    require_close(figures["fig22"]["cycle_gas_k"]["total_nonanon"][index_100], 8225.4, "chapter4 fig22 N=100 public baseline gas", tol=1e-6)

    circuits = {item["name"]: item for item in raw["zk_circuits"]}
    expected_constraints = {"AnonyReg": 1088, "AnonyStake": 1328, "PresentCred": 5817}
    for name, expected in expected_constraints.items():
        assert circuits[name]["witness_status"] == "PASS", f"chapter4 {name} witness check failed"
        assert circuits[name]["constraint_count"] == expected, f"chapter4 {name} constraint count mismatch"
    circuit_stats = {item["name"]: item for item in figures["fig22"]["circuit_stats"]}
    for name in expected_constraints:
        assert circuit_stats[name]["constraint_count"] == circuits[name]["constraint_count"], f"chapter4 fig22 {name} circuit stats not raw-derived"

    trace = figures["fig14"]["trace_rho_035"]
    require_close(trace["rho"], 0.35, "chapter4 fig14 trace rho")
    require_close(trace["rho_beacon"], 0.4038604305864885, "chapter4 fig14 beacon share")
    require_close(trace["rho_eff"], 0.3495555154074177, "chapter4 fig14 effective CTWR share")
    peaks = {
        rho: (case["k_star"], max(case["ctwr"]))
        for rho, case in figures["fig16_17"].items()
    }
    assert peaks["0.20"][0] == 248 and peaks["0.35"][0] == 535 and peaks["0.50"][0] == 787, "chapter4 fig16 CTWR split peaks mismatch"
    require_close(peaks["0.35"][1], 4.911880862225459e-10, "chapter4 fig16 rho=0.35 CTWR peak")
    fig18 = figures["fig18"]
    require_close(fig18["k_threshold_s_adv_over_s_cap"], 8.0, "chapter4 fig18 S_A/S_cap threshold")
    assert fig18["k_star_profit"] == 8, "chapter4 fig18 profit peak mismatch"
    require_close(fig18["k_break_even"], 14.285714285714286, "chapter4 fig18 break-even point")

    out_dir = ROOT / "chapter4/visualization/正确图片输出"
    for index in range(14, 23):
        assert (out_dir / f"fig{index}.png").exists(), f"chapter4 output image missing fig{index}.png"
    return ["chapter4: figures 14-22 covered by CTWR formula traces, Circom witness checks, and Foundry gas"]


def check_chapter5() -> list[str]:
    raw = load_json(ROOT / "chapter5/experiment/logs/raw_experiment_log.json")
    data = load_json(ROOT / "chapter5/visualization/chapter5_experiment_data.json")
    messages: list[str] = []

    for key in ["ranck_traces", "senck_traces", "monte_carlo", "overhead_traces", "gas_trace"]:
        assert key in raw, f"chapter5 raw log missing {key}"
    assert raw["ranck_traces"][0]["cont_audit"]["passed"], "chapter5 honest RanCk trace failed"
    assert raw["senck_traces"][0]["sent_report"]["passed"], "chapter5 honest SenCk trace failed"
    assert any(t["cont_audit"]["detected"] for t in raw["ranck_traces"][1:]), "chapter5 RanCk deviations not detected"
    assert any(t["sent_report"]["detected"] for t in raw["senck_traces"][1:]), "chapter5 SenCk deviations not detected"

    gas = raw["gas_trace"]
    pod = gas.get("pod_baseline", {})
    foundry_tests = {item["test"]: item for item in gas.get("foundry_parsed", [])}
    for test_name in [
        "testTrackInitGas",
        "testHBRespondGas",
        "testContAuditGas",
        "testSentReportGas",
        "testDisputeGas",
        "testSentProveGas",
        "testPoDMineBountyGas",
    ]:
        assert foundry_tests.get(test_name, {}).get("gas", 0) > 0, f"chapter5 missing Foundry gas for {test_name}"
    assert pod.get("test") == "testPoDMineBountyGas", "chapter5 PoD MineBounty missing Foundry test provenance"
    assert pod.get("function") == "podMineBounty", "chapter5 PoD MineBounty missing Solidity function provenance"
    assert pod.get("gas_per_epoch") == gas["pod_gas_per_epoch"], "chapter5 PoD epoch gas mismatch"
    assert pod.get("gas_per_epoch") == foundry_tests["testPoDMineBountyGas"]["gas"], "chapter5 PoD baseline is not using Foundry podMineBounty gas"
    assert pod.get("monthly_gas") == gas["monthly_by_scheme"]["pod_theta_0.9"], "chapter5 PoD monthly gas mismatch"
    expected_pod_monthly = int(pod["epochs_per_month"] * pod["theta"] * pod["validator_count"] * pod["gas_per_epoch"])
    assert pod.get("monthly_gas") == expected_pod_monthly, "chapter5 PoD monthly gas is not formula-derived"
    assert pod.get("reference_sources"), "chapter5 PoD MineBounty missing reference sources"
    assert pod.get("validation_claim"), "chapter5 PoD MineBounty missing validation claim"
    pod_trace = gas.get("pod_experiment_trace", {})
    assert pod_trace.get("honest_all_valid"), "chapter5 PoD honest watchtower trace failed"
    assert pod_trace.get("lazy_detected") and pod_trace.get("invalid_proof_count", 0) > 0, "chapter5 PoD lazy watchtower trace not detected"
    assert len(pod_trace.get("epochs", [])) == pod_trace.get("watchtower_count", 0) * pod_trace.get("epoch_count", 0), "chapter5 PoD epoch trace incomplete"

    experiments = raw.get("comparison_experiments", [])
    rows = raw.get("comparison_table_11", [])
    assert len(experiments) == len(rows) == 4, "chapter5 table 11 comparison experiments missing"
    by_scheme = {item["scheme"]: item for item in experiments}
    expected = {
        "TrueBit": ("no", "partial", "-", "no"),
        "Arbitrum": ("no", "no", "-", "no"),
        "PoD": ("yes", "no", "limited", "yes"),
        "RanCk+SenCk": ("yes", "yes", "yes", "no"),
    }
    for row in rows:
        scheme = row["scheme"]
        exp = by_scheme[scheme]
        assert len(exp.get("scenarios", [])) == 3, f"chapter5 {scheme} missing scenario evidence"
        assert len(exp.get("protocol_flow", [])) >= 4, f"chapter5 {scheme} missing protocol flow evidence"
        assert exp.get("benchmark_mode"), f"chapter5 {scheme} missing benchmark mode"
        assert exp.get("reproduce") and exp.get("validation_claim"), f"chapter5 {scheme} missing comparison provenance"
        observed = (row["online_audit"], row["semantic_diligence"], row["unpredictable_audit"], row["external_network"])
        assert observed == expected[scheme], f"chapter5 {scheme} table 11 mismatch"
        assert exp.get("source_kind"), f"chapter5 {scheme} source kind missing"

    assert "comparison_experiments" in data, "chapter5 structured data missing comparison experiments"
    assert data["gas"]["pod_baseline"]["monthly_gas"] == data["gas"]["monthly_by_scheme"]["pod_theta_0.9"], "chapter5 structured PoD monthly mismatch"
    assert data["gas"]["pod_baseline"].get("source_kind") == "solidity_pod_mine_bounty_benchmark_plus_paper_protocol_trace", "chapter5 PoD MineBounty source kind mismatch"
    assert data["gas"]["pod_baseline"].get("reference_sources"), "chapter5 structured PoD reference sources missing"

    pi_sweep = data["gas"]["pi_h_sweep"]
    assert all(point["monthly_gas"] <= data["gas"]["monthly_by_scheme"]["pod_theta_0.9"] for point in pi_sweep), "chapter5 figure 31 sweep exceeds PoD MineBounty unexpectedly"
    provenance = data["gas"]["measurement_provenance"]
    assert provenance.get("measurement_status") == "checked", "chapter5 gas measurement status missing"
    assert provenance.get("foundry_test_level_gas"), "chapter5 Foundry gas provenance missing"
    assert data["gas"]["default_per_validator_gas"] > 0, "chapter5 default per-validator gas missing"
    assert data["gas"]["monthly_by_scheme"]["ours_pi_h_0.01"] > 0, "chapter5 monthly gas missing"
    assert data["gas"]["pod_gas_per_epoch"] == pod["gas_per_epoch"], "chapter5 PoD gas not derived from baseline"
    operations = {item["name"]: item for item in data["gas"]["operations"]}
    expected_default = (
        operations["RanCk TrackInit"]["gas"] * operations["RanCk TrackInit"]["default_uses"]
        + operations["RanCk HBRespond"]["gas"] * operations["RanCk HBRespond"]["default_uses"]
        + operations["RanCk ContAudit"]["gas"] * operations["RanCk ContAudit"]["default_uses"]
        + operations["SenCk SentReport"]["gas"] * operations["SenCk SentReport"]["default_uses"]
    )
    assert data["gas"]["default_per_validator_gas"] == expected_default, "chapter5 default per-validator gas is not operation-derived"
    sent_prove = foundry_tests["testSentProveGas"]["gas"]
    verseg_replay = int(round(1.2 * raw["params"]["b"]))
    assert operations["SentProve/VerSeg"]["gas"] == sent_prove + verseg_replay, "chapter5 SentProve/VerSeg gas is not formula-derived"

    messages.append("chapter5: figures 24-31 and tables 9-11 covered by protocol traces, scenario experiments, and Foundry gas")
    return messages


def main() -> None:
    checks = check_chapter3() + check_chapter4() + check_chapter5()
    for message in checks:
        print(f"PASS {message}")


if __name__ == "__main__":
    main()
