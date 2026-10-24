// Generated from Go by TestFrontendVocabulary (backend/internal/server/frontend_test.go); do not edit.
// Rerun: cd backend && go test ./internal/server -run TestFrontend -update

export const LINES = ['line-1'] as const
export type LineId = (typeof LINES)[number]

export const VERDICTS = ['pass', 'fail', 'unreviewed'] as const
export type Verdict = (typeof VERDICTS)[number]

export const ISSUE_CODES = ['missing', 'id_collision'] as const
export type IssueCode = (typeof ISSUE_CODES)[number]

export const COLUMNS = ['id', 'line_id', 'tx1_id', 'tx1_lat', 'tx1_lon', 'tx1_alt', 'tx2_id', 'tx2_lat', 'tx2_lon', 'tx2_alt', 'rx1_id', 'rx1_lat', 'rx1_lon', 'rx1_alt', 'rx2_id', 'rx2_lat', 'rx2_lon', 'rx2_alt', 'timestamp', 'array_type', 'symmetry', 'stacks', 'input_current', 'contact_resistance', 'apparent_resistivity', 'apparent_resistivity_err', 'chargeability', 'chargeability_err', 'decay_curve', 'qc_review'] as const
export type Column = (typeof COLUMNS)[number]

export const SLOTS = ['tx1', 'tx2', 'rx1', 'rx2'] as const
export type Slot = (typeof SLOTS)[number]

export const ERROR_CODES = ['validation_error', 'reading_not_found', 'not_found', 'body_too_large', 'internal_error'] as const
export type ErrorCode = (typeof ERROR_CODES)[number]

export const ADVISORIES = ['pass', 'fail', 'needs_review'] as const
export type Advisory = (typeof ADVISORIES)[number]

export const CRITERIA = ['import_issue', 'resistivity_sign', 'resistivity_relative_error', 'chargeability_relative_error', 'transmitter_contact', 'decay_trend', 'frozen_repeat', 'repeat_agreement', 'suspect_electrode'] as const
export type Criterion = (typeof CRITERIA)[number]

export const RESULT_STATES = ['met', 'violated', 'review'] as const
export type ResultState = (typeof RESULT_STATES)[number]

export const UNITS = ['ohm-m', 'ohm', 'V/V', 'fraction', 'V/V per ms', ''] as const
export type Unit = (typeof UNITS)[number]

export const ISSUE_LABELS: Record<IssueCode, string> = { 'missing': 'Missing value', 'id_collision': 'Repeated acquisition ID' }

export const CRITERION_LABELS: Record<Criterion, string> = { 'import_issue': 'Import issue', 'resistivity_sign': 'Resistivity sign', 'resistivity_relative_error': 'Resistivity error', 'chargeability_relative_error': 'Chargeability error', 'transmitter_contact': 'Transmitter contact', 'decay_trend': 'Decay trend', 'frozen_repeat': 'Frozen repeat', 'repeat_agreement': 'Repeat agreement', 'suspect_electrode': 'Suspect electrode' }

export const RESULT_STATE_LABELS: Record<ResultState, string> = { 'met': 'Met', 'violated': 'Violated', 'review': 'Review' }

export const ADVISORY_LABELS: Record<Advisory, string> = { 'pass': 'Pass', 'fail': 'Fail', 'needs_review': 'Needs review' }

export const ADVISORY_MEANINGS: Record<Advisory, string> = { 'pass': 'every criterion met', 'fail': 'a criterion was violated', 'needs_review': 'no criterion was violated but some need review' }

export const VERDICT_LABELS: Record<Verdict, string> = { 'pass': 'Pass', 'fail': 'Fail', 'unreviewed': 'Unreviewed' }

export const RESULT_STATE_ORDER = ['violated', 'review', 'met'] as const
export type ResultStateInOrder = (typeof RESULT_STATE_ORDER)[number]

export const SOURCE_KINDS = ['rule', 'doc', 'result', 'fact'] as const
export type SourceKind = (typeof SOURCE_KINDS)[number]

export const OUTCOMES = ['shown', 'withheld', 'unanswerable', 'unavailable'] as const
export type Outcome = (typeof OUTCOMES)[number]

export const UNAVAILABLE_REASONS = ['no_key', 'provider_error', 'invalid_response'] as const
export type UnavailableReason = (typeof UNAVAILABLE_REASONS)[number]

export const MAX_QUESTION = 300

export const RECORDS_REASON: Record<Verdict, boolean> = { 'pass': true, 'fail': true, 'unreviewed': false }
