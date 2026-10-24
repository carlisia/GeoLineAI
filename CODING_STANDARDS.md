# Coding standards

Review rules for judgement calls no tool can make. Mechanical rules are enforced by `scripts/check.sh`; run it rather than reviewing for them. `CLAUDE.md` holds the naming rule (the Terms in docs/domain.md) and the invariants; apply those too.

Apply every rule below to every changed file.

## One home

A limit, a validation or a count lives in one constant or function, and every caller uses that one. A second copy of the same test is a defect even when the copies agree today: they drift the first time one is edited. When two modules need the same test, the one that defines the concept owns it and exports it. The schema cannot call Go, so a limit it repeats as a CHECK is pinned to its Go constant by a test in the SQLite adapter.

## Values have one spelling

A domain value (verdict, result state, advisory value, criterion name) is written through its constant everywhere outside the constant's definition, and SQL receives it as a bound argument. A value set that both the backend and the frontend hold is generated from Go into one frontend file, `vocabulary.gen.ts`; every response type the frontend reads is generated from the Go type the server answers with, into one `*.gen.ts` file per Go package; and every response is pinned as a literal in a golden file, `responses.golden.ts`. The frontend imports these and holds no copy. In TypeScript a value of a generated union type may be written as its literal (`'unreviewed'`), since the compiler rejects any spelling outside the set; the rule there is that the type comes from the generated file. A new shared value set or a new endpoint's response joins its generator; each file's header names the test that writes it and the command that reruns it. Each schema CHECK list is pinned to its Go list by a test in the SQLite adapter. Tests and the shared test helpers send wire values as a client would, so they keep literals.

## Fail fast on the impossible

A function that handles only some values of its input handles the rest explicitly: an error, or a panic naming the value when reaching it is a programming error. A default branch that maps every other value to a real answer is a defect, and so is a map lookup whose zero value reaches output: over a closed set of values, check the key is present or name every value in a switch.

## Stated assumptions

When correctness rests on a guarantee another module provides, the code that relies on it says so in a comment: one sentence stating what must hold, whole on its own. The test is a rename: the sentence stays true after every other file in this repository is renamed, so it names none of their contents, in parentheses or otherwise. Where a name would help, say what the thing does.

## Comments

A comment on a declaration, the doc comment, says what the thing is for its caller and what it guarantees; how it is built belongs in the code, or in a comment inside the function where the reason is not evident from the code alone. No comment names a spec, an ADR, a standards document, a test, a script or a file path: those move and get renamed, and the comment is then a pointer that leads nowhere. What a rule is for can be said in the comment itself, in a sentence; where the rule came from is the specs' and the ADRs' job, and they are read on their own. The one exception is a generated file, whose header names the test that writes it and the command that reruns it, because a reader of that file has to know not to edit it. Nor does a comment name a person: the human at the page is the reviewer, the term docs/domain.md defines, and a name ties the code to one person the way a file path ties it to one layout. The check script fails on any comment that breaks this.

A package with two or more files keeps its package comment in its own file, `doc.go`, holding nothing else, so the package's summary has one place whichever file a reader opens.

## Testing depth

A test exists to catch a **consequential**, **silent** failure. Consequential: the reviewer acts on a wrong answer or loses work. Silent: the browser check before hand-off (`CLAUDE.md`) would not reveal it. **Visible** failures (a crash, a blank screen, wrong text a glance catches) are the browser check's job; types, framework behaviour and wire shapes no caller reads need no test. A bug that reached the user past the browser check was silent by evidence: when its failure is consequential, the fix arrives with a test at the seam that owns the rule, **red** on the code before the fix. A bug the browser check itself caught was visible, and its fix arrives with none.

These rules decide which tests exist for every process that writes them, a skill's red-green loop included: a slice gets a test only when it meets this paragraph, at a seam the table below names, so no seam needs agreeing, and the test is proved the one way below (Every test can go red), whichever order the test and its code were written in.

### The ladder

Each rung costs more code than the one above. Go down a rung only when the rung above leaves a consequential failure silent:

