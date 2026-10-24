# GeoLineAI and ModuleWise

This branch uses GeoLineAI as a fixed application for learning what ModuleWise's composable runtime provides today, one boundary at a time, and for finding contributions ModuleWise would want on their own merits. It never merges into `main`; `main` stays the application its specification describes, and the tag `baseline` marks the commit every chapter is compared against.

- [GLOSSARY.md](GLOSSARY.md): the showcase's terms.
- [evidence.md](evidence.md): every claim about ModuleWise, with its status and proof.
- [features.md](features.md): the conversions and extensions, in chapter order, and the asks.
- [contributions.md](contributions.md): candidate contributions and their drafts.
- [adr/](adr/): decisions that hold for every chapter.

## Rules on this branch

- Nothing about ModuleWise is claimed without an evidence row. Inferred is not verified, and an unknown that a chapter depends on is settled by a proof of concept before GeoLineAI code changes.
- A conversion keeps the user task, the data and every invariant in the root CLAUDE.md. The only scope it adds is the one boundary its chapter names.
- An extension may add behavior the specification rules out, and only here.
- Drafts are posted by the author, never by an agent.
