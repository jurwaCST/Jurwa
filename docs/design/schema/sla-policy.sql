-- Service: sla-policy (owner: Nyendrak Yoezer Zangmo)
-- Database: sla_policy (PostgreSQL 16). Only the sla-policy service connects to it.
-- One SLA window per category (FR-13, FR-28). The window values are not decided
-- yet and are set by the System Administrator, so no rows are seeded here.

CREATE TABLE IF NOT EXISTS sla_policies (
    category_code TEXT        PRIMARY KEY,           -- category-taxonomy code
    window_hours  INTEGER     NOT NULL,
    updated_by    UUID,                              -- admin user id
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sla_policies_window_chk CHECK (window_hours > 0)
);
