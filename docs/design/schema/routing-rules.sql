-- Service: routing-rules (owner: Sonam Zangmo)
-- Database: routing_rules (PostgreSQL 16). Only the routing-rules service connects to it.
-- A rule maps a category (and optionally a building) to a Category Head position.
-- Routing uses only category and location, never the description.

CREATE TABLE IF NOT EXISTS routing_rules (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    category_code   TEXT        NOT NULL,             -- category-taxonomy code
    building        TEXT        NOT NULL DEFAULT '',  -- '' means any building in the category
    target_position TEXT        NOT NULL,             -- department-directory position code, e.g. DSA
    priority        INTEGER     NOT NULL DEFAULT 100, -- lower number wins when several rules match
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT routing_rules_uq UNIQUE (category_code, building)
);

CREATE INDEX IF NOT EXISTS routing_rules_category_idx ON routing_rules (category_code) WHERE is_active;

-- The agreed default rules (SCRUM-158, domain.DefaultRules): Hostel and Others
-- to the SSO, each department of the Academic Block to its HoD, and the DAA for
-- an Academic Block report that names no department.
INSERT INTO routing_rules (category_code, building, target_position) VALUES
    ('HOSTEL',         '', 'SSO'),
    ('ACADEMIC_BLOCK', '', 'DAA'),
    ('OTHERS',         '', 'SSO'),
    ('ACADEMIC_BLOCK', 'Architecture',                         'HOD_ARCH'),
    ('ACADEMIC_BLOCK', 'Civil & Environmental Engineering',    'HOD_CIVIL'),
    ('ACADEMIC_BLOCK', 'Computing Technologies',               'HOD_CT'),
    ('ACADEMIC_BLOCK', 'Electrical & Electronics Engineering', 'HOD_EEE'),
    ('ACADEMIC_BLOCK', 'Mechanical Engineering',               'HOD_MECH'),
    ('ACADEMIC_BLOCK', 'Science & Humanities',                 'HOD_SH')
ON CONFLICT (category_code, building) DO NOTHING;

-- A database seeded before SCRUM-158 sends Hostel to the DSA. Moved once to the
-- SSO; a rule an administrator set to anything else is left alone.
UPDATE routing_rules SET target_position = 'SSO', updated_at = now()
WHERE category_code = 'HOSTEL' AND building = '' AND target_position = 'DSA';
