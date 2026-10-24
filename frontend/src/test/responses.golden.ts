// Generated from real responses by TestFrontendResponses (backend/internal/server/frontend_test.go); do not edit.
// Rerun: cd backend && go test ./internal/server -run TestFrontend -update
import type { ErrorBody, ExplanationResponse, ReadingDetail, SaveResponse, ScenariosResponse } from '../lib/api'

export const scenariosResponse = {
  "scenarios": [
    {
      "label": "Negative resistivity",
      "reading_id": "1"
    }
  ]
} satisfies ScenariosResponse

export const saveResponse = {
  "action_id": "1",
  "affected_reading_ids": [
    "1"
  ],
  "unchanged_reading_ids": []
} satisfies SaveResponse

export const detailResponse = {
  "acquisition_id": "neg",
  "advisory": {
    "results": [
      {
        "criterion": "resistivity_sign",
        "explanation": "Apparent resistivity -50 ohm-m is at or below 0.",
        "limit": 0,
        "state": "violated",
        "unit": "ohm-m",
        "value": -50
      },
      {
        "criterion": "resistivity_relative_error",
        "explanation": "Resistivity error 1.2 ohm-m is 0.024 of the value -50 ohm-m, at most the 0.5 allowed.",
        "limit": 0.5,
        "state": "met",
        "unit": "fraction",
        "value": 0.024
      },
      {
        "criterion": "chargeability_relative_error",
        "explanation": "Chargeability error 0.001 V/V is 0.0909091 of the value 0.011 V/V, at most the 1 allowed.",
        "limit": 1,
        "state": "met",
        "unit": "fraction",
        "value": 0.09090909090909091
      },
      {
        "criterion": "transmitter_contact",
        "explanation": "Transmitter contact resistance 800 ohm is within 0 to 10000 ohm.",
        "limit": 10000,
        "state": "met",
        "unit": "ohm",
        "value": 800
      },
      {
        "criterion": "decay_trend",
        "explanation": "Decay slope -0.0103999 V/V per ms is below 0: the curve falls.",
        "limit": 0,
        "state": "met",
        "unit": "V/V per ms",
        "value": -0.010399891709594775
      },
      {
        "criterion": "frozen_repeat",
        "explanation": "No compared repeat has the same resistivity, chargeability and decay samples (3 compared).",
        "limit": null,
        "state": "met",
        "unit": "",
        "value": null
      },
      {
        "criterion": "repeat_agreement",
        "explanation": "Resistivity -50 ohm-m differs from the reference 100 ohm-m, the median of 3 repeats, by 150.0%, above the 20.0% limit.",
        "limit": 0.2,
        "state": "review",
        "unit": "fraction",
        "value": 1.5
      },
      {
        "criterion": "suspect_electrode",
        "explanation": "No electrode of this reading is suspect.",
        "limit": null,
        "state": "met",
        "unit": "",
        "value": null
      }
    ],
    "value": "fail"
  },
  "apparent_resistivity": -50,
  "apparent_resistivity_err": 1.2,
  "array_type": "dipole-dipole",
  "chargeability": 0.011,
  "chargeability_err": 0.001,
  "contact_resistance": 800,
  "decay": {
    "samples": [
      10,
      9,
      8,
      7,
      6,
      5,
      4,
      3,
      2,
      1
    ]
  },
  "disagrees": true,
  "electrodes": [
    {
      "alt": 101,
      "electrode_id": "line-1-e01",
      "lat": 33.201,
      "lon": -117.101,
      "slot": "tx1"
    },
    {
      "alt": 102,
      "electrode_id": "line-1-e02",
      "lat": 33.202,
      "lon": -117.102,
      "slot": "tx2"
    },
    {
      "alt": 103,
      "electrode_id": "line-1-e03",
      "lat": 33.203,
      "lon": -117.103,
      "slot": "rx1"
    },
    {
      "alt": 104,
      "electrode_id": "line-1-e04",
      "lat": 33.204,
      "lon": -117.104,
      "slot": "rx2"
    }
  ],
  "history": [
    {
      "action_id": "1",
      "entry_count": 1,
      "new_verdict": "fail",
      "previous_verdict": "unreviewed",
      "reason": "Advisory: Fail, a criterion was violated. Apparent resistivity -50 ohm-m is at or below 0. Resistivity -50 ohm-m differs from the reference 100 ohm-m, the median of 3 repeats, by 150.0%, above the 20.0% limit. Reviewer selected Fail.",
      "recorded_at": "2026-09-30T12:00:00.000Z",
      "reviewer": "reviewer"
    }
  ],
  "import_issues": [],
  "input_current": 1.5,
  "line_id": "line-1",
  "reading_id": "1",
  "reason": "Advisory: Fail, a criterion was violated. Apparent resistivity -50 ohm-m is at or below 0. Resistivity -50 ohm-m differs from the reference 100 ohm-m, the median of 3 repeats, by 150.0%, above the 20.0% limit. Reviewer selected Fail.",
  "repeat_difference": 1.5,
  "repeat_group_id": "1",
  "repeats": [
    {
      "acquisition_id": "ok1",
      "apparent_resistivity": 100,
      "chargeability": 0.012,
      "disagrees": false,
      "left_out": false,
      "reading_id": "2",
      "repeat_difference": 0.2,
      "timestamp": "2026-08-17T14:00:12.000000000Z",
      "verdict": "unreviewed"
    },
    {
      "acquisition_id": "ok2",
      "apparent_resistivity": 100,
      "chargeability": 0.013,
      "disagrees": false,
      "left_out": false,
      "reading_id": "3",
      "repeat_difference": 0.2,
      "timestamp": "2026-08-17T14:00:12.000000000Z",
      "verdict": "unreviewed"
    },
    {
      "acquisition_id": "far",
      "apparent_resistivity": 150,
      "chargeability": 0.014,
      "disagrees": true,
      "left_out": false,
      "reading_id": "4",
      "repeat_difference": 0.5,
      "timestamp": "2026-08-17T14:00:12.000000000Z",
      "verdict": "unreviewed"
    },
    {
      "acquisition_id": "missing",
      "apparent_resistivity": null,
      "chargeability": 0.015,
      "disagrees": false,
      "left_out": true,
      "reading_id": "5",
      "repeat_difference": null,
      "timestamp": "2026-08-17T14:00:12.000000000Z",
      "verdict": "unreviewed"
    }
  ],
  "row_number": 2,
  "source_filename": "line-1.csv",
  "stacks": 4,
  "symmetry": 1,
  "timestamp": "2026-08-17T14:00:12.000000000Z",
  "verdict": "fail",
  "would_save_reasons": {
    "fail": "Advisory: Fail, a criterion was violated. Apparent resistivity -50 ohm-m is at or below 0. Resistivity -50 ohm-m differs from the reference 100 ohm-m, the median of 3 repeats, by 150.0%, above the 20.0% limit. Reviewer selected Fail.",
    "pass": "Advisory: Fail, a criterion was violated. Apparent resistivity -50 ohm-m is at or below 0. Resistivity -50 ohm-m differs from the reference 100 ohm-m, the median of 3 repeats, by 150.0%, above the 20.0% limit. Reviewer selected Pass. This differs from the advisory."
  }
} satisfies ReadingDetail

