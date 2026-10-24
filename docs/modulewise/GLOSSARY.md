# ModuleWise showcase

The vocabulary of the `modulewise` branch: GeoLineAI used as a fixed application against which ModuleWise's composable runtime is tried, one boundary at a time. Geophysical review terms live in [docs/domain.md](../domain.md) and are not repeated here.

## Comparison

**Baseline**:
GeoLineAI as tagged `baseline`, the commit every chapter is compared against.
_Avoid_: before version, original, main

**Conversion**:
A change that moves one existing boundary of GeoLineAI behind a ModuleWise component while the user task, the data and every correctness rule stay the same as the baseline.
_Avoid_: port, migration, rewrite

**Extension**:
A change that adds product behavior the baseline's specification rules out, built on top of the conversions.
_Avoid_: feature, enhancement

**Chapter**:
One conversion or extension, recorded with what changed, why, what stayed equivalent, what it cost, and when plain Go would be the better choice.
_Avoid_: step, milestone, phase

**Parity**:
The state in which a converted path and the baseline path give the same output for every input the chapter's proof covers.
_Avoid_: equivalence, compatibility

**Oracle**:
The baseline's Go implementation of a converted function, kept only to supply the expected output for generated inputs.
_Avoid_: reference implementation, golden

## Capabilities and evidence

**Capability**:
One thing a composable runtime can do for an application, named C01 to C27 in [evidence.md](evidence.md).
_Avoid_: feature, module

**Claim**:
A statement that ModuleWise provides, lacks or will provide a capability or behavior.
_Avoid_: fact, finding

**Evidence row**:
One claim with its status, the commit it applies to and its proof, identified as E01, E02 and so on.
_Avoid_: note, reference

**Verified**:
The status of a claim this project proved at a pinned commit, by running the behavior or, for a claim about what the code or a registry contains, by a command over it, with the command and its output recorded.
_Avoid_: confirmed, known

**Stated**:
The status of a claim a public ModuleWise source makes and this project has not yet run.
_Avoid_: documented, claimed

**Inferred**:
The status of a claim this project reasoned to, with the reasoning written out, and has not proved.
_Avoid_: assumed, likely

**Unknown**:
The status of a question that matters to a chapter and that no source answers yet.
_Avoid_: open question, TBD

**Ask**:
A capability or behavior this project would use and that no public ModuleWise source provides, recorded with the GeoLineAI scenario that needs it.
_Avoid_: request, wish, gap

## Contributions

**Contribution**:
A change offered to ModuleWise or Componentized that they would want even if GeoLineAI did not exist.
_Avoid_: patch, PR

**Draft**:
The text of an issue or comment written for the author to post, kept in [drafts/](drafts/) until it is posted.
_Avoid_: proposal
