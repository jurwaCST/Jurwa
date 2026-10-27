-- Service: report-history (owner: Sonam Choki)
-- Database: report_history (PostgreSQL 16). Only this service connects to it.
-- Read models built from events: one summary row per report (for lists and queues)
-- and a timeline of everything that happened to it (FR-08, FR-18, FR-22).

CREATE TABLE IF NOT EXISTS report_summaries (
    report_id           UUID        PRIMARY KEY,
    reporter_id         UUID        NOT NULL,
    category_code       TEXT        NOT NULL,
    building            TEXT        NOT NULL,
    wing                TEXT        NOT NULL DEFAULT '',
    area                TEXT        NOT NULL DEFAULT '',
    status              TEXT        NOT NULL,
    current_assignee_id UUID,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_summaries_status_chk CHECK (
        status IN ('OPEN', 'ASSIGNED', 'IN_PROGRESS', 'RESOLVED', 'ESCALATED'))
);

CREATE INDEX IF NOT EXISTS report_summaries_reporter_idx
    ON report_summaries (reporter_id, created_at DESC);
CREATE INDEX IF NOT EXISTS report_summaries_assignee_idx
    ON report_summaries (current_assignee_id, status);

CREATE TABLE IF NOT EXISTS report_timeline (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id   UUID        NOT NULL REFERENCES report_summaries (report_id),
    event_type  TEXT        NOT NULL,                -- e.g. report.submitted, status.changed
    from_status TEXT,
    to_status   TEXT,
    actor_id    UUID,                                -- NULL when the system acted
    summary     TEXT        NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS report_timeline_report_idx ON report_timeline (report_id, occurred_at);
