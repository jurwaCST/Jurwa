-- Service: reassignment (owner: Nyendrak Yoezer Zangmo)
-- Database: reassignment (PostgreSQL 16). Only this service connects to it.
-- Who a report is assigned to, and every hand-over. Exactly one current row per report.

CREATE TABLE IF NOT EXISTS assignments (
    id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id              UUID        NOT NULL,
    assignee_user_id       UUID        NOT NULL,     -- identity user id
    assignee_position_code TEXT,                     -- department-directory position, if any
    kind                   TEXT        NOT NULL,
    assigned_by            UUID,                     -- NULL when done by the system
    reason                 TEXT        NOT NULL DEFAULT '',
    is_current             BOOLEAN     NOT NULL DEFAULT TRUE,
    assigned_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT assignments_kind_chk CHECK (
        kind IN ('ROUTED', 'REASSIGNED', 'ESCALATED', 'COMPENSATED'))
);

-- Only one current assignment per report.
CREATE UNIQUE INDEX IF NOT EXISTS assignments_current_uq
    ON assignments (report_id) WHERE is_current;
-- A Category Head's or staff user's queue.
CREATE INDEX IF NOT EXISTS assignments_assignee_current_idx
    ON assignments (assignee_user_id) WHERE is_current;
CREATE INDEX IF NOT EXISTS assignments_report_time_idx
    ON assignments (report_id, assigned_at);
