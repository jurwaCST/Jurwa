-- Service: status-lifecycle (owner: Nyendrak Yoezer Zangmo)
-- Database: status_lifecycle (PostgreSQL 16). Only this service connects to it.
-- Source of truth for the current status of every report.
-- previous_status is kept so an escalation can be undone (Saga compensation).
-- The full history of changes is kept by report-history and audit-log.

CREATE TABLE IF NOT EXISTS report_status (
    report_id       UUID        PRIMARY KEY,         -- report-submission report id
    status          TEXT        NOT NULL DEFAULT 'OPEN',
    previous_status TEXT,                            -- NULL until the first change
    version         INTEGER     NOT NULL DEFAULT 1,  -- optimistic locking: +1 on every change
    updated_by      UUID,                            -- NULL when the system made the change
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT report_status_status_chk CHECK (
        status IN ('OPEN', 'ASSIGNED', 'IN_PROGRESS', 'RESOLVED', 'ESCALATED')),
    CONSTRAINT report_status_previous_chk CHECK (
        previous_status IS NULL OR
        previous_status IN ('OPEN', 'ASSIGNED', 'IN_PROGRESS', 'RESOLVED', 'ESCALATED'))
);

CREATE INDEX IF NOT EXISTS report_status_status_idx ON report_status (status);
