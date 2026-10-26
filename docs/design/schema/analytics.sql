-- Service: analytics (owner: Gayley Choden)
-- Database: analytics (PostgreSQL 16). Only this service connects to it.
-- A read model built from events, used for the College President dashboard (FR-29).

CREATE TABLE IF NOT EXISTS report_facts (
    report_id     UUID        PRIMARY KEY,
    category_code TEXT        NOT NULL,
    building      TEXT        NOT NULL,
    status        TEXT        NOT NULL,
    was_escalated BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL,
    assigned_at   TIMESTAMPTZ,
    resolved_at   TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_facts_status_chk CHECK (
        status IN ('OPEN', 'ASSIGNED', 'IN_PROGRESS', 'RESOLVED', 'ESCALATED'))
);

CREATE INDEX IF NOT EXISTS report_facts_category_status_idx ON report_facts (category_code, status);
CREATE INDEX IF NOT EXISTS report_facts_created_idx         ON report_facts (created_at);
CREATE INDEX IF NOT EXISTS report_facts_escalated_idx       ON report_facts (report_id) WHERE was_escalated;
