# Synthetic line generator

generate.py writes data/line-1.csv, GeoLineAI's synthetic DCIP survey line. The application never runs it: the CSV is committed, and setup imports that file. Regenerate only to change the dataset.

## What it models

1. **Physics.** A SimPEG 2D DC and IP forward model of one conductive, chargeable body (25 ohm-m, chargeability 0.06, 120 m wide, 40 m to 100 m deep, centred under electrode 26) in a resistive background (250 ohm-m, chargeability 0.004). It gives the true apparent resistivity and chargeability of every dipole-dipole geometry: 40 electrodes 20 m apart, a = 1, n factor 1 to 8.
2. **Field conditions.** A seeded pass simulates acquisition: 4 to 8 repeats per geometry, 2 to 8 stacks, contact resistance per electrode, transmitter current from a fixed voltage, receiver noise and the reported error it implies, exponential decay curves whose mean equals the chargeability, timestamps and coordinates.
3. **Defects.** Seven field mechanisms, each applied across the line by a stated rule, never to a chosen row: a drying electrode, swapped receiver leads, a wind-noise window, inductive coupling at large n, a frozen instrument, GPS dropouts and a logger restart. The constants under DEFECT RULES in generate.py define each one.

## Running it

```bash
uv venv -p 3.12 scripts/generate-line/.venv
uv pip install -p scripts/generate-line/.venv/bin/python -r scripts/generate-line/requirements.txt
scripts/generate-line/.venv/bin/python scripts/generate-line/generate.py --out data/line-1.csv
```

The output is byte-identical for the same seed and the pinned requirements; the committed file's SHA-256 is e2a19e173871d5a18b7a950fcd6ff160b3cfabfc2b181782f54b91dbf974c0db.

## After regenerating

The line's numbers appear in the documentation and the tests. Print them with `python3 -I scripts/dataset-facts.py data/line-1.csv`, then update docs/research.md and docs/domain.md, rebuild the citation passages the explanations quote, re-pin the backend acceptance test, and re-check that data/scenarios.json still names the first row meeting each scenario rule in docs/spec.md.
