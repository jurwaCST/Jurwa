-- Service: authorization (owner: Gayley Choden)
-- Database: authorization_db (PostgreSQL 16). Only the authorization service connects to it.
-- ("authorization" is a reserved word in SQL, so the database name has a suffix.)
-- user_id values come from the identity service. There is no foreign key across services.

CREATE TABLE IF NOT EXISTS roles (
    code        TEXT PRIMARY KEY,
    description TEXT NOT NULL
);

INSERT INTO roles (code, description) VALUES
    ('STUDENT',       'Submits reports and sees their own reports'),
    ('CATEGORY_HEAD', 'DSA, DAA or SSO: receives routed reports and reassigns them'),
    ('STAFF',         'Councillor, HoD or estate staff reached through reassignment'),
    ('SYSTEM_ADMIN',  'Full technical control: rules, users, lists'),
    ('PRESIDENT',     'College President: read-only dashboard')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_roles (
    user_id    UUID        NOT NULL,                 -- identity users.id
    role_code  TEXT        NOT NULL REFERENCES roles (code),
    granted_by UUID,                                 -- admin user id, NULL for seed data
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_code)
);

CREATE INDEX IF NOT EXISTS user_roles_role_idx ON user_roles (role_code);
