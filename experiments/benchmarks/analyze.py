#!/usr/bin/env python3
"""
analyze.py — Parse Go benchmark output and run statistical comparisons.

Usage:
  python analyze.py <baseline.txt> <experiment.txt>     # two-run comparison
  python analyze.py <single.txt>                         # single-run summary

Produces a markdown report with:
  - Percentile distributions (p50, p95, p99)
  - Two-sample KS test + Cohen's d (two-run mode)
  - Summary table
"""

import sys
import re
import math
from collections import defaultdict
from pathlib import Path


def parse_benchmark_output(text: str) -> dict[str, list[float]]:
    """
    Parse `go test -bench` output.
    Returns { "BenchmarkName/variant": [ns_per_op, ...] }.
    """
    # Pattern: BenchmarkFoo/bar/scale-N-cpu   iterations   ns/op
    # e.g.: BenchmarkGraphSnapshot/memory/scale-100-16         31527     38830 ns/op
    pattern = re.compile(
        r'^(Benchmark\S+?/\S+?)(?:-\d+)?\s+\d+\s+([\d.]+)\s+ns/op'
    )
    results = defaultdict(list)
    for line in text.strip().splitlines():
        m = pattern.match(line.strip())
        if m:
            name = m.group(1)
            ns = float(m.group(2))
            results[name].append(ns)
    return dict(results)


def percentiles(data: list[float]) -> dict[str, float]:
    """Compute p50, p95, p99."""
    s = sorted(data)
    n = len(s)
    if n == 0:
        return {"p50": 0, "p95": 0, "p99": 0, "n": 0}

    def _p(pct: float) -> float:
        k = (pct / 100) * (n - 1)
        f = math.floor(k)
        c = math.ceil(k)
        if f == c:
            return s[int(k)]
        d0 = s[int(f)] * (c - k)
        d1 = s[int(c)] * (k - f)
        return d0 + d1

    return {
        "n": n,
        "min": s[0],
        "max": s[-1],
        "mean": sum(s) / n,
        "std": math.sqrt(sum((x - sum(s) / n) ** 2 for x in s) / (n - 1)) if n > 1 else 0,
        "p50": _p(50),
        "p95": _p(95),
        "p99": _p(99),
    }


def cohens_d(a: list[float], b: list[float]) -> float:
    """Cohen's d effect size (pooled standard deviation)."""
    n1, n2 = len(a), len(b)
    if n1 < 2 or n2 < 2:
        return 0.0

    m1 = sum(a) / n1
    m2 = sum(b) / n2

    v1 = sum((x - m1) ** 2 for x in a) / (n1 - 1)
    v2 = sum((x - m2) ** 2 for x in b) / (n2 - 1)

    pooled = math.sqrt((v1 + v2) / 2)
    if pooled == 0:
        return 0.0
    return (m1 - m2) / pooled


def ks_test(a: list[float], b: list[float]) -> tuple[float, float]:
    """
    Two-sample Kolmogorov-Smirnov test.
    Returns (D-statistic, approximate p-value).
    """
    a_sorted = sorted(a)
    b_sorted = sorted(b)
    n1, n2 = len(a_sorted), len(b_sorted)

    i, j = 0, 0
    D = 0.0
    while i < n1 and j < n2:
        if a_sorted[i] < b_sorted[j]:
            d = abs((i + 1) / n1 - j / n2)
        else:
            d = abs(j / n2 - i / n1)

        if d > D:
            D = d

        if a_sorted[i] < b_sorted[j]:
            i += 1
        else:
            j += 1

    # Approximate p-value using the asymptotic formula
    n = (n1 * n2) / (n1 + n2)
    lam = (math.sqrt(n) + 0.12 + 0.11 / math.sqrt(n)) * D

    # Kolmogorov distribution approximation
    p = 2 * sum((-1) ** (k - 1) * math.exp(-2 * k ** 2 * lam ** 2) for k in range(1, 100))
    p = max(0.0, min(1.0, p))

    return D, p


def confidence_level(p: float) -> str:
    """Map p-value to confidence level descriptor."""
    if p < 0.001:
        return "*** (p < 0.001)"
    elif p < 0.01:
        return "** (p < 0.01)"
    elif p < 0.05:
        return "* (p < 0.05)"
    else:
        return "not significant"


