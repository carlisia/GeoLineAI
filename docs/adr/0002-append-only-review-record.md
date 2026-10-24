# Keep an append-only review record, written in one write transaction

Every save is a review action with one history entry for the one reading it changed. An entry stores only the verdict its change replaced; the reading's qc_review holds the current verdict, so an entry's new verdict is the next entry's replaced verdict, or qc_review for the latest, derived when the detail is read. Actions and entries are never rewritten, which the database enforces with triggers. A save reads the target's verdict, plans (existence, the unchanged rule) and writes the action, its entry and the reading's verdict in one write transaction that takes SQLite's write lock at its start (BEGIN IMMEDIATE), so two saves run one after the other and a save that reads before it writes waits for an import running in another process instead of failing midway.

## Considered options

- **Each entry storing its new verdict as well as the one it replaced:** every verdict would be stored twice, as one entry's new verdict and the next entry's replaced one, and qc_review would become a duplicate.
- **Leaving qc_review blank until a reading's first save:** Unreviewed would have two spellings, NULL and unreviewed, and every query would test both.
- **Client-held expected revisions, with a 409 on a mismatch:** with one local reviewer, the only conflict is a second tab of the same person; the single write transaction keeps history correct without it.
- **An undo that deletes or edits a past action:** the record would no longer show what happened. A reviewer who changes their mind saves another verdict, which appends.

## Consequences

- The client never sends the state it saw, so a stale page saves over a newer verdict and history records both changes.
- A save of the verdict a reading already has creates no action.
- A save names exactly one reading; there is no bulk write path to keep atomic across readings.
- A save that fails partway, after its action and entry are written, rolls back whole; a test injects that failure.
