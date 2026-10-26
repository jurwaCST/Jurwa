-- Service: department-directory (owner: Sonam Zangmo)
-- Database: department_directory (PostgreSQL 16). Only this service connects to it.
-- Positions (DSA, DAA, SSO, HoDs, councillors, staff) and who currently holds them.
-- holder_user_id comes from the identity service. No foreign key across services.

CREATE TABLE IF NOT EXISTS departments (
    code       TEXT        PRIMARY KEY,              -- one row per departmental HoD area
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS positions (
    code            TEXT        PRIMARY KEY,         -- e.g. DSA, DAA, SSO, HOD_<dept>
    title           TEXT        NOT NULL,
    kind            TEXT        NOT NULL,
    department_code TEXT        REFERENCES departments (code),
    holder_user_id  UUID,                            -- NULL while the position is vacant
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT positions_kind_chk CHECK (kind IN ('CATEGORY_HEAD', 'HOD', 'COUNCILLOR', 'STAFF'))
);

CREATE INDEX IF NOT EXISTS positions_holder_idx ON positions (holder_user_id);
CREATE INDEX IF NOT EXISTS positions_kind_idx   ON positions (kind) WHERE is_active;

-- The three Category Head positions (decision 2). Holders and the 5 HoD
-- departments are entered by the System Administrator.
INSERT INTO positions (code, title, kind) VALUES
    ('DSA', 'Dean of Student Affairs',  'CATEGORY_HEAD'),
    ('DAA', 'Dean of Academic Affairs', 'CATEGORY_HEAD'),
    ('SSO', 'Student Service Officer',  'CATEGORY_HEAD')
ON CONFLICT (code) DO NOTHING;
