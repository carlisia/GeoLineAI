"""Generate GeoLineAI's synthetic DCIP survey line, data/line-1.csv.

The line is produced in two stages, both deterministic for a given seed:

1. Physics. A SimPEG 2D DC and IP forward model of one conductive, chargeable body in
   a resistive background gives the true apparent resistivity and apparent
   chargeability of every dipole-dipole geometry (a = 1, n = 1 to 8, 40 electrodes).

2. Field conditions. A seeded pass simulates acquisition: repeat counts, stacks,
   contact resistance, transmitter current, measurement noise and its reported error,
   decay curves, timestamps and coordinates. It then applies the defect rules listed
   in DEFECT RULES below, each across the whole line by a stated mechanism and rate.
   No row is edited individually.

Usage, from the repository root:

    scripts/generate-line/.venv/bin/python scripts/generate-line/generate.py \
        --out data/line-1.csv

The output is byte-identical for the same seed and pinned dependencies.
"""

from __future__ import annotations

import argparse
import csv
import io
import math
from dataclasses import dataclass, field
from datetime import datetime, timedelta, timezone

import numpy as np
from discretize import TensorMesh
from simpeg import maps
from simpeg.utils import get_default_solver
from simpeg.electromagnetics.static import induced_polarization as ip
from simpeg.electromagnetics.static import resistivity as dc
from simpeg.electromagnetics.static.utils.static_utils import (
    apparent_resistivity_from_voltage,
)

SEED = 20261001

# --- Survey facts the application's code fixes -----------------------------------

LINE = "line-1"
ELECTRODES = 40
N_LEVELS = range(1, 9)
DECAY_GATES_MS = np.array(
    [24.32, 35.96, 53.18, 78.64, 116.30, 171.97, 254.31, 376.06, 556.10, 822.34]
)

# --- Survey choices ----------------------------------------------------------------

SPACING_M = 20.0
BEARING_DEG = 110.0
# First electrode, in open desert east of Amboy, California. Flat ground.
ORIGIN_LAT = 34.5300000
ORIGIN_LON = -115.7300000
ELEVATION_M = 192.0
# 08:30 local time (PDT, UTC-7) on 2026-10-01.
START = datetime(2026, 10, 1, 15, 30, 0, tzinfo=timezone.utc)
CADENCE_S = 10
REPEATS = (4, 8)  # inclusive range of readings per geometry
STACKS = (2, 8)  # inclusive range of stacks per reading

# --- Earth model -------------------------------------------------------------------

BACKGROUND_RHO = 250.0  # ohm-m
BACKGROUND_ETA = 0.004  # V/V
BODY_RHO = 25.0
BODY_ETA = 0.06
BODY_CENTRE_ELECTRODE = 26
BODY_WIDTH_M = 120.0
BODY_TOP_M = 40.0
BODY_BOTTOM_M = 100.0

# --- Instrument --------------------------------------------------------------------

TX_VOLTAGE = 1000.0  # V; current = voltage / contact resistance, clamped below
CURRENT_RANGE = (0.1, 2.5)  # A
RHO_ERR_FLOOR = 0.015  # fractional error that remains at any signal
VOLTAGE_NOISE = 4e-4  # V; per-stack receiver noise
ETA_ERR_FLOOR = 0.0003  # V/V
ETA_NOISE = 1e-5  # V/V . V
DECAY_TAU_MS = 400.0
DECAY_TAU_SPREAD = 0.10  # +/- fraction, uniform per reading
DECAY_GATE_NOISE = 0.03  # fractional, per gate
ELECTRODE_CONTACT_MEDIAN = 300.0  # ohm, per electrode
ELECTRODE_CONTACT_SIGMA = 0.35  # lognormal sigma
PAIR_CONTACT_JITTER = 0.08  # fractional, per reading

