# Calculation values: kinds, evidence and choices

Every value a calculation, a validation or a limit uses, grouped by where it comes from, with the evidence behind it and why that value or range was chosen. Then the alternatives not used, and what the research supports. [domain.md](domain.md) states each rule; this file holds its basis.

The evidence comes from the shipped line, data/line-1.csv (1,603 readings), and from the domain rules D1 to D6 in domain.md. The line is synthetic, and its generator was written with these limits in mind: where a gap separates sound readings from defective ones, the gap exists by construction, not as a discovery about field data. No limit here is a published geophysical standard. Section 8 reproduces every number.

## 1. Domain rules

Values a domain rule states or supports.

| Value | Rule | Evidence | Why this value |
| --- | --- | --- | --- |
| Apparent resistivity above 0 ohm-m | resistivity_sign is Violated at or below 0 | D1: apparent resistivity at or below zero is physically implausible. 23 readings of the line are negative, all from one run of swapped receiver leads. 9 of them have two or more positive repeats to compare against, and all 9 disagree with them. | 0 is the physical boundary D1 names. Counting exactly 0 as a violation is our extension (section 4). |
| Decay slope below 0 | decay_trend is Violated at or above 0 | D3: a decay falls overall, and a fitted line with a negative slope is a sound first test. 46 curves of the line rise; 25 of them belong to readings whose chargeability is negative, where noise pushed the curve below zero so that it climbs back toward it. The slopes of falling and rising curves meet near 0, with no empty gap between. | 0 separates falling from rising by definition. The method is an unweighted least-squares line with a free intercept over the ten gate centres (section 7 shows it verified). |

## 2. From a gap in the data

A limit placed inside an empty stretch of the line's distribution, so every reading on one side is ordinary and every reading on the other carries a defect. On a synthetic line, the gap is a property of the generator.

| Value | Rule | Evidence | Why this value |
| --- | --- | --- | --- |
| Transmitter contact resistance 0 to 10,000 ohm, inclusive | transmitter_contact | Every contact resistance is either at most 1,091.6 ohm or at least 20,077.1 ohm: 19 readings lie above 10,000 ohm, all transmitting through line-1-e31 after it dried out; none is negative. A negative contact resistance is impossible, so it violates too. | Any limit from 1,091.6 to 20,077.1 flags the same 19 readings; 10,000 sits inside the gap. |

## 3. Provisional

A limit with no gap in the data to place it in: ordinary and defective values run into each other, so the data cannot say where one ends and the other begins. The value stands as a stated choice (D5).

| Value | Rule | Evidence | Why this value |
| --- | --- | --- | --- |
| Resistivity relative error at most 0.50 | Reported error / absolute resistivity | No gap: 1,460 readings below 0.05, 57 from 0.05 to 0.15, 52 from 0.15 to 0.50, 19 from 0.50 to 1.0 and 15 above 1.0. The nearest values straddle the limit: 0.4714 below, 0.5247 above. The large errors come from the wind window and from receiving on the dried electrode. | At 0.50 the reported uncertainty is half the value, which leaves the reading too uncertain to trust its magnitude. That is our reasoning, not a finding. |
| Chargeability relative error at most 1.00 | Reported error / absolute chargeability | No gap: 1,364 readings below 0.2, 136 from 0.2 to 0.4, then a thinning tail (24, 14 and 13 across 0.4 to 1.0), 52 above 1.0. The nearest values straddle the limit: 0.9937 below, 1.0020 above. | At 1.00 the reported error equals the value itself, so the measurement carries no information beyond its own uncertainty. That is our reasoning, not a finding: the data offers no better place. |
| Repeat agreement within 0.20 of the reference | repeat_agreement; Review above | No clean gap. Of 1,531 readings compared: 1,503 within 0.15, 4 from 0.15 to 0.20, 13 from 0.20 to 0.40 and 11 beyond 0.40. Of the 24 above 0.20, 19 use the suspect electrode or have a Violated single-reading result; the other 5 have neither. Four of those lie inside the wind window; the fifth, at row 1536, is ordinary measurement noise at n factor 8 over the conductive body, 26% from its repeats' median: a sound reading the check sends to review. D4 names outlier statistics as the first test but gives no number. | Ordinary disagreement stays below 0.15. 0.20 leaves room above that for a sound reading's noise, and still flags a reading a fifth away from what its repeats agree on. |

## 4. Design choices

Rules the domain rules and the data leave open, decided for a stated reason.