export const errorResponse = {
  "error": "reading_not_found",
  "message": "Reading 99999 does not exist.",
  "missing_reading_ids": [
    "99999"
  ]
} satisfies ErrorBody

export const shownExplanationResponse = {
  "models": {
    "judge": "fake-judge",
    "writer": "fake-writer"
  },
  "outcome": "shown",
  "reason": null,
  "sentences": [
    {
      "problems": [],
      "sources": [
        "S1"
      ],
      "supported": 0.9,
      "tells_verdict": 0.1,
      "text": "See the cited source."
    }
  ],
  "sources": [
    {
      "doc": null,
      "id": "S1",
      "kind": "result",
      "text": "Resistivity sign is Violated for neg. Apparent resistivity -50 ohm-m is at or below 0.",
      "title": "Result: Resistivity sign"
    },
    {
      "doc": null,
      "id": "S2",
      "kind": "result",
      "text": "Resistivity error is Met for neg. Resistivity error 1.2 ohm-m is 0.024 of the value -50 ohm-m, at most the 0.5 allowed.",
      "title": "Result: Resistivity error"
    },
    {
      "doc": null,
      "id": "S3",
      "kind": "result",
      "text": "Chargeability error is Met for neg. Chargeability error 0.001 V/V is 0.0909091 of the value 0.011 V/V, at most the 1 allowed.",
      "title": "Result: Chargeability error"
    },
    {
      "doc": null,
      "id": "S4",
      "kind": "result",
      "text": "Transmitter contact is Met for neg. Transmitter contact resistance 800 ohm is within 0 to 10000 ohm.",
      "title": "Result: Transmitter contact"
    },
    {
      "doc": null,
      "id": "S5",
      "kind": "result",
      "text": "Decay trend is Met for neg. Decay slope -0.0103999 V/V per ms is below 0: the curve falls.",
      "title": "Result: Decay trend"
    },
    {
      "doc": null,
      "id": "S6",
      "kind": "result",
      "text": "Frozen repeat is Met for neg. No compared repeat has the same resistivity, chargeability and decay samples (3 compared).",
      "title": "Result: Frozen repeat"
    },
    {
      "doc": null,
      "id": "S7",
      "kind": "result",
      "text": "Repeat agreement is Review for neg. Resistivity -50 ohm-m differs from the reference 100 ohm-m, the median of 3 repeats, by 150.0%, above the 20.0% limit.",
      "title": "Result: Repeat agreement"
    },
    {
      "doc": null,
      "id": "S8",
      "kind": "result",
      "text": "Suspect electrode is Met for neg. No electrode of this reading is suspect.",
      "title": "Result: Suspect electrode"
    },
    {
      "doc": null,
      "id": "S9",
      "kind": "fact",
      "text": "The advisory of neg is Fail: of its 8 recorded results, 1 are Violated, 1 are Review and 6 are Met.",
      "title": "Fact: the advisory of this reading"
    },
    {
      "doc": null,
      "id": "S10",
      "kind": "fact",
      "text": "Criteria with a recorded result for neg: Resistivity sign, Resistivity error, Chargeability error, Transmitter contact, Decay trend, Frozen repeat, Repeat agreement, Suspect electrode. Every criterion other than Import issue recorded a result for neg.",
      "title": "Fact: which criteria recorded a result"
    },
    {
      "doc": null,
      "id": "S11",
      "kind": "fact",
      "text": "Values of neg: apparent resistivity -50 ohm-m, reported resistivity error 1.2 ohm-m; chargeability 0.011 V/V, reported chargeability error 0.001 V/V; transmitter-pair contact resistance 800 ohm; input current 1.5 A; stack count 4; n factor 1.",
      "title": "Fact: values of this reading"
    },
    {
      "doc": null,
      "id": "S12",
      "kind": "fact",
      "text": "Electrodes of neg: tx1 line-1-e01, tx2 line-1-e02, rx1 line-1-e03, rx2 line-1-e04. tx1 and tx2 are the transmitting pair; rx1 and rx2 the receiving pair.",
      "title": "Fact: electrodes of this reading"
    },
    {
      "doc": null,
      "id": "S13",
      "kind": "fact",
      "text": "Decay samples of neg, gate 1 to gate 10 at 24.32 to 822.34 ms: 10, 9, 8, 7, 6, 5, 4, 3, 2, 1 V/V. No sample is above the one before it.",
      "title": "Fact: decay samples of this reading"
    },
    {
      "doc": null,
      "id": "S14",
      "kind": "fact",
      "text": "No import issues are recorded for neg.",
      "title": "Fact: import issues of this reading"
    },
    {
      "doc": null,
      "id": "S15",
      "kind": "fact",
      "text": "Repeats of neg, the other readings of its import with the same four electrode IDs: ok1 at 100 ohm-m, counts toward the reference; ok2 at 100 ohm-m, counts toward the reference; far at 150 ohm-m, counts toward the reference; missing with no resistivity, left out of the reference. 3 of its 4 repeats count toward the reference.",
      "title": "Fact: repeats of this reading"
    },
    {
      "doc": null,
      "id": "S16",
      "kind": "rule",
      "text": "The advisory is Fail when any result is Violated, Needs review when no result is Violated but some are Review, and Pass otherwise, so Pass means every recorded result is Met. It is decided from the results alone at import and never changes afterwards, and it never sets the verdict, which only the reviewer's save changes.",
      "title": "Rule: the advisory"
    },
    {
      "doc": null,
      "id": "S17",
      "kind": "rule",
      "text": "Import issue is recorded once for each import issue on the reading, a missing value or a repeated acquisition ID, and is always Review: a cell the import cannot store rejects its whole row instead, so no import issue is Violated.",
      "title": "Rule: Import issue"
    },
    {
      "doc": null,
      "id": "S18",
      "kind": "rule",
      "text": "Resistivity sign is Met when the apparent resistivity is above 0 ohm-m and Violated when it is at or below 0 ohm-m. It records no result when the apparent resistivity is unavailable.",
      "title": "Rule: Resistivity sign"
    },
    {
      "doc": null,
      "id": "S19",
      "kind": "rule",
      "text": "Resistivity error divides the reported apparent resistivity error by the absolute apparent resistivity. It is Met when that fraction is at most 0.5 and Violated when it is above 0.5 or too large to represent. It records no result when either value is unavailable or the apparent resistivity is 0.",
      "title": "Rule: Resistivity error"
    },
    {
      "doc": null,
      "id": "S20",
      "kind": "rule",
      "text": "Chargeability error divides the reported chargeability error by the absolute chargeability. It is Met when that fraction is at most 1 and Violated when it is above 1 or too large to represent. It records no result when either value is unavailable or the chargeability is 0.",
      "title": "Rule: Chargeability error"
    },
    {
      "doc": null,
      "id": "S21",
      "kind": "rule",
      "text": "Transmitter contact is Met when the contact resistance is from 0 to 10000 ohm and Violated when it is below 0 ohm or above 10000 ohm. The contact resistance belongs to the transmitting pair of the reading, its tx1 and tx2 electrodes, and says nothing about its receiving pair. It records no result when the contact resistance is unavailable.",
      "title": "Rule: Transmitter contact"
    },
    {
      "doc": null,
      "id": "S22",
      "kind": "rule",
      "text": "Decay trend fits an unweighted least-squares line, with a free intercept, through the ten decay samples against the ten gate centres, from 24.32 to 822.34 ms, both on linear scales. It is Met when the slope is below 0 V/V per ms, so the curve falls, and Violated when the slope is at or above 0 V/V per ms or is not finite. It records no result when the decay curve is unavailable.",
      "title": "Rule: Decay trend"
    },
    {
      "doc": null,
      "id": "S23",
      "kind": "rule",
      "text": "Frozen repeat compares the reading with each of its repeats that has an apparent resistivity, a chargeability and all ten decay samples. It is Violated when all of those values exactly equal a repeat's, since independent measurements never agree to every digit, and Met when they differ from every compared repeat. It records no result when the reading lacks one of those values or no repeat has them all.",
      "title": "Rule: Frozen repeat"
    },
    {
      "doc": null,
      "id": "S24",
      "kind": "rule",
      "text": "Repeat agreement compares the apparent resistivity with the reference, the median resistivity of the reading's repeats, leaving out repeats without a resistivity and repeats with a Violated single-reading result. The repeat difference is the absolute difference between the resistivity and the reference, divided by the reference. It is Met when the repeat difference is at most 20.0% and Review when it is above 20.0%. It records no result when the reading has no resistivity or fewer than two repeats count toward the reference.",
      "title": "Rule: Repeat agreement"
    },
    {
      "doc": null,
      "id": "S25",
      "kind": "rule",
      "text": "An electrode is suspect when the readings in which it transmits, in the tx1 or tx2 slot, violate Transmitter contact with at least 2 different partner electrodes. Suspect electrode is Review when the reading uses a suspect electrode in any slot and Met when it uses none; it marks an association to review, not a proven fault. It records no result when the reading has no electrode ID.",
      "title": "Rule: Suspect electrode"
    },
    {
      "doc": "docs/research.md#1-domain-rules",
      "id": "S26",
      "kind": "doc",
      "text": "D1: apparent resistivity at or below zero is physically implausible. 23 readings of the line are negative, all from one run of swapped receiver leads. 9 of them have two or more positive repeats to compare against, and all 9 disagree with them.",
      "title": "Evidence for the resistivity sign limit (Apparent resistivity above 0 ohm-m)"
    },
    {
      "doc": "docs/research.md#1-domain-rules",
      "id": "S27",
      "kind": "doc",
      "text": "0 is the physical boundary D1 names. Counting exactly 0 as a violation is our extension (section 4).",
      "title": "Why the resistivity sign limit has its value"
    },
    {
      "doc": "docs/research.md#1-domain-rules",
      "id": "S28",
      "kind": "doc",
      "text": "D3: a decay falls overall, and a fitted line with a negative slope is a sound first test. 46 curves of the line rise; 25 of them belong to readings whose chargeability is negative, where noise pushed the curve below zero so that it climbs back toward it. The slopes of falling and rising curves meet near 0, with no empty gap between.",
      "title": "Evidence for the decay trend limit (Decay slope below 0)"
    },
    {
      "doc": "docs/research.md#1-domain-rules",
      "id": "S29",
      "kind": "doc",
      "text": "0 separates falling from rising by definition. The method is an unweighted least-squares line with a free intercept over the ten gate centres (section 7 shows it verified).",
      "title": "Why the decay trend limit has its value"
    },
    {
      "doc": "docs/research.md#2-from-a-gap-in-the-data",
      "id": "S30",
      "kind": "doc",
      "text": "Every contact resistance is either at most 1,091.6 ohm or at least 20,077.1 ohm: 19 readings lie above 10,000 ohm, all transmitting through line-1-e31 after it dried out; none is negative. A negative contact resistance is impossible, so it violates too.",
      "title": "Evidence for the transmitter contact limit (Transmitter contact resistance 0 to 10,000 ohm, inclusive)"
    },
    {
      "doc": "docs/research.md#2-from-a-gap-in-the-data",
      "id": "S31",
      "kind": "doc",
      "text": "Any limit from 1,091.6 to 20,077.1 flags the same 19 readings; 10,000 sits inside the gap.",
      "title": "Why the transmitter contact limit has its value"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S32",
      "kind": "doc",
      "text": "No gap: 1,460 readings below 0.05, 57 from 0.05 to 0.15, 52 from 0.15 to 0.50, 19 from 0.50 to 1.0 and 15 above 1.0. The nearest values straddle the limit: 0.4714 below, 0.5247 above. The large errors come from the wind window and from receiving on the dried electrode.",
      "title": "Evidence for the resistivity error limit (Resistivity relative error at most 0.50)"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S33",
      "kind": "doc",
      "text": "At 0.50 the reported uncertainty is half the value, which leaves the reading too uncertain to trust its magnitude. That is our reasoning, not a finding.",
      "title": "Why the resistivity error limit has its value"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S34",
      "kind": "doc",
      "text": "No gap: 1,364 readings below 0.2, 136 from 0.2 to 0.4, then a thinning tail (24, 14 and 13 across 0.4 to 1.0), 52 above 1.0. The nearest values straddle the limit: 0.9937 below, 1.0020 above.",
      "title": "Evidence for the chargeability error limit (Chargeability relative error at most 1.00)"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S35",
      "kind": "doc",
      "text": "At 1.00 the reported error equals the value itself, so the measurement carries no information beyond its own uncertainty. That is our reasoning, not a finding: the data offers no better place.",
      "title": "Why the chargeability error limit has its value"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S36",
      "kind": "doc",
      "text": "No clean gap. Of 1,531 readings compared: 1,503 within 0.15, 4 from 0.15 to 0.20, 13 from 0.20 to 0.40 and 11 beyond 0.40. Of the 24 above 0.20, 19 use the suspect electrode or have a Violated single-reading result; the other 5 have neither. Four of those lie inside the wind window; the fifth, at row 1536, is ordinary measurement noise at n factor 8 over the conductive body, 26% from its repeats' median: a sound reading the check sends to review. D4 names outlier statistics as the first test but gives no number.",
      "title": "Evidence for the repeat agreement limit (Repeat agreement within 0.20 of the reference)"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S37",
      "kind": "doc",
      "text": "Ordinary disagreement stays below 0.15. 0.20 leaves room above that for a sound reading's noise, and still flags a reading a fifth away from what its repeats agree on.",
      "title": "Why the repeat agreement limit has its value"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S38",
      "kind": "doc",
      "text": "A zero apparent resistivity is no more physical than a negative one. No reading of the line is 0 (the smallest positive is 14.712), so the data never tests it.",
      "title": "Design choice: Resistivity of exactly 0 violates the sign criterion"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S39",
      "kind": "doc",
      "text": "Dividing by 0 has no answer, and a negative value would make the ratio negative. A zero resistivity is judged by its sign criterion alone; a zero chargeability by no criterion, since no domain rule says one is wrong.",
      "title": "Design choice: A relative error is not computed when the value is 0, and divides by the absolute value"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S40",
      "kind": "doc",
      "text": "One repeat is not a reference; a median at or below 0 cannot be divided by; a repeat already known to be bad would drag the reference toward itself.",
      "title": "Design choice: The repeat-agreement reference needs at least 2 eligible repeats, a median above 0, and excludes repeats with a Violated single-reading result"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S41",
      "kind": "doc",
      "text": "Independent measurements never agree to every digit. The line's 5 frozen readings are one whole repeat group, taken within 40 seconds; nothing looser was needed to find them, and a partial comparison would call readings frozen on missing data.",
      "title": "Design choice: Frozen repeat means resistivity, chargeability and all ten decay samples exactly equal, compared only between repeats with every value available"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S42",
      "kind": "doc",
      "text": "Three electrodes have failing contact. line-1-e31 fails with two partners, its neighbours on both sides; line-1-e30 and line-1-e32 fail with one partner each, always e31. A count of 2 separates the faulty electrode from its neighbours exactly. D6 supports looking at shared electrodes; it gives no count. Contact is measured on the transmitting pair, so only tx slots count.",
      "title": "Design choice: A suspect electrode needs failing transmitter contact with at least 2 distinct partners, counting the tx1 and tx2 slots"
    },
    {
      "doc": "docs/research.md#6-chosen-not-to-use",
      "id": "S43",
      "kind": "doc",
      "text": "Rises are common: 675 of the 1,557 falling curves have at least one. Ratio limits 1.10, 1.25, 1.50 and 2.00 flag 28, 21, 0 and 0 readings, of which 7, 0, 0 and 0 are not already rising curves. No labelled truth says the extra ones are defects.",
      "title": "Chosen not to use: A local decay rise: one sample more than a ratio above the one before"
    },
    {
      "doc": "docs/research.md#6-chosen-not-to-use",
      "id": "S44",
      "kind": "doc",
      "text": "A rising or sign-reversed exponential fits as well as a falling one, so a low error cannot establish a good decay.",
      "title": "Chosen not to use: A log-space decay fit error"
    },
    {
      "doc": "docs/research.md#7-what-the-research-supports",
      "id": "S45",
      "kind": "doc",
      "text": "No public source validates 0.50, 1.00, 10,000 ohm or 0.20 for this survey.",
      "title": "What the research supports: Calling those limits data-based or provisional, never standards."
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S46",
      "kind": "doc",
      "text": "Apparent resistivity at or below zero is physically implausible: the ground cannot produce it.",
      "title": "Domain rule D1"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S47",
      "kind": "doc",
      "text": "A reading with an implausible or missing value is kept and flagged rather than dropped, so the reviewer can see that it happened.",
      "title": "Domain rule D2"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S48",
      "kind": "doc",
      "text": "An induced-polarization decay falls overall after the current is switched off, though not necessarily at every gate; a straight line fitted with a negative slope is a sound first test.",
      "title": "Domain rule D3"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S49",
      "kind": "doc",
      "text": "Comparing a reading with its repeats, an outlier statistic, is a sound first test for a reading that disagrees.",
      "title": "Domain rule D4"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S50",
      "kind": "doc",
      "text": "No published formula or threshold fixes these limits for this survey; each one is a stated choice.",
      "title": "Domain rule D5"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S51",
      "kind": "doc",
      "text": "Readings that share an electrode with known-bad readings deserve a second look.",
      "title": "Domain rule D6"
    },
    {
      "doc": "docs/domain.md#dataset",
      "id": "S52",
      "kind": "doc",
      "text": "GeoLineAI ships one survey line, data/line-1.csv, and it is synthetic: scripts/generate-line/ produced it with a SimPEG 2D DC and IP forward model of one conductive, chargeable body in a resistive background, then added field conditions and defects by seeded rules applied across the whole line. No row was edited by hand.",
      "title": "The dataset is synthetic"
    },
    {
      "doc": "docs/domain.md#criteria",
      "id": "S53",
      "kind": "doc",
      "text": "The criteria do not assess reciprocity, receiver contact, raw waveform noise, inversion residual, coordinate consistency between readings, or instrument-specific operating ranges.",
      "title": "What the criteria do not assess"
    },
    {
      "doc": "docs/domain.md#terms",
      "id": "S54",
      "kind": "doc",
      "text": "A suspect electrode is one whose transmitting readings violate transmitter contact with two or more different partner electrodes; it marks an association to review, not a proven fault.",
      "title": "Term: suspect electrode"
    },
    {
      "doc": "docs/domain.md#terms",
      "id": "S55",
      "kind": "doc",
      "text": "A reading's repeat group is the readings from its import that share all four electrode IDs; its repeats are the group's other readings.",
      "title": "Term: repeat group"
    },
    {
      "doc": "docs/adr/0003-deterministic-advisory.md#consequences",
      "id": "S56",
      "kind": "doc",
      "text": "The advisory is Pass, Fail or Needs review; there is no score, band or confidence.",
      "title": "What the advisory is"
    },
    {
      "doc": "docs/adr/0003-deterministic-advisory.md#consequences",
      "id": "S57",
      "kind": "doc",
      "text": "On the shipped line, repeat agreement is the only criterion that marks 5 readings for review: they disagree with their repeats while using no suspect electrode and having no Violated result. It is the outlier statistic D4 names; of the readings it marks, most also use the suspect electrode or have a Violated result.",
      "title": "What repeat agreement adds"
    },
    {
      "doc": "docs/adr/0003-deterministic-advisory.md#consequences",
      "id": "S58",
      "kind": "doc",
      "text": "A suspect electrode needs failing contact readings with two different partners, so a faulty electrode at the end of a line, which transmits with one partner only, is not detected; its readings still violate transmitter contact themselves.",
      "title": "What the suspect electrode criterion cannot detect"
    }
  ]
} satisfies ExplanationResponse

