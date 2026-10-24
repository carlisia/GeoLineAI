"""Print the dataset facts that docs/research.md and docs/domain.md cite.

Every number those documents state about data/line-1.csv is printed here, computed
from the CSV by the rules the criteria document, so a regenerated line or a changed
rule shows which sentences to update. The Go acceptance test pins the counts the
application computes; this script is the independent reading.

Run from the repository root with the standard library only:

    python3 -I scripts/dataset-facts.py data/line-1.csv
"""

import collections
import csv
import statistics
import sys
from datetime import datetime

GATES = [24.32, 35.96, 53.18, 78.64, 116.30, 171.97, 254.31, 376.06, 556.10, 822.34]

# The criteria's limits, written out so this reading stays independent of the code; a
# backend test pins each to the value the criteria compare against.
RESISTIVITY_ERROR_LIMIT = 0.50
CHARGEABILITY_ERROR_LIMIT = 1.00
CONTACT_MAX = 10000
REPEAT_LIMIT = 0.20


def num(text):
    try:
        return float(text)
    except ValueError:
        return None


def slope(samples):
    mx, my = sum(GATES) / 10, sum(samples) / 10
    return sum((g - mx) * (v - my) for g, v in zip(GATES, samples)) / sum((g - mx) ** 2 for g in GATES)


def band(values, edges):
    """Counts of values in [edges[i], edges[i+1]) for each consecutive pair."""
    return [sum(lo <= v < hi for v in values) for lo, hi in zip(edges, edges[1:])]