# --- DEFECT RULES ------------------------------------------------------------------
#
# Each rule is a field mechanism applied across the line. Rates are probabilities or
# fixed windows chosen once from the seed.
#
# 1. Dry electrode. After the lunch break, electrode e31 dries out: its contact
#    resistance climbs past 10,000 ohm and keeps rising. Any reading transmitting
#    through e31 has failing contact and low current. Readings that receive on e31
#    pick up extra noise.
DRY_ELECTRODE = 31
DRY_CONTACT = (12000.0, 30000.0)  # ohm, rising linearly over the afternoon
DRY_RX_NOISE_FACTOR = 25.0
# 2. Swapped receiver leads. Once, the crew swaps the receiver leads and notices only
#    after SWAP_RUN readings; those readings have inverted voltage (negative apparent
#    resistivity). Chargeability is a ratio of two inverted voltages and is unchanged.
SWAP_RUN = 14
# 3. Wind noise. One window of WIND_RUN consecutive readings has receiver noise
#    multiplied by WIND_FACTOR.
WIND_RUN = 40
WIND_FACTOR = 40.0
# 4. Inductive coupling. At n >= 7 the long cables couple inductively; with
#    probability COUPLING_RATE a reading's decay gains a rising ramp.
COUPLING_RATE = 0.04
COUPLING_MIN_N = 7
# 5. Frozen instrument. Once, the instrument hangs for one repeat group: every
#    reading in that group repeats the first reading's measured values.
# 6. GPS dropout. GPS_DROPOUTS readings lose one electrode's coordinates.
GPS_DROPOUTS = 6
# 7. Logger restart. At the equipment-swap break the logger restarts from its last
#    checkpoint and reissues the previous RESTART_REUSE reading IDs.
RESTART_REUSE = 4

LUNCH_AFTER_LOCAL = (12, 0)  # the first break starts once local time passes 12:00
LUNCH_MINUTES = 45
SWAP_BREAK_FRACTION = 0.8  # equipment swap after this fraction of readings
SWAP_BREAK_MINUTES = 20
LOCAL_OFFSET = timedelta(hours=-7)

COLUMNS = [
    "id", "line_id",
    "tx1_id", "tx1_lat", "tx1_lon", "tx1_alt",
    "tx2_id", "tx2_lat", "tx2_lon", "tx2_alt",
    "rx1_id", "rx1_lat", "rx1_lon", "rx1_alt",
    "rx2_id", "rx2_lat", "rx2_lon", "rx2_alt",
    "timestamp", "array_type", "symmetry", "stacks", "input_current",
    "contact_resistance", "apparent_resistivity", "apparent_resistivity_err",
    "chargeability", "chargeability_err", "decay_curve", "qc_review",
]


@dataclass(frozen=True)
class Geometry:
    tx1: int
    n: int

    @property
    def tx2(self) -> int:
        return self.tx1 + 1

    @property
    def rx1(self) -> int:
        return self.tx2 + self.n

    @property
    def rx2(self) -> int:
        return self.rx1 + 1

    @property
    def electrodes(self) -> tuple[int, int, int, int]:
        return (self.tx1, self.tx2, self.rx1, self.rx2)


def geometries() -> list[Geometry]:
    """Acquisition order: each n level rolls along the whole line before the next."""
    out = []
    for n in N_LEVELS:
        for tx1 in range(1, ELECTRODES - n - 1):
            out.append(Geometry(tx1, n))
    return out


def x_of(electrode: int) -> float:
    return (electrode - 1) * SPACING_M


def forward_model(geoms: list[Geometry]) -> tuple[np.ndarray, np.ndarray]:
    """Return true apparent resistivity (ohm-m) and chargeability (V/V) per geometry."""
    core = 2.5
    length = x_of(ELECTRODES)
    ncx = int(length / core) + 40
    hx = [(core, 16, -1.3), (core, ncx), (core, 16, 1.3)]
    hz = [(core, 16, -1.3), (core, int(220 / core))]
    mesh = TensorMesh([hx, hz], origin="CN")
    mesh.origin = np.r_[-mesh.h[0][:16].sum() - 20 * core, mesh.origin[1]]

    cc = mesh.cell_centers
    centre = x_of(BODY_CENTRE_ELECTRODE)
    body = (
        (np.abs(cc[:, 0] - centre) <= BODY_WIDTH_M / 2)
        & (cc[:, 1] <= -BODY_TOP_M)
        & (cc[:, 1] >= -BODY_BOTTOM_M)
    )
    rho = np.full(mesh.n_cells, BACKGROUND_RHO)
    rho[body] = BODY_RHO
    eta = np.full(mesh.n_cells, BACKGROUND_ETA)
    eta[body] = BODY_ETA

    def survey(data_type: str):
        sources = []
        for g in geoms:
            m = np.array([[x_of(g.rx1), 0.0]])
            nn = np.array([[x_of(g.rx2), 0.0]])
            rx = dc.receivers.Dipole(m, nn, data_type=data_type)
            sources.append(
                dc.sources.Dipole(
                    [rx], np.r_[x_of(g.tx1), 0.0], np.r_[x_of(g.tx2), 0.0]
                )
            )
        return dc.Survey(sources)

    solver = get_default_solver()
    dc_survey = survey("volt")
    dc_sim = dc.Simulation2DNodal(
        mesh, survey=dc_survey, rhoMap=maps.IdentityMap(mesh), solver=solver
    )
    volts = dc_sim.dpred(rho)
    rho_app = apparent_resistivity_from_voltage(dc_survey, volts)

    ip_survey = survey("volt")
    ip_sim = ip.Simulation2DNodal(
        mesh, survey=ip_survey, etaMap=maps.IdentityMap(mesh), rho=rho, solver=solver
    )
    secondary = ip_sim.dpred(eta)
    eta_app = secondary / volts
    return np.asarray(rho_app), np.asarray(eta_app)