| Choice | Why |
| --- | --- |
| Resistivity of exactly 0 violates the sign criterion | A zero apparent resistivity is no more physical than a negative one. No reading of the line is 0 (the smallest positive is 14.712), so the data never tests it. |
| A relative error is not computed when the value is 0, and divides by the absolute value | Dividing by 0 has no answer, and a negative value would make the ratio negative. A zero resistivity is judged by its sign criterion alone; a zero chargeability by no criterion, since no domain rule says one is wrong. |
| The repeat-agreement reference needs at least 2 eligible repeats, a median above 0, and excludes repeats with a Violated single-reading result | One repeat is not a reference; a median at or below 0 cannot be divided by; a repeat already known to be bad would drag the reference toward itself. |
| A median of an even count is the midpoint of the two middle values | The standard definition: [1, 2, 3, 4] has median 2.5. |
| Frozen repeat means resistivity, chargeability and all ten decay samples exactly equal, compared only between repeats with every value available | Independent measurements never agree to every digit. The line's 5 frozen readings are one whole repeat group, taken within 40 seconds; nothing looser was needed to find them, and a partial comparison would call readings frozen on missing data. |
| A suspect electrode needs failing transmitter contact with at least 2 distinct partners, counting the tx1 and tx2 slots | Three electrodes have failing contact. line-1-e31 fails with two partners, its neighbours on both sides; line-1-e30 and line-1-e32 fail with one partner each, always e31. A count of 2 separates the faulty electrode from its neighbours exactly. D6 supports looking at shared electrodes; it gives no count. Contact is measured on the transmitting pair, so only tx slots count. |
| Electrode IDs are not checked when a row's line_id is invalid | They are checked against the row's own line, which then does not exist; the row is rejected either way, and fixing line_id shows any real electrode error on reimport. |
| Timestamps must fall in UTC years 0000 to 9999 with at most nine fractional digits | Stored times are fixed-width text so that text order is time order; outside that range the format breaks. |

## 5. Dataset contract

Facts of the data format, enforced so a value is never stored dishonestly.

| Value | Evidence | Why enforced this way |
| --- | --- | --- |
| Ten decay samples, at gate centres 24.32, 35.96, 53.18, 78.64, 116.30, 171.97, 254.31, 376.06, 556.10 and 822.34 ms | The acquisition's fixed gates | The slope fit needs one sample per gate. |
| Electrode index 1 to 40 on the reading's own line; dipole-dipole array | The survey layout | Repeat groups and the suspect-electrode search treat the ID as an electrode of the line; the criteria assume dipole-dipole geometry. |
| Stack count and n factor at least 1 | The line: stack count 2 to 8, n factor 1 to 8 | 0 or a negative count is meaningless. No upper bound: the data's maximum describes this survey, not a limit on the next. |
| Reported errors at least 0 | - | A negative error would read as a met relative-error criterion. |
| Coordinates finite, with no range check | Every latitude is between 34.5276 and 34.5300 and every longitude between -115.7300 and -115.7220: one small survey area | A bound no value comes near has no basis in the data, and no criterion reads coordinates. |

## 6. Chosen not to use

| Alternative | Why not |
| --- | --- |
| Sign criteria for chargeability or input current | [ADR 0003](adr/0003-deterministic-advisory.md), Considered options. |
| Negative resistivity as Review rather than Fail | [ADR 0003](adr/0003-deterministic-advisory.md), Considered options. |
| A local decay rise: one sample more than a ratio above the one before | Rises are common: 675 of the 1,557 falling curves have at least one. Ratio limits 1.10, 1.25, 1.50 and 2.00 flag 28, 21, 0 and 0 readings, of which 7, 0, 0 and 0 are not already rising curves. No labelled truth says the extra ones are defects. |
| A log-space decay fit error | A rising or sign-reversed exponential fits as well as a falling one, so a low error cannot establish a good decay. |
| Suspect electrodes by the share of their readings that fail | [ADR 0003](adr/0003-deterministic-advisory.md), Considered options. |
| Repeat disagreement as Fail | [ADR 0003](adr/0003-deterministic-advisory.md), Considered options. |
| Model inference for any advisory | [ADR 0003](adr/0003-deterministic-advisory.md), Considered options. |

## 7. What the research supports

The research supports methods; it validates none of the thresholds above.

| Finding | What it supports |
| --- | --- |
| The handwritten least-squares slope matched an exact-rational reference on all 1,603 curves (largest difference 5.4e-20), and so did numpy's polyfit (8.1e-20); neither disagreed with the reference on the sign of any slope. | The decay-trend method, verified two independent ways. |
| Gonum's quantile returns 2 for [1, 2, 3, 4], where the specified median is 2.5. | A handwritten median. |
| No public source validates 0.50, 1.00, 10,000 ohm or 0.20 for this survey. | Calling those limits data-based or provisional, never standards. |

## 8. Reproducing the numbers

Every number in this file about the line comes from the standard-library script, run from the repository root:

```bash
python3 -I scripts/dataset-facts.py data/line-1.csv
```

The backend acceptance test (backend/internal/server/acceptance_test.go) pins the advisory, criterion, repeat-group and import-issue counts the application itself computes. The slope verification in section 7 needs numpy, which the generator's environment holds (scripts/generate-line/README.md).