def main(path):
    rows = list(csv.DictReader(open(path, newline="")))
    out = {}
    readings = []
    for i, r in enumerate(rows):
        d = {
            "row": i + 2,
            "id": r["id"],
            "geo": tuple(r[s + "_id"] for s in ("tx1", "tx2", "rx1", "rx2")),
            "time": r["timestamp"],
            "rho": num(r["apparent_resistivity"]),
            "rho_err": num(r["apparent_resistivity_err"]),
            "m": num(r["chargeability"]),
            "m_err": num(r["chargeability_err"]),
            "contact": num(r["contact_resistance"]),
            "decay": [float(x) for x in r["decay_curve"].split(";")],
            "n": int(r["symmetry"]),
            "stacks": int(r["stacks"]),
        }
        d["slope"] = slope(d["decay"])
        d["rho_rel"] = d["rho_err"] / abs(d["rho"]) if d["rho"] else None
        d["m_rel"] = d["m_err"] / abs(d["m"]) if d["m"] else None
        readings.append(d)

    out["readings"] = len(readings)
    out["n range"] = (min(d["n"] for d in readings), max(d["n"] for d in readings))
    out["stacks range"] = (min(d["stacks"] for d in readings), max(d["stacks"] for d in readings))
    lats = [num(r[s + "_lat"]) for r in rows for s in ("tx1", "tx2", "rx1", "rx2") if num(r[s + "_lat"]) is not None]
    lons = [num(r[s + "_lon"]) for r in rows for s in ("tx1", "tx2", "rx1", "rx2") if num(r[s + "_lon"]) is not None]
    out["latitude range"] = (min(lats), max(lats))
    out["longitude range"] = (min(lons), max(lons))

    # Single-reading criteria.
    for d in readings:
        d["violated"] = {
            "resistivity_sign": d["rho"] <= 0,
            "resistivity_relative_error": d["rho_rel"] is not None and d["rho_rel"] > RESISTIVITY_ERROR_LIMIT,
            "chargeability_relative_error": d["m_rel"] is not None and d["m_rel"] > CHARGEABILITY_ERROR_LIMIT,
            "transmitter_contact": d["contact"] < 0 or d["contact"] > CONTACT_MAX,
            "decay_trend": d["slope"] >= 0,
        }
    groups = collections.defaultdict(list)
    for d in readings:
        groups[d["geo"]].append(d)
    for d in readings:
        d["frozen"] = any(o is not d and o["rho"] == d["rho"] and o["m"] == d["m"] and o["decay"] == d["decay"] for o in groups[d["geo"]])
        d["left_out"] = any(d["violated"].values())

    negatives = [d for d in readings if d["rho"] <= 0]
    out["resistivity at or below 0"] = len(negatives)
    out["resistivity exactly 0"] = sum(d["rho"] == 0 for d in readings)
    out["smallest positive resistivity"] = min(d["rho"] for d in readings if d["rho"] > 0)

    def against_positive_repeats(d):
        others = [o["rho"] for o in groups[d["geo"]] if o is not d and o["rho"] > 0]
        return len(others) >= 2 and abs(d["rho"] - statistics.median(others)) / statistics.median(others) > REPEAT_LIMIT

    def has_positive_repeats(d):
        return sum(o is not d and o["rho"] > 0 for o in groups[d["geo"]]) >= 2

    out["negative readings with 2 or more positive repeats"] = sum(has_positive_repeats(d) for d in negatives)
    out["negative readings disagreeing with their positive repeats"] = sum(against_positive_repeats(d) for d in negatives)

    rising = [d for d in readings if d["slope"] >= 0]
    falling = [d for d in readings if d["slope"] < 0]
    out["rising decays"] = len(rising)
    out["rising decays with negative chargeability"] = sum(d["m"] < 0 for d in rising)
    out["falling slope closest to 0"] = max(d["slope"] for d in falling)
    out["smallest rising slope"] = min(d["slope"] for d in rising)

    rho_rel = [d["rho_rel"] for d in readings if d["rho_rel"] is not None]
    out["resistivity relative error bands [0,.05,.15,.5,1,inf)"] = band(rho_rel, [0, 0.05, 0.15, 0.5, 1.0, float("inf")])
    out["resistivity relative error: largest at or below 0.50"] = max(x for x in rho_rel if x <= RESISTIVITY_ERROR_LIMIT)
    out["resistivity relative error: smallest above 0.50"] = min(x for x in rho_rel if x > RESISTIVITY_ERROR_LIMIT)

    m_rel = [d["m_rel"] for d in readings if d["m_rel"] is not None]
    out["chargeability relative error bands [0,.2,.4,.6,.8,1,inf)"] = band(m_rel, [0, 0.2, 0.4, 0.6, 0.8, 1.0, float("inf")])
    out["chargeability relative error: largest at or below 1.00"] = max(x for x in m_rel if x <= CHARGEABILITY_ERROR_LIMIT)
    out["chargeability relative error: smallest above 1.00"] = min(x for x in m_rel if x > CHARGEABILITY_ERROR_LIMIT)

    contact = [d["contact"] for d in readings]
    out["contact: negative"] = sum(c < 0 for c in contact)
    out["contact: largest at or below 10,000"] = max(c for c in contact if c <= CONTACT_MAX)
    out["contact: smallest above 10,000"] = min(c for c in contact if c > CONTACT_MAX)
    out["contact: above 10,000"] = sum(c > CONTACT_MAX for c in contact)

    partners = collections.defaultdict(set)
    for d in readings:
        if d["violated"]["transmitter_contact"]:
            partners[d["geo"][0]].add(d["geo"][1])
            partners[d["geo"][1]].add(d["geo"][0])
    out["failing-contact partners per electrode"] = {e: len(p) for e, p in sorted(partners.items())}
    suspect = {e for e, p in partners.items() if len(p) >= 2}
    out["readings using a suspect electrode"] = sum(any(e in suspect for e in d["geo"]) for d in readings)

    frozen = [d for d in readings if d["frozen"]]
    clusters = collections.defaultdict(list)
    for d in frozen:
        clusters[d["geo"]].append(d)
    spans = []
    for members in clusters.values():
        times = sorted(datetime.strptime(d["time"], "%Y-%m-%dT%H:%M:%SZ") for d in members)
        spans.append(((times[-1] - times[0]).seconds, len(members), len(groups[members[0]["geo"]])))
    out["frozen readings"] = len(frozen)
    out["frozen clusters (seconds spanned, readings, group size)"] = spans

    # Repeat agreement, the criterion's own reference: the median of the other repeats
    # not left out, at least 2 of them, a median above 0.
    diffs = []
    for d in readings:
        if d["rho"] is None:
            continue
        ref = [o["rho"] for o in groups[d["geo"]] if o is not d and not o["left_out"]]
        if len(ref) < 2 or statistics.median(ref) <= 0:
            continue
        med = statistics.median(ref)
        diffs.append((abs(d["rho"] - med) / med, d))
    values = [x for x, _ in diffs]
    out["repeat agreement compared"] = len(diffs)
    out["repeat agreement bands [0,.15,.2,.4,inf)"] = band(values, [0, 0.15, 0.2, 0.4, float("inf")])

    def explained(d):
        return any(e in suspect for e in d["geo"]) or d["left_out"]

    out["repeat agreement 0.15 to 0.20: with a suspect electrode or left out"] = sum(0.15 <= x < REPEAT_LIMIT and explained(d) for x, d in diffs)
    out["repeat agreement above 0.20: with a suspect electrode or left out"] = sum(x > REPEAT_LIMIT and explained(d) for x, d in diffs)
    out["repeat agreement above 0.20: neither"] = sum(x > REPEAT_LIMIT and not explained(d) for x, d in diffs)
    out["repeat agreement above 0.20: largest among those explained by neither"] = max((x for x, d in diffs if x > REPEAT_LIMIT and not explained(d)), default=None)
    out["disagreeing readings"] = sum(x > REPEAT_LIMIT for x in values)
    out["groups holding a disagreeing reading"] = len({d["geo"] for x, d in diffs if x > REPEAT_LIMIT})
    out["repeat group sizes"] = dict(sorted(collections.Counter(len(g) for g in groups.values()).items()))
    out["repeat groups"] = len(groups)

    # Alternatives chosen not to use: a local rise between consecutive samples.
    with_rise = sum(any(b > a for a, b in zip(d["decay"], d["decay"][1:])) for d in falling)
    out["falling curves with at least one local rise"] = (with_rise, len(falling))
    for ratio in (1.10, 1.25, 1.50, 2.00):
        flagged = [d for d in readings if any(a > 0 and b > ratio * a for a, b in zip(d["decay"], d["decay"][1:]))]
        out[f"local rise above ratio {ratio:.2f}: flagged, not already rising"] = (len(flagged), sum(d["slope"] < 0 for d in flagged))

    for key, value in out.items():
        print(f"{key}: {value}")


if __name__ == "__main__":
    main(sys.argv[1])
