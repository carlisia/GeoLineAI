# Save a reason the server builds for each reading from its advisory

This extends ADR 0002. A Pass or Fail saves, on the history entry, a reason the server builds for that reading from its stored advisory: the advisory, the results that mattered and the verdict chosen, including whether it differs from the advisory. Unreviewed saves none and clears the current reason. The client sends only the target and the verdict. Saving the reason, rather than rebuilding it on read, keeps history showing exactly what was saved.

## Considered options

- **Client-written reason text:** any caller could store text that history presents as the system's explanation; the reviewer chooses the verdict, and the server owns the words saved with it.
- **Model-written reason text:** a model's words would enter the review record, which models never write.
- **Rebuilding the reason on read:** history would show text that was never saved.
- **Comparing the reason as well as the verdict:** the two differ only after the reason template changes in code, and then re-saving the same verdict would add history entries whose verdict did not change.

## Consequences

- Reasons are built inside the save's write transaction, from the stored advisory, so the plan and the write see the same state.
- The unchanged rule stays verdict-only: a reason is built from an advisory that never changes, so the same verdict on the same reading builds the same reason.
- That a saved reason is null exactly for Unreviewed is enforced by the save's plan, not by a database check, because an entry does not store its new verdict.
- The detail shows the reason a save of Pass or of Fail would record, built the way the save builds it.
