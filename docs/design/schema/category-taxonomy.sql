-- Service: category-taxonomy (owner: Sonam Zangmo)
-- Database: category_taxonomy (PostgreSQL 16). Only this service connects to it.

CREATE TABLE IF NOT EXISTS categories (
    code          TEXT        PRIMARY KEY,
    name          TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    sort_order    INTEGER     NOT NULL DEFAULT 0,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    head_position TEXT        NOT NULL DEFAULT '',   -- department-directory position code: DSA, DAA or SSO
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Added in Sprint 3 (SCRUM-124); safe on a table created from the earlier file.
ALTER TABLE categories ADD COLUMN IF NOT EXISTS head_position TEXT NOT NULL DEFAULT '';

-- The three agreed categories (decisions 1 and 2), the same values as
-- services/category-taxonomy/domain DefaultCategories.
INSERT INTO categories (code, name, description, sort_order, head_position) VALUES
    ('HOSTEL',         'Hostel',         'Rooms, bathrooms, water and electricity in the hostels', 1, 'DSA'),
    ('ACADEMIC_BLOCK', 'Academic Block', 'Classrooms, labs and equipment in the academic blocks',  2, 'DAA'),
    ('OTHERS',         'Others',         'Campus places such as the library, mess, gym and sports grounds', 3, 'SSO')
ON CONFLICT (code) DO NOTHING;

-- SCRUM-150: Others are places, not services (docs/design/real-locations.md).
-- A database seeded before this still has the old text, which the INSERT above
-- does not change. This replaces only that old default, so an administrator's
-- own wording is kept.
UPDATE categories SET description = 'Campus places such as the library, mess, gym and sports grounds'
WHERE code = 'OTHERS' AND description = 'Campus-wide issues such as Wi-Fi, roads and grounds';
