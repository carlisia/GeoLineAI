# Domain reference

What GeoLineAI checks, the words it uses, the dataset it ships, and how an explanation's citations resolve. [research.md](research.md) holds the evidence and reasoning behind each limit; [spec.md](spec.md) holds the product rules.

## Domain rules

The criteria rest on these rules. They are ordinary properties of DC resistivity and induced-polarization (DCIP) surveys, stated in GeoLineAI's own words; none is a published threshold.

| # | Rule |
| --- | --- |
| D1 | Apparent resistivity at or below zero is physically implausible: the ground cannot produce it. |
| D2 | A reading with an implausible or missing value is kept and flagged rather than dropped, so the reviewer can see that it happened. |
| D3 | An induced-polarization decay falls overall after the current is switched off, though not necessarily at every gate; a straight line fitted with a negative slope is a sound first test. |
| D4 | Comparing a reading with its repeats, an outlier statistic, is a sound first test for a reading that disagrees. |
| D5 | No published formula or threshold fixes these limits for this survey; each one is a stated choice. |
| D6 | Readings that share an electrode with known-bad readings deserve a second look. |

## Terms

- Reading: one retained CSV row, with one verdict.
- Measurement: one measured value on a reading: apparent resistivity, chargeability, a reported error or one decay sample.
- Reported error: the instrument's uncertainty on a measurement, supplied beside it.
- Electrode: a stake with one identifier, used in one of four slots: tx1 and tx2, the transmitting pair, and rx1 and rx2, the receiving pair. Electrode ID: its identifier, such as line-1-e01.
- Acquisition ID: the source id text, kept verbatim and shown only; two readings may share it. The internal reading ID is the application's identity for a reading.
- n factor: the dipole-dipole separation, supplied in the source symmetry column. Stack count: pulses averaged inside one reading; it never counts repeats.
- Import issue: a problem recorded on a retained reading, a missing value or a repeated acquisition ID.
- Criterion result: Met, Violated or Review, with an explanation; a criterion with nothing to compare records no result.
- Advisory: what the results say, decided at import: Fail when any result is Violated, Needs review when none is but some are Review, and Pass otherwise.
- Verdict: the reviewer's judgment on one reading, Pass, Fail or Unreviewed. It is Unreviewed from import, and only an explicit human save changes it.
- A reading's repeat group is the readings from its import that share all four electrode IDs; its repeats are the group's other readings.
- Reference: the median apparent resistivity of a reading's repeats, leaving out repeats without a resistivity and repeats with a Violated single-reading result. Repeat difference: the absolute difference from the reference, divided by the reference.
- A suspect electrode is one whose transmitting readings violate transmitter contact with two or more different partner electrodes; it marks an association to review, not a proven fault.
- Reviewer: the fixed label reviewer that every save is recorded under; attribution, not an authenticated identity.
- Scenario: one of the three sample readings the page offers; see [Scenarios](#scenarios).
- Explanation: model help on one reading: a reviewer's question about its checks, and the answer to it.
- Answer: the response to one question: its outcome, its written sentences with what the checks found, and every source it was offered.
- Outcome: what became of an answer, shown, withheld, unanswerable or unavailable, as the [outcome contract](spec.md#outcome-contract) defines them.
- Source: one numbered piece of evidence, S1, S2 and so on, that an answer may cite; see [Citations](#citations).

## Criteria

| Criterion | Violated or Review when | States |
| --- | --- | --- |
| Import issue | the reading has a missing value or a repeated acquisition ID | Review |
| Resistivity sign | apparent resistivity is at or below 0 ohm-m | Met or Violated |
| Resistivity error | reported error divided by absolute resistivity is above 0.50 | Met or Violated |
| Chargeability error | reported error divided by absolute chargeability is above 1.00 | Met or Violated |
| Transmitter contact | contact resistance is below 0 or above 10,000 ohm | Met or Violated |
| Decay trend | the least-squares slope through the ten decay samples is at or above 0 | Met or Violated |
| Frozen repeat | resistivity, chargeability and all ten decay samples equal a repeat's exactly | Met or Violated |
| Repeat agreement | the repeat difference is above 0.20 | Met or Review |
| Suspect electrode | the reading uses a suspect electrode in any slot | Met or Review |

The criteria do not assess reciprocity, receiver contact, raw waveform noise, inversion residual, coordinate consistency between readings, or instrument-specific operating ranges.

## Dataset

GeoLineAI ships one survey line, data/line-1.csv, and it is synthetic: scripts/generate-line/ produced it with a SimPEG 2D DC and IP forward model of one conductive, chargeable body in a resistive background, then added field conditions and defects by seeded rules applied across the whole line. No row was edited by hand.

| Fact | Value |
| --- | --- |
| Readings | 1,603, none rejected at import |
| Electrodes | 40, line-1-e01 to line-1-e40, 20 m apart on a straight line |
| Array | dipole-dipole, a = 1, n factor 1 to 8 |
| Repeat groups | 268, of 4 to 8 readings each |
| Stack count | 2 to 8 |
| Survey day | 2026-10-01 |
| Advisories | 1,326 Pass, 163 Needs review, 114 Fail |

The defects come from these generator rules, each a field mechanism:

| Mechanism | What it produces |
| --- | --- |
| Electrode line-1-e31 dries out after the lunch break | 19 readings with transmitter contact above 10,000 ohm, a suspect electrode, and noisy readings that receive on it |
| Receiver leads swapped for one run of readings | 23 readings with negative apparent resistivity |
| One window of wind noise | inflated reported errors and disagreeing repeats |
| Inductive coupling on long cables at n of 7 or more | some decay curves that rise |
| The instrument hangs for one repeat group | 5 readings that repeat each other exactly |
| GPS dropouts | 6 readings with one electrode's coordinates missing |
| The logger restarts and reissues IDs | 8 readings that share 4 acquisition IDs |

scripts/dataset-facts.py prints every number this file and research.md state about the line, and a backend acceptance test pins the counts the application computes.

### Scenarios

The page offers three readings, chosen from the computed results by the rules in [spec.md](spec.md#scenarios) and named in data/scenarios.json by file and row:

| Scenario | Row | Why |
| --- | --- | --- |
| Passing reading | 2 | the first reading whose advisory is Pass |
| Violated criterion | 262 | the first reading with a Violated measurement criterion: negative resistivity from the swapped leads |
| Disagreeing repeats | 1065 | the first reading whose only unresolved result is repeat agreement: inside the wind window, its resistivity is far from its repeats' median while every single-reading check is Met |

## Citations

An explanation cites numbered sources, S1, S2 and so on, built for the selected reading:

1. Result sources: each of the reading's criterion results, with its explanation.
2. Fact sources: the reading's values, electrodes, decay samples, import issues, repeats, and the failing transmitter contact of its electrodes elsewhere on the line, all read from storage.
3. Rule sources: the advisory rule and each criterion's rule, generated from the limits the code compares against.
4. Documentation sources: passages quoted verbatim from this file, research.md and the ADRs, each with its document path and section.

Every [Sn] in an answer links to that source's card on the page, which shows the source in full. A test fails when a documentation passage no longer appears verbatim in its document or its section heading is gone, so a citation never points at text the repository no longer holds.