export const unavailableExplanationResponse = {
  "models": {
    "judge": "",
    "writer": ""
  },
  "outcome": "unavailable",
  "reason": "invalid_response",
  "sentences": [],
  "sources": [
    {
      "doc": null,
      "id": "S1",
      "kind": "result",
      "text": "Resistivity sign is Violated for neg. Apparent resistivity -50 ohm-m is at or below 0.",
      "title": "Result: Resistivity sign"
    },
    {
      "doc": null,
      "id": "S2",
      "kind": "result",
      "text": "Resistivity error is Met for neg. Resistivity error 1.2 ohm-m is 0.024 of the value -50 ohm-m, at most the 0.5 allowed.",
      "title": "Result: Resistivity error"
    },
    {
      "doc": null,
      "id": "S3",
      "kind": "result",
      "text": "Chargeability error is Met for neg. Chargeability error 0.001 V/V is 0.0909091 of the value 0.011 V/V, at most the 1 allowed.",
      "title": "Result: Chargeability error"
    },
    {
      "doc": null,
      "id": "S4",
      "kind": "result",
      "text": "Transmitter contact is Met for neg. Transmitter contact resistance 800 ohm is within 0 to 10000 ohm.",
      "title": "Result: Transmitter contact"
    },
    {
      "doc": null,
      "id": "S5",
      "kind": "result",
      "text": "Decay trend is Met for neg. Decay slope -0.0103999 V/V per ms is below 0: the curve falls.",
      "title": "Result: Decay trend"
    },
    {
      "doc": null,
      "id": "S6",
      "kind": "result",
      "text": "Frozen repeat is Met for neg. No compared repeat has the same resistivity, chargeability and decay samples (3 compared).",
      "title": "Result: Frozen repeat"
    },
    {
      "doc": null,
      "id": "S7",
      "kind": "result",
      "text": "Repeat agreement is Review for neg. Resistivity -50 ohm-m differs from the reference 100 ohm-m, the median of 3 repeats, by 150.0%, above the 20.0% limit.",
      "title": "Result: Repeat agreement"
    },
    {
      "doc": null,
      "id": "S8",
      "kind": "result",
      "text": "Suspect electrode is Met for neg. No electrode of this reading is suspect.",
      "title": "Result: Suspect electrode"
    },
    {
      "doc": null,
      "id": "S9",
      "kind": "fact",
      "text": "The advisory of neg is Fail: of its 8 recorded results, 1 are Violated, 1 are Review and 6 are Met.",
      "title": "Fact: the advisory of this reading"
    },
    {
      "doc": null,
      "id": "S10",
      "kind": "fact",
      "text": "Criteria with a recorded result for neg: Resistivity sign, Resistivity error, Chargeability error, Transmitter contact, Decay trend, Frozen repeat, Repeat agreement, Suspect electrode. Every criterion other than Import issue recorded a result for neg.",
      "title": "Fact: which criteria recorded a result"
    },
    {
      "doc": null,
      "id": "S11",
      "kind": "fact",
      "text": "Values of neg: apparent resistivity -50 ohm-m, reported resistivity error 1.2 ohm-m; chargeability 0.011 V/V, reported chargeability error 0.001 V/V; transmitter-pair contact resistance 800 ohm; input current 1.5 A; stack count 4; n factor 1.",
      "title": "Fact: values of this reading"
    },
    {
      "doc": null,
      "id": "S12",
      "kind": "fact",
      "text": "Electrodes of neg: tx1 line-1-e01, tx2 line-1-e02, rx1 line-1-e03, rx2 line-1-e04. tx1 and tx2 are the transmitting pair; rx1 and rx2 the receiving pair.",
      "title": "Fact: electrodes of this reading"
    },
    {
      "doc": null,
      "id": "S13",
      "kind": "fact",
      "text": "Decay samples of neg, gate 1 to gate 10 at 24.32 to 822.34 ms: 10, 9, 8, 7, 6, 5, 4, 3, 2, 1 V/V. No sample is above the one before it.",
      "title": "Fact: decay samples of this reading"
    },
    {
      "doc": null,
      "id": "S14",
      "kind": "fact",
      "text": "No import issues are recorded for neg.",
      "title": "Fact: import issues of this reading"
    },
    {
      "doc": null,
      "id": "S15",
      "kind": "fact",
      "text": "Repeats of neg, the other readings of its import with the same four electrode IDs: ok1 at 100 ohm-m, counts toward the reference; ok2 at 100 ohm-m, counts toward the reference; far at 150 ohm-m, counts toward the reference; missing with no resistivity, left out of the reference. 3 of its 4 repeats count toward the reference.",
      "title": "Fact: repeats of this reading"
    },
    {
      "doc": null,
      "id": "S16",
      "kind": "rule",
      "text": "The advisory is Fail when any result is Violated, Needs review when no result is Violated but some are Review, and Pass otherwise, so Pass means every recorded result is Met. It is decided from the results alone at import and never changes afterwards, and it never sets the verdict, which only the reviewer's save changes.",
      "title": "Rule: the advisory"
    },
    {
      "doc": null,
      "id": "S17",
      "kind": "rule",
      "text": "Import issue is recorded once for each import issue on the reading, a missing value or a repeated acquisition ID, and is always Review: a cell the import cannot store rejects its whole row instead, so no import issue is Violated.",
      "title": "Rule: Import issue"
    },
    {
      "doc": null,
      "id": "S18",
      "kind": "rule",
      "text": "Resistivity sign is Met when the apparent resistivity is above 0 ohm-m and Violated when it is at or below 0 ohm-m. It records no result when the apparent resistivity is unavailable.",
      "title": "Rule: Resistivity sign"
    },
    {
      "doc": null,
      "id": "S19",
      "kind": "rule",
      "text": "Resistivity error divides the reported apparent resistivity error by the absolute apparent resistivity. It is Met when that fraction is at most 0.5 and Violated when it is above 0.5 or too large to represent. It records no result when either value is unavailable or the apparent resistivity is 0.",
      "title": "Rule: Resistivity error"
    },
    {
      "doc": null,
      "id": "S20",
      "kind": "rule",
      "text": "Chargeability error divides the reported chargeability error by the absolute chargeability. It is Met when that fraction is at most 1 and Violated when it is above 1 or too large to represent. It records no result when either value is unavailable or the chargeability is 0.",
      "title": "Rule: Chargeability error"
    },
    {
      "doc": null,
      "id": "S21",
      "kind": "rule",
      "text": "Transmitter contact is Met when the contact resistance is from 0 to 10000 ohm and Violated when it is below 0 ohm or above 10000 ohm. The contact resistance belongs to the transmitting pair of the reading, its tx1 and tx2 electrodes, and says nothing about its receiving pair. It records no result when the contact resistance is unavailable.",
      "title": "Rule: Transmitter contact"
    },
    {
      "doc": null,
      "id": "S22",
      "kind": "rule",
      "text": "Decay trend fits an unweighted least-squares line, with a free intercept, through the ten decay samples against the ten gate centres, from 24.32 to 822.34 ms, both on linear scales. It is Met when the slope is below 0 V/V per ms, so the curve falls, and Violated when the slope is at or above 0 V/V per ms or is not finite. It records no result when the decay curve is unavailable.",
      "title": "Rule: Decay trend"
    },
    {
      "doc": null,
      "id": "S23",
      "kind": "rule",
      "text": "Frozen repeat compares the reading with each of its repeats that has an apparent resistivity, a chargeability and all ten decay samples. It is Violated when all of those values exactly equal a repeat's, since independent measurements never agree to every digit, and Met when they differ from every compared repeat. It records no result when the reading lacks one of those values or no repeat has them all.",
      "title": "Rule: Frozen repeat"
    },
    {
      "doc": null,
      "id": "S24",
      "kind": "rule",
      "text": "Repeat agreement compares the apparent resistivity with the reference, the median resistivity of the reading's repeats, leaving out repeats without a resistivity and repeats with a Violated single-reading result. The repeat difference is the absolute difference between the resistivity and the reference, divided by the reference. It is Met when the repeat difference is at most 20.0% and Review when it is above 20.0%. It records no result when the reading has no resistivity or fewer than two repeats count toward the reference.",
      "title": "Rule: Repeat agreement"
    },
    {
      "doc": null,
      "id": "S25",
      "kind": "rule",
      "text": "An electrode is suspect when the readings in which it transmits, in the tx1 or tx2 slot, violate Transmitter contact with at least 2 different partner electrodes. Suspect electrode is Review when the reading uses a suspect electrode in any slot and Met when it uses none; it marks an association to review, not a proven fault. It records no result when the reading has no electrode ID.",
      "title": "Rule: Suspect electrode"
    },
    {
      "doc": "docs/research.md#1-domain-rules",
      "id": "S26",
      "kind": "doc",
      "text": "D1: apparent resistivity at or below zero is physically implausible. 23 readings of the line are negative, all from one run of swapped receiver leads. 9 of them have two or more positive repeats to compare against, and all 9 disagree with them.",
      "title": "Evidence for the resistivity sign limit (Apparent resistivity above 0 ohm-m)"
    },
    {
      "doc": "docs/research.md#1-domain-rules",
      "id": "S27",
      "kind": "doc",
      "text": "0 is the physical boundary D1 names. Counting exactly 0 as a violation is our extension (section 4).",
      "title": "Why the resistivity sign limit has its value"
    },
    {
      "doc": "docs/research.md#1-domain-rules",
      "id": "S28",
      "kind": "doc",
      "text": "D3: a decay falls overall, and a fitted line with a negative slope is a sound first test. 46 curves of the line rise; 25 of them belong to readings whose chargeability is negative, where noise pushed the curve below zero so that it climbs back toward it. The slopes of falling and rising curves meet near 0, with no empty gap between.",
      "title": "Evidence for the decay trend limit (Decay slope below 0)"
    },
    {
      "doc": "docs/research.md#1-domain-rules",
      "id": "S29",
      "kind": "doc",
      "text": "0 separates falling from rising by definition. The method is an unweighted least-squares line with a free intercept over the ten gate centres (section 7 shows it verified).",
      "title": "Why the decay trend limit has its value"
    },
    {
      "doc": "docs/research.md#2-from-a-gap-in-the-data",
      "id": "S30",
      "kind": "doc",
      "text": "Every contact resistance is either at most 1,091.6 ohm or at least 20,077.1 ohm: 19 readings lie above 10,000 ohm, all transmitting through line-1-e31 after it dried out; none is negative. A negative contact resistance is impossible, so it violates too.",
      "title": "Evidence for the transmitter contact limit (Transmitter contact resistance 0 to 10,000 ohm, inclusive)"
    },
    {
      "doc": "docs/research.md#2-from-a-gap-in-the-data",
      "id": "S31",
      "kind": "doc",
      "text": "Any limit from 1,091.6 to 20,077.1 flags the same 19 readings; 10,000 sits inside the gap.",
      "title": "Why the transmitter contact limit has its value"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S32",
      "kind": "doc",
      "text": "No gap: 1,460 readings below 0.05, 57 from 0.05 to 0.15, 52 from 0.15 to 0.50, 19 from 0.50 to 1.0 and 15 above 1.0. The nearest values straddle the limit: 0.4714 below, 0.5247 above. The large errors come from the wind window and from receiving on the dried electrode.",
      "title": "Evidence for the resistivity error limit (Resistivity relative error at most 0.50)"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S33",
      "kind": "doc",
      "text": "At 0.50 the reported uncertainty is half the value, which leaves the reading too uncertain to trust its magnitude. That is our reasoning, not a finding.",
      "title": "Why the resistivity error limit has its value"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S34",
      "kind": "doc",
      "text": "No gap: 1,364 readings below 0.2, 136 from 0.2 to 0.4, then a thinning tail (24, 14 and 13 across 0.4 to 1.0), 52 above 1.0. The nearest values straddle the limit: 0.9937 below, 1.0020 above.",
      "title": "Evidence for the chargeability error limit (Chargeability relative error at most 1.00)"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S35",
      "kind": "doc",
      "text": "At 1.00 the reported error equals the value itself, so the measurement carries no information beyond its own uncertainty. That is our reasoning, not a finding: the data offers no better place.",
      "title": "Why the chargeability error limit has its value"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S36",
      "kind": "doc",
      "text": "No clean gap. Of 1,531 readings compared: 1,503 within 0.15, 4 from 0.15 to 0.20, 13 from 0.20 to 0.40 and 11 beyond 0.40. Of the 24 above 0.20, 19 use the suspect electrode or have a Violated single-reading result; the other 5 have neither. Four of those lie inside the wind window; the fifth, at row 1536, is ordinary measurement noise at n factor 8 over the conductive body, 26% from its repeats' median: a sound reading the check sends to review. D4 names outlier statistics as the first test but gives no number.",
      "title": "Evidence for the repeat agreement limit (Repeat agreement within 0.20 of the reference)"
    },
    {
      "doc": "docs/research.md#3-provisional",
      "id": "S37",
      "kind": "doc",
      "text": "Ordinary disagreement stays below 0.15. 0.20 leaves room above that for a sound reading's noise, and still flags a reading a fifth away from what its repeats agree on.",
      "title": "Why the repeat agreement limit has its value"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S38",
      "kind": "doc",
      "text": "A zero apparent resistivity is no more physical than a negative one. No reading of the line is 0 (the smallest positive is 14.712), so the data never tests it.",
      "title": "Design choice: Resistivity of exactly 0 violates the sign criterion"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S39",
      "kind": "doc",
      "text": "Dividing by 0 has no answer, and a negative value would make the ratio negative. A zero resistivity is judged by its sign criterion alone; a zero chargeability by no criterion, since no domain rule says one is wrong.",
      "title": "Design choice: A relative error is not computed when the value is 0, and divides by the absolute value"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S40",
      "kind": "doc",
      "text": "One repeat is not a reference; a median at or below 0 cannot be divided by; a repeat already known to be bad would drag the reference toward itself.",
      "title": "Design choice: The repeat-agreement reference needs at least 2 eligible repeats, a median above 0, and excludes repeats with a Violated single-reading result"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S41",
      "kind": "doc",
      "text": "Independent measurements never agree to every digit. The line's 5 frozen readings are one whole repeat group, taken within 40 seconds; nothing looser was needed to find them, and a partial comparison would call readings frozen on missing data.",
      "title": "Design choice: Frozen repeat means resistivity, chargeability and all ten decay samples exactly equal, compared only between repeats with every value available"
    },
    {
      "doc": "docs/research.md#4-design-choices",
      "id": "S42",
      "kind": "doc",
      "text": "Three electrodes have failing contact. line-1-e31 fails with two partners, its neighbours on both sides; line-1-e30 and line-1-e32 fail with one partner each, always e31. A count of 2 separates the faulty electrode from its neighbours exactly. D6 supports looking at shared electrodes; it gives no count. Contact is measured on the transmitting pair, so only tx slots count.",
      "title": "Design choice: A suspect electrode needs failing transmitter contact with at least 2 distinct partners, counting the tx1 and tx2 slots"
    },
    {
      "doc": "docs/research.md#6-chosen-not-to-use",
      "id": "S43",
      "kind": "doc",
      "text": "Rises are common: 675 of the 1,557 falling curves have at least one. Ratio limits 1.10, 1.25, 1.50 and 2.00 flag 28, 21, 0 and 0 readings, of which 7, 0, 0 and 0 are not already rising curves. No labelled truth says the extra ones are defects.",
      "title": "Chosen not to use: A local decay rise: one sample more than a ratio above the one before"
    },
    {
      "doc": "docs/research.md#6-chosen-not-to-use",
      "id": "S44",
      "kind": "doc",
      "text": "A rising or sign-reversed exponential fits as well as a falling one, so a low error cannot establish a good decay.",
      "title": "Chosen not to use: A log-space decay fit error"
    },
    {
      "doc": "docs/research.md#7-what-the-research-supports",
      "id": "S45",
      "kind": "doc",
      "text": "No public source validates 0.50, 1.00, 10,000 ohm or 0.20 for this survey.",
      "title": "What the research supports: Calling those limits data-based or provisional, never standards."
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S46",
      "kind": "doc",
      "text": "Apparent resistivity at or below zero is physically implausible: the ground cannot produce it.",
      "title": "Domain rule D1"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S47",
      "kind": "doc",
      "text": "A reading with an implausible or missing value is kept and flagged rather than dropped, so the reviewer can see that it happened.",
      "title": "Domain rule D2"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S48",
      "kind": "doc",
      "text": "An induced-polarization decay falls overall after the current is switched off, though not necessarily at every gate; a straight line fitted with a negative slope is a sound first test.",
      "title": "Domain rule D3"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S49",
      "kind": "doc",
      "text": "Comparing a reading with its repeats, an outlier statistic, is a sound first test for a reading that disagrees.",
      "title": "Domain rule D4"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S50",
      "kind": "doc",
      "text": "No published formula or threshold fixes these limits for this survey; each one is a stated choice.",
      "title": "Domain rule D5"
    },
    {
      "doc": "docs/domain.md#domain-rules",
      "id": "S51",
      "kind": "doc",
      "text": "Readings that share an electrode with known-bad readings deserve a second look.",
      "title": "Domain rule D6"
    },
    {
      "doc": "docs/domain.md#dataset",
      "id": "S52",
      "kind": "doc",
      "text": "GeoLineAI ships one survey line, data/line-1.csv, and it is synthetic: scripts/generate-line/ produced it with a SimPEG 2D DC and IP forward model of one conductive, chargeable body in a resistive background, then added field conditions and defects by seeded rules applied across the whole line. No row was edited by hand.",
      "title": "The dataset is synthetic"
    },
    {
      "doc": "docs/domain.md#criteria",
      "id": "S53",
      "kind": "doc",
      "text": "The criteria do not assess reciprocity, receiver contact, raw waveform noise, inversion residual, coordinate consistency between readings, or instrument-specific operating ranges.",
      "title": "What the criteria do not assess"
    },
    {
      "doc": "docs/domain.md#terms",
      "id": "S54",
      "kind": "doc",
      "text": "A suspect electrode is one whose transmitting readings violate transmitter contact with two or more different partner electrodes; it marks an association to review, not a proven fault.",
      "title": "Term: suspect electrode"
    },
    {
      "doc": "docs/domain.md#terms",
      "id": "S55",
      "kind": "doc",
      "text": "A reading's repeat group is the readings from its import that share all four electrode IDs; its repeats are the group's other readings.",
      "title": "Term: repeat group"
    },
    {
      "doc": "docs/adr/0003-deterministic-advisory.md#consequences",
      "id": "S56",
      "kind": "doc",
      "text": "The advisory is Pass, Fail or Needs review; there is no score, band or confidence.",
      "title": "What the advisory is"
    },
    {
      "doc": "docs/adr/0003-deterministic-advisory.md#consequences",
      "id": "S57",
      "kind": "doc",
      "text": "On the shipped line, repeat agreement is the only criterion that marks 5 readings for review: they disagree with their repeats while using no suspect electrode and having no Violated result. It is the outlier statistic D4 names; of the readings it marks, most also use the suspect electrode or have a Violated result.",
      "title": "What repeat agreement adds"
    },
    {
      "doc": "docs/adr/0003-deterministic-advisory.md#consequences",
      "id": "S58",
      "kind": "doc",
      "text": "A suspect electrode needs failing contact readings with two different partners, so a faulty electrode at the end of a line, which transmits with one partner only, is not detected; its readings still violate transmitter contact themselves.",
      "title": "What the suspect electrode criterion cannot detect"
    }
  ]
} satisfies ExplanationResponse
