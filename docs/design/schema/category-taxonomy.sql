-- Service: category-taxonomy (owner: Sonam Zangmo)
-- Database: category_taxonomy (PostgreSQL 16). Only this service connects to it.

CREATE TABLE IF NOT EXISTS categories (
    code        TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The three agreed categories (decisions 1 and 2).
INSERT INTO categories (code, name, description, sort_order) VALUES
    ('HOSTEL',         'Hostel',               'Hostel facilities',         1),
    ('ACADEMIC_BLOCK', 'Academic Block',       'Academic block facilities', 2),
    ('OTHERS',         'Others (campus-wide)', 'Campus-wide facilities',    3)
ON CONFLICT (code) DO NOTHING;