def effect_magnitude(d: float) -> str:
    """Describe Cohen's d magnitude."""
    d_abs = abs(d)
    if d_abs < 0.2:
        return "negligible"
    elif d_abs < 0.5:
        return "small"
    elif d_abs < 0.8:
        return "medium"
    else:
        return "large"


def format_ns(ns: float) -> str:
    """Format nanoseconds in a human-readable way."""
    if ns >= 1_000_000:
        return f"{ns / 1_000_000:.2f} ms/op"
    elif ns >= 1_000:
        return f"{ns / 1_000:.2f} µs/op"
    else:
        return f"{ns:.0f} ns/op"


def single_run_report(results: dict[str, list[float]]) -> str:
    """Generate markdown report for a single benchmark run."""
    lines = ["# Benchmark Results — Single Run\n"]
    lines.append("| Variant | Count | Mean | p50 | p95 | p99 | StdDev |")
    lines.append("|---------|-------|------|-----|-----|-----|--------|")

    for name in sorted(results.keys()):
        data = results[name]
        stats = percentiles(data)
        lines.append(
            f"| `{name}` | {stats['n']} | {format_ns(stats['mean'])} "
            f"| {format_ns(stats['p50'])} | {format_ns(stats['p95'])} "
            f"| {format_ns(stats['p99'])} | {format_ns(stats['std'])} |"
        )

    return "\n".join(lines) + "\n"


def comparison_report(
    baseline: dict[str, list[float]],
    experiment: dict[str, list[float]],
    baseline_label: str = "baseline",
    experiment_label: str = "experiment",
) -> str:
    """Generate markdown report comparing two benchmark runs."""
    lines = [
        "# Benchmark Comparison Report\n",
        f"- **Baseline:** {baseline_label}",
        f"- **Experiment:** {experiment_label}\n",
    ]

    lines.append("| Variant | Base Mean | Exp Mean | Δ% | Cohen's d | Magnitude | KS D | p-value | Confidence |")
    lines.append("|---------|-----------|----------|-----|-----------|-----------|------|---------|------------|")

    all_names = sorted(set(baseline.keys()) & set(experiment.keys()))

    for name in all_names:
        base = baseline[name]
        exp = experiment[name]

        base_mean = sum(base) / len(base)
        exp_mean = sum(exp) / len(exp)
        delta_pct = ((exp_mean - base_mean) / base_mean) * 100

        d = cohens_d(base, exp)
        D, p = ks_test(base, exp)

        lines.append(
            f"| `{name}` | {format_ns(base_mean)} | {format_ns(exp_mean)} "
            f"| {delta_pct:+.1f}% | {d:+.3f} | {effect_magnitude(d)} "
            f"| {D:.4f} | {p:.4f} | {confidence_level(p)} |"
        )

    # Per-variant detail
    lines.append("\n## Per-Variant Detail\n")
    for name in all_names:
        base = baseline[name]
        exp = experiment[name]
        base_stats = percentiles(base)
        exp_stats = percentiles(exp)
        d = cohens_d(base, exp)
        D, p = ks_test(base, exp)

        lines.append(f"### `{name}`\n")
        lines.append(f"- **Cohen's d:** {d:+.3f} ({effect_magnitude(d)})")
        lines.append(f"- **KS test:** D={D:.4f}, p={p:.4f} ({confidence_level(p)})\n")
        lines.append("| Metric | Baseline | Experiment |")
        lines.append("|--------|----------|------------|")
        for metric in ["n", "min", "max", "mean", "std", "p50", "p95", "p99"]:
            bv = base_stats[metric]
            ev = exp_stats[metric]
            if metric == "n":
                lines.append(f"| {metric} | {bv:.0f} | {ev:.0f} |")
            else:
                lines.append(f"| {metric} | {format_ns(bv)} | {format_ns(ev)} |")
        lines.append("")

    return "\n".join(lines) + "\n"


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(1)

    path1 = Path(sys.argv[1])
    text1 = path1.read_text()
    results1 = parse_benchmark_output(text1)

    if len(sys.argv) >= 3:
        path2 = Path(sys.argv[2])
        text2 = path2.read_text()
        results2 = parse_benchmark_output(text2)

        label1 = path1.stem
        label2 = path2.stem
        report = comparison_report(results1, results2, label1, label2)
    else:
        report = single_run_report(results1)

    print(report)


if __name__ == "__main__":
    main()
