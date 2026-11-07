"""Prints the SCRUM-109 numbers from a JMeter results file (CSV .jtl).

    python3 tests/load/summarize.py tests/load/out/<time>/results.jtl

For each request: count, errors, average, median, 90th, 95th and 99th
percentile, maximum (ms), throughput, and the HTTP codes seen. Then whether the
agreed target is met: 95% of POST /reports under 500 ms and under 1% errors.
"""

import collections
import csv
import statistics
import sys

P95_TARGET_MS = 500
ERROR_TARGET_PCT = 1.0


def percentile(values, p):
    values = sorted(values)
    k = (len(values) - 1) * p / 100
    f = int(k)
    c = min(f + 1, len(values) - 1)
    return values[f] + (values[c] - values[f]) * (k - f)


def main(path):
    by_label = collections.defaultdict(list)
    with open(path, newline="") as f:
        for row in csv.DictReader(f):
            by_label[row["label"]].append(row)

    result = None
    for label, rows in by_label.items():
        times = [int(r["elapsed"]) for r in rows]
        errors = sum(r["success"] != "true" for r in rows)
        stamps = [int(r["timeStamp"]) for r in rows]
        seconds = max((max(stamps) - min(stamps)) / 1000, 1)
        codes = collections.Counter(r["responseCode"] for r in rows)
        error_pct = 100 * errors / len(rows)
        p95 = percentile(times, 95)
        print(f"{label}")
        print(f"  requests {len(rows)}, errors {errors} ({error_pct:.2f}%), throughput {len(rows) / seconds:.1f}/s")
        print(
            f"  ms: average {statistics.mean(times):.1f}, median {percentile(times, 50):.0f}, "
            f"p90 {percentile(times, 90):.0f}, p95 {p95:.0f}, p99 {percentile(times, 99):.0f}, max {max(times)}"
        )
        print(f"  HTTP codes: {dict(sorted(codes.items()))}")
        if label == "POST /reports":
            result = (p95, error_pct)

    if result is None:
        print("No POST /reports samples found")
        return 1
    p95, error_pct = result
    met = p95 < P95_TARGET_MS and error_pct < ERROR_TARGET_PCT
    print()
    print(
        f"Target (p95 < {P95_TARGET_MS} ms and errors < {ERROR_TARGET_PCT}%): "
        f"p95 {p95:.0f} ms, errors {error_pct:.2f}% -> {'MET' if met else 'NOT MET'}"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1]))
