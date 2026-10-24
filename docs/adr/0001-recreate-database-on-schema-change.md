# Recreate the database on any schema change; refuse a database that differs

One schema file is the only schema. Opening an empty database applies it in one transaction; any other database must already match it, ignoring comments and whitespace, or opening fails, saying the schema differs and how to recreate the database. There are no migrations: the database is local state, and reimporting the shipped line restores the readings.

## Considered options

- **Migrations:** upgrade code, populated old-database fixtures and failure handling for every change, to keep data no release promises to keep.
- **Applying an idempotent schema to whatever database is there on every start:** it adds new tables but leaves a changed table in its old shape, so the code and the database drift apart unnoticed.

## Consequences

- After any change to a schema statement other than its comments or spacing, recreate the database (scripts/recreate-db.sh) and run setup again, or point GEOLINEAI_DB_PATH at a new file. Recreating discards reviews and ID counters with everything else.
- Statements stay idempotent, so two openers that both find the database empty do not collide; an opener that reads it while the other is still creating it is refused and can retry.
- The check compares definitions only. A change to what the import stores inside a column is not detected, so recreate after that too.
- The refusal happens before any application table is read or written, so a mismatched database keeps its contents; opening it does switch it to WAL mode and create its -wal and -shm files.
