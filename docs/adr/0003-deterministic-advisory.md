# Decide every advisory with deterministic criteria stored at import; no model inference

The import evaluates every retained reading against fixed criteria and stores its advisory with it: Fail when any result is Violated, Needs review when none is Violated but some are Review, Pass otherwise. Criteria look at one reading's validated values and import issues and, once the whole file is parsed, at its repeat group and electrodes, because a line's repeats and electrodes all arrive in its one file; they never look at a verdict. Every limit rests on a domain rule or on the shipped data ([research.md](../research.md) holds the evidence for each), and the advisory is exactly what the results say: a reading that needs a human is labelled Needs review, not resolved by inference.

## Considered options

- **Model inference for Needs review, or for the whole advisory:** what remains once the criteria are applied, such as missing coordinates, repeated IDs and suspect-electrode readings, is where the honest answer is that a human decides. Inference would put an uncalibrated judgment on exactly those readings.
- **Negative resistivity as Review rather than Fail:** the advisory never sets the verdict, so a computed Fail is already a flag the reviewer can overrule.
- **Suspect electrodes by the share of their readings that fail:** a faulty electrode's neighbours fail in many of their transmitting readings, only ever paired with it, so a share threshold either misses the faulty electrode or flags its neighbours. Counting distinct failing partners separates them.
- **Criteria for chargeability or input current at or below 0:** no domain rule says such a value is wrong, and negative IP effects occur in real surveys.
- **Repeat disagreement as Fail:** when most of a repeat group comes from a faulty period, the group median cannot say which side is right.

## Consequences

- The advisory is Pass, Fail or Needs review; there is no score, band or confidence.
- The import needs no provider key and stores every retained row with its advisory, in its file's transaction.
- An acceptance test pins the advisory and criterion counts on the shipped line, so a rule that reads the data differently fails a test rather than shifting silently.
- On the shipped line, repeat agreement is the only criterion that marks 5 readings for review: they disagree with their repeats while using no suspect electrode and having no Violated result. It is the outlier statistic D4 names; of the readings it marks, most also use the suspect electrode or have a Violated result.
- A suspect electrode needs failing contact readings with two different partners, so a faulty electrode at the end of a line, which transmits with one partner only, is not detected; its readings still violate transmitter contact themselves.
- The criteria, limits and explanations are constants in code; changing one means recreating the database and reimporting (ADR 0001).
