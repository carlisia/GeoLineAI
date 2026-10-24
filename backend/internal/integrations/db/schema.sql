-- The only schema. Opening an empty database creates it; any other database
-- must already match it, apart from comments and whitespace, or opening fails
-- (ADR 0001). Statements stay idempotent so two first openers cannot collide.

-- One row per successful file import; a failed file leaves none.
CREATE TABLE IF NOT EXISTS imports (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    source_path    TEXT NOT NULL, -- absolute and resolved in the importing environment
    imported_at    TEXT NOT NULL, -- UTC, RFC 3339 with three fractional digits
    retained_count INTEGER NOT NULL CHECK (retained_count >= 0),
    rejected_count INTEGER NOT NULL CHECK (rejected_count >= 0)
);

-- One row per repeat group: the readings of one import that share all four
-- electrode IDs, numbered in order of first appearance in the file.
CREATE TABLE IF NOT EXISTS repeat_groups (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    import_id INTEGER NOT NULL REFERENCES imports (id),
    line_id   TEXT NOT NULL CHECK (line_id IN ('line-1')),
    tx1_id    TEXT NOT NULL,
    tx2_id    TEXT NOT NULL,
    rx1_id    TEXT NOT NULL,
    rx2_id    TEXT NOT NULL
);

-- Retained readings. A NULL value is unavailable and its import issue says why;
-- qc_review is the current verdict, the one column a save changes.
CREATE TABLE IF NOT EXISTS readings (
    id                       INTEGER PRIMARY KEY AUTOINCREMENT,
    import_id                INTEGER NOT NULL REFERENCES imports (id),
    repeat_group_id          INTEGER REFERENCES repeat_groups (id), -- NULL when a slot is blank
    -- source numbers a row by the physical line its record starts on, and the header
    -- comes first, so no row is below 2.
    row_number               INTEGER NOT NULL CHECK (row_number >= 2),
    line_id                  TEXT NOT NULL CHECK (line_id IN ('line-1')),
    acquisition_id           TEXT NOT NULL, -- verbatim, including blank text
    tx1_id                   TEXT,
    tx1_lat                  REAL,
    tx1_lon                  REAL,
    tx1_alt                  REAL,
    tx2_id                   TEXT,
    tx2_lat                  REAL,
    tx2_lon                  REAL,
    tx2_alt                  REAL,
    rx1_id                   TEXT,
    rx1_lat                  REAL,
    rx1_lon                  REAL,
    rx1_alt                  REAL,
    rx2_id                   TEXT,
    rx2_lat                  REAL,
    rx2_lon                  REAL,
    rx2_alt                  REAL,
    timestamp                TEXT, -- UTC, nine fractional digits, so text order is time order
    array_type               TEXT CHECK (array_type IN ('dipole-dipole')),
    symmetry                 INTEGER CHECK (symmetry >= 1),
    stacks                   INTEGER CHECK (stacks >= 1),
    input_current            REAL,
    contact_resistance       REAL,
    apparent_resistivity     REAL,
    apparent_resistivity_err REAL CHECK (apparent_resistivity_err >= 0),
    chargeability            REAL,
    chargeability_err        REAL CHECK (chargeability_err >= 0),
    decay_curve              TEXT, -- JSON array of ten numbers
    qc_review                TEXT NOT NULL CHECK (qc_review IN ('pass', 'fail', 'unreviewed')),
    -- What the criteria said at import, decided from the results below and never
    -- changed afterwards; the verdict is the reviewer's, this is not.
    advisory                 TEXT NOT NULL CHECK (advisory IN ('pass', 'fail', 'needs_review')),
    -- Whether the reading is left out of its repeats' reference, decided at import with
    -- the advisory: 1 for no resistivity or a violated single-reading result.
    left_out                 INTEGER NOT NULL CHECK (left_out IN (0, 1))
);