1. **Outcome:** one test per consequential behaviour docs/spec.md's acceptance criteria assign, at the seam that owns it (Tests drive a seam).
2. **Limits:** table rows for equal, just over, zero and absent. A new case is a row, not a function.
3. **Derivations and order:** a value computed from other rows (a neighbouring entry, the latest of a set, a join), or an ordering, has a test that goes red when the derivation is reversed, pointed at the wrong row or removed. End-state tests alone pass while a derivation is wrong.
4. **Atomicity and races:** only where the spec promises all-or-nothing.

### Every test can go red

Name each test after the failure it catches; a failure you cannot state in one sentence gets no test. Expected values come from an independent source: a known literal, a worked example, the spec. A test that recomputes its expected value the way the code does is **tautological**: it passes by construction and can never disagree with the code. A test that breaks on a refactor while behaviour holds is **implementation-coupled**: assert only what a caller observes at the seam.

Prove each new test can fail: break the code it guards (revert the fix, invert the condition, remove the derivation), watch the test go **red**, then restore the code. `scripts/mutate.py FILE OLD NEW` makes that check: it builds first, so a mutant that does not compile is reported INVALID rather than caught, names the tests that fail, and restores the file. A test that stays green against broken code tests nothing; fix it or delete it.

Fixtures that write data do so in a fixed order, from a slice, never by ranging over a map: Go randomises map order, so the IDs the database assigns change from run to run, and an assertion that names one passes most runs and fails the rest. Expected IDs come from the fixture's own state, not from literals, and a new test runs once with `-count=20` before it is trusted.

### Doubles and probes

A test double for an external library encodes an assumption about that library; before writing one, observe the real library once in a browser and make the double reproduce what was observed.

A measurement taken in the browser to prove a change (counting matching pixels, rows or elements) is checked on the failing state as well as the passing one: run it before and after the change in the same state, and match on the property that tells the cases apart rather than on an exact value, because a probe tuned on an isolated example can read zero on a crowded one.

### Budget

`scripts/check.sh` fails when test code outgrows its budget: `scripts/test-budget.py` measures it, `TEST_BUDGET` sets it, and the target is 0.5. Bring it back under by folding test functions into table rows first, then cutting tests from the deepest rung up.

## Tests drive a seam

A test drives one seam and asserts only what a caller observes there: responses and status codes, persisted state after reopening the database, rendered text and the requests sent. The import command has no test seam of its own: scripts/setup.sh and scripts/verify.sh exercise its output and exit codes. A rule is tested at the seam that owns it, once.

| Seam | What a test drives | What it asserts |
| --- | --- | --- |
| HTTP router | The router the server runs, through `httptest`, against a temporary on-disk database | Status codes and bodies of every endpoint, and persisted state read back through the API, including after reopening the database |
| Rendered app | The app rendered with Testing Library in jsdom against a fake backend | The requests sent, and what a glance would not catch: superseded responses, locking, a kept choice after a failed save |
| Pure functions (exception) | A function called directly | Only rules whose cases are too many to push through a higher seam: the criteria, the repeat median, the parser, the provider adapter, the saved-reason template. The HTTP seam then checks only that a save records its preview, built from the reading's own evidence |

## Callers beyond the tests

Every field, code or function a module exposes has a reader outside its own tests: the frontend, another package, or a person reading a message. When a test is the only reader, remove the surface and move the assertion to what a caller does read (Tests drive a seam). Text a person reads is the right surface even though rewording breaks it: the one failing assertion names the change. The generated response types in `frontend/src/lib/*.gen.ts` are the exception: each describes a whole response, fields the frontend ignores included, because it is generated from the Go type the server answers with, and `responses.golden.ts` checks real responses against it with `satisfies` (Values have one spelling), a check that rejects a field the type leaves out.

## Pinned tools

A tool the checks depend on is pinned in the repository (`tool` in `backend/go.mod`, `devDependencies` in `frontend/package.json`), never taken from whatever is installed on the machine.
