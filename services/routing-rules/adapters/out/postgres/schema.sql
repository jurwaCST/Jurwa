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

-- The three agreed default rules (decision 2).
INSERT INTO routing_rules (category_code, building, target_position) VALUES
    ('HOSTEL',         '', 'DSA'),
    ('ACADEMIC_BLOCK', '', 'DAA'),
    ('OTHERS',         '', 'SSO')
ON CONFLICT (category_code, building) DO NOTHING;