-- One row per criterion result of a reading, in the order the criteria are
-- evaluated and shown; value and limit_value are NULL when the result compares none.
CREATE TABLE IF NOT EXISTS results (
    reading_id  INTEGER NOT NULL REFERENCES readings (id),
    position    INTEGER NOT NULL CHECK (position >= 0),
    criterion   TEXT NOT NULL CHECK (criterion IN ('import_issue', 'resistivity_sign', 'resistivity_relative_error', 'chargeability_relative_error', 'transmitter_contact', 'decay_trend', 'frozen_repeat', 'repeat_agreement', 'suspect_electrode')),
    state       TEXT NOT NULL CHECK (state IN ('met', 'violated', 'review')),
    explanation TEXT NOT NULL,
    value       REAL,
    limit_value REAL,
    unit        TEXT NOT NULL CHECK (unit IN ('ohm-m', 'ohm', 'V/V', 'fraction', 'V/V per ms', '')),
    PRIMARY KEY (reading_id, position)
) WITHOUT ROWID;

-- Serves a reading's repeat agreement and a line's results by criterion.
CREATE INDEX IF NOT EXISTS results_by_criterion ON results (criterion, state, reading_id);

-- Serves a line's readings in reading order.
CREATE INDEX IF NOT EXISTS readings_line_order ON readings (line_id, timestamp, id);

-- Serves a group's readings, for the repeats of one reading.
CREATE INDEX IF NOT EXISTS readings_by_group ON readings (repeat_group_id, timestamp, id);

CREATE TABLE IF NOT EXISTS import_issues (
    reading_id  INTEGER NOT NULL REFERENCES readings (id),
    -- Every source column except line_id (which rejects its row) and qc_review
    -- (which is ignored).
    field       TEXT NOT NULL CHECK (field IN ('id', 'tx1_id', 'tx1_lat', 'tx1_lon', 'tx1_alt', 'tx2_id', 'tx2_lat', 'tx2_lon', 'tx2_alt', 'rx1_id', 'rx1_lat', 'rx1_lon', 'rx1_alt', 'rx2_id', 'rx2_lat', 'rx2_lon', 'rx2_alt', 'timestamp', 'array_type', 'symmetry', 'stacks', 'input_current', 'contact_resistance', 'apparent_resistivity', 'apparent_resistivity_err', 'chargeability', 'chargeability_err', 'decay_curve')),
    code        TEXT NOT NULL CHECK (code IN ('missing', 'id_collision')),
    source_text TEXT NOT NULL,
    PRIMARY KEY (reading_id, field, code)
) WITHOUT ROWID;

-- The review record: one action per save, and one history entry per reading the
-- action changed, holding the verdict the change replaced. Both are written only by
-- saves and never rewritten, which the triggers below enforce.
CREATE TABLE IF NOT EXISTS review_actions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    reviewer    TEXT NOT NULL,
    recorded_at TEXT NOT NULL -- UTC, RFC 3339 with three fractional digits
);

CREATE TABLE IF NOT EXISTS review_history (
    action_id        INTEGER NOT NULL REFERENCES review_actions (id),
    reading_id       INTEGER NOT NULL REFERENCES readings (id),
    previous_verdict TEXT NOT NULL CHECK (previous_verdict IN ('pass', 'fail', 'unreviewed')),
    -- The reason saved with the verdict this entry set: NULL exactly for Unreviewed,
    -- which the save's plan enforces, since an entry does not store its new verdict.
    reason           TEXT,
    PRIMARY KEY (action_id, reading_id)
) WITHOUT ROWID;

-- Serves a reading's history, newest action first.
CREATE INDEX IF NOT EXISTS review_history_by_reading ON review_history (reading_id, action_id);

CREATE TRIGGER IF NOT EXISTS review_actions_are_final_update BEFORE UPDATE ON review_actions
BEGIN SELECT RAISE(ABORT, 'review actions are never rewritten'); END;
CREATE TRIGGER IF NOT EXISTS review_actions_are_final_delete BEFORE DELETE ON review_actions
BEGIN SELECT RAISE(ABORT, 'review actions are never rewritten'); END;
CREATE TRIGGER IF NOT EXISTS review_history_is_final_update BEFORE UPDATE ON review_history
BEGIN SELECT RAISE(ABORT, 'review history is never rewritten'); END;
CREATE TRIGGER IF NOT EXISTS review_history_is_final_delete BEFORE DELETE ON review_history
BEGIN SELECT RAISE(ABORT, 'review history is never rewritten'); END;