def geometric_factor(g: Geometry) -> float:
    a = SPACING_M
    return math.pi * a * g.n * (g.n + 1) * (g.n + 2)


def coords(electrode: int) -> tuple[float, float]:
    d = x_of(electrode)
    b = math.radians(BEARING_DEG)
    earth = 6371008.8
    lat = ORIGIN_LAT + math.degrees(d * math.cos(b) / earth)
    lon = ORIGIN_LON + math.degrees(
        d * math.sin(b) / (earth * math.cos(math.radians(ORIGIN_LAT)))
    )
    return lat, lon


@dataclass
class Reading:
    geom: Geometry
    time: datetime
    stacks: int
    contact: float = 0.0
    current: float = 0.0
    rho: float = 0.0
    rho_err: float = 0.0
    eta: float = 0.0
    eta_err: float = 0.0
    decay: list[float] = field(default_factory=list)
    id_number: int = 0
    blank_slot: int | None = None  # 0..3: tx1, tx2, rx1, rx2


def simulate(seed: int) -> list[Reading]:
    rng = np.random.default_rng(seed)
    geoms = geometries()
    rho_true, eta_true = forward_model(geoms)
    truth = {g: (r, e) for g, r, e in zip(geoms, rho_true, eta_true)}

    # Acquisition plan: repeat counts, stacks and timestamps.
    plan: list[tuple[Geometry, int]] = []
    for g in geoms:
        count = int(rng.integers(REPEATS[0], REPEATS[1] + 1))
        for _ in range(count):
            plan.append((g, int(rng.integers(STACKS[0], STACKS[1] + 1))))

    readings: list[Reading] = []
    t = START
    lunch_taken = False
    swap_index = int(len(plan) * SWAP_BREAK_FRACTION)
    lunch_index = None
    for i, (g, stacks) in enumerate(plan):
        local = t + LOCAL_OFFSET
        if not lunch_taken and (local.hour, local.minute) >= LUNCH_AFTER_LOCAL:
            t += timedelta(minutes=LUNCH_MINUTES)
            lunch_taken = True
            lunch_index = i
        if i == swap_index:
            t += timedelta(minutes=SWAP_BREAK_MINUTES)
        readings.append(Reading(geom=g, time=t, stacks=stacks, id_number=i + 1))
        t += timedelta(seconds=CADENCE_S)
    assert lunch_index is not None, "survey ended before lunch"
    afternoon_start = readings[lunch_index].time
    afternoon_end = readings[-1].time

    # Per-electrode contact resistance (ohm), fixed for the day.
    contact = ELECTRODE_CONTACT_MEDIAN * np.exp(
        rng.normal(0.0, ELECTRODE_CONTACT_SIGMA, ELECTRODES + 1)
    )

    def electrode_contact(e: int, when: datetime) -> float:
        if e == DRY_ELECTRODE and when >= afternoon_start:
            span = (afternoon_end - afternoon_start).total_seconds() or 1.0
            frac = (when - afternoon_start).total_seconds() / span
            return DRY_CONTACT[0] + frac * (DRY_CONTACT[1] - DRY_CONTACT[0])
        return float(contact[e])

    # One-off windows, chosen from the seed.
    swap_start = int(rng.integers(len(readings) // 10, len(readings) // 3))
    wind_start = int(rng.integers(len(readings) // 2, len(readings) - WIND_RUN))

    for i, r in enumerate(readings):
        g = r.geom
        pair = electrode_contact(g.tx1, r.time) + electrode_contact(g.tx2, r.time)
        r.contact = pair * (1.0 + rng.normal(0.0, PAIR_CONTACT_JITTER))
        r.current = float(np.clip(TX_VOLTAGE / r.contact, *CURRENT_RANGE))

        rho_t, eta_t = truth[g]
        signal = rho_t * r.current / geometric_factor(g)  # V at the receiver
        noise = VOLTAGE_NOISE
        if wind_start <= i < wind_start + WIND_RUN:
            noise *= WIND_FACTOR
        if DRY_ELECTRODE in (g.rx1, g.rx2) and r.time >= afternoon_start:
            noise *= DRY_RX_NOISE_FACTOR
        root = math.sqrt(r.stacks)
        rel_err = RHO_ERR_FLOOR + noise / (root * signal)
        eta_err = ETA_ERR_FLOOR + ETA_NOISE * (noise / VOLTAGE_NOISE) / (root * signal)

        r.rho = float(rho_t * (1.0 + rng.normal(0.0, rel_err)))
        r.rho_err = rel_err * abs(r.rho)
        r.eta = float(eta_t + rng.normal(0.0, eta_err))
        r.eta_err = eta_err

        tau = DECAY_TAU_MS * (1.0 + rng.uniform(-DECAY_TAU_SPREAD, DECAY_TAU_SPREAD))
        shape = np.exp(-DECAY_GATES_MS / tau)
        gates = shape * (r.eta / shape.mean())
        gates = gates * (1.0 + rng.normal(0.0, DECAY_GATE_NOISE, gates.size))
        if g.n >= COUPLING_MIN_N and rng.random() < COUPLING_RATE:
            ramp = DECAY_GATES_MS / DECAY_GATES_MS[-1]
            gates = gates + 2.5 * abs(r.eta) * ramp
        r.decay = [float(v) for v in gates]

        if swap_start <= i < swap_start + SWAP_RUN:
            r.rho = -r.rho

    # Frozen instrument: one repeat group, chosen among groups of five or more.
    groups: dict[Geometry, list[int]] = {}
    for i, r in enumerate(readings):
        groups.setdefault(r.geom, []).append(i)
    candidates = [g for g, idx in groups.items() if len(idx) >= 5]
    frozen = candidates[int(rng.integers(len(candidates)))]
    first = readings[groups[frozen][0]]
    for i in groups[frozen][1:]:
        r = readings[i]
        r.rho, r.rho_err = first.rho, first.rho_err
        r.eta, r.eta_err = first.eta, first.eta_err
        r.decay = list(first.decay)

    # GPS dropouts.
    for i in rng.choice(len(readings), size=GPS_DROPOUTS, replace=False):
        readings[int(i)].blank_slot = int(rng.integers(0, 4))

    # Logger restart at the equipment-swap break.
    for k in range(RESTART_REUSE):
        readings[swap_index + k].id_number = swap_index - RESTART_REUSE + k + 1

    return readings


def fmt(value: float, places: int) -> str:
    text = f"{value:.{places}f}"
    return "0." + "0" * places if text == "-0." + "0" * places else text


def write(readings: list[Reading]) -> str:
    out = io.StringIO()
    w = csv.writer(out, lineterminator="\n")
    w.writerow(COLUMNS)
    for r in readings:
        row = [f"{LINE}-r{r.id_number:05d}", LINE]
        for slot, e in enumerate(r.geom.electrodes):
            lat, lon = coords(e)
            if r.blank_slot == slot:
                row += [f"{LINE}-e{e:02d}", "", "", ""]
            else:
                row += [f"{LINE}-e{e:02d}", fmt(lat, 7), fmt(lon, 7), fmt(ELEVATION_M, 1)]
        row += [
            r.time.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "dipole-dipole",
            str(r.geom.n),
            str(r.stacks),
            fmt(r.current, 3),
            fmt(r.contact, 1),
            fmt(r.rho, 3),
            fmt(r.rho_err, 3),
            fmt(r.eta, 5),
            fmt(r.eta_err, 5),
            ";".join(fmt(v, 5) for v in r.decay),
            "",
        ]
        w.writerow(row)
    return out.getvalue()


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--out", required=True)
    parser.add_argument("--seed", type=int, default=SEED)
    args = parser.parse_args()
    text = write(simulate(args.seed))
    with open(args.out, "w", newline="") as f:
        f.write(text)
    print(f"wrote {args.out}: {text.count(chr(10)) - 1} readings")


if __name__ == "__main__":
    main()
