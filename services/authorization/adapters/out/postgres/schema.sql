-- Service: authorization (owner: Gayley Choden)
-- Database: authorization_db (PostgreSQL 16). Only the authorization service connects to it.
-- ("authorization" is a reserved word in SQL, so the database name has a suffix.)
-- Holds the permission table: which roles may do each action. A user's roles
-- are stored by identity and travel in the token, so there is no user table here.

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

-- One row per (action, role) that is allowed. Anything not listed is denied.
-- Actions are the ones in docs/design/api-routes.md.
CREATE TABLE IF NOT EXISTS role_permissions (
    action     TEXT        NOT NULL,
    role_code  TEXT        NOT NULL REFERENCES roles (code),
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (action, role_code)
);

-- The default table, the same as DefaultPolicy in services/authorization/domain.
INSERT INTO role_permissions (action, role_code) VALUES
    ('me.read', 'STUDENT'), ('me.read', 'CATEGORY_HEAD'), ('me.read', 'STAFF'), ('me.read', 'SYSTEM_ADMIN'), ('me.read', 'PRESIDENT'),
    ('category.list', 'STUDENT'), ('category.list', 'CATEGORY_HEAD'), ('category.list', 'STAFF'), ('category.list', 'SYSTEM_ADMIN'), ('category.list', 'PRESIDENT'),
    ('location.list', 'STUDENT'), ('location.list', 'CATEGORY_HEAD'), ('location.list', 'STAFF'), ('location.list', 'SYSTEM_ADMIN'), ('location.list', 'PRESIDENT'),
    ('notification.list', 'STUDENT'), ('notification.list', 'CATEGORY_HEAD'), ('notification.list', 'STAFF'), ('notification.list', 'SYSTEM_ADMIN'), ('notification.list', 'PRESIDENT'),
    ('photo.upload', 'STUDENT'),
    ('photo.read', 'STUDENT'), ('photo.read', 'CATEGORY_HEAD'), ('photo.read', 'STAFF'), ('photo.read', 'SYSTEM_ADMIN'),
    ('report.submit', 'STUDENT'),
    ('report.read', 'STUDENT'), ('report.read', 'CATEGORY_HEAD'), ('report.read', 'STAFF'), ('report.read', 'SYSTEM_ADMIN'),
    ('report.list_own', 'STUDENT'),
    ('report.list_assigned', 'CATEGORY_HEAD'), ('report.list_assigned', 'STAFF'),
    ('report.timeline.read', 'STUDENT'), ('report.timeline.read', 'CATEGORY_HEAD'), ('report.timeline.read', 'STAFF'), ('report.timeline.read', 'SYSTEM_ADMIN'),
    ('report.status.read', 'STUDENT'), ('report.status.read', 'CATEGORY_HEAD'), ('report.status.read', 'STAFF'), ('report.status.read', 'SYSTEM_ADMIN'),
    ('report.status.change', 'CATEGORY_HEAD'), ('report.status.change', 'STAFF'),
    ('report.reassign', 'CATEGORY_HEAD'), ('report.reassign', 'STAFF'),
    ('report.assignments.read', 'CATEGORY_HEAD'), ('report.assignments.read', 'STAFF'), ('report.assignments.read', 'SYSTEM_ADMIN'),
    ('report.notes.add', 'CATEGORY_HEAD'), ('report.notes.add', 'STAFF'),
    ('report.notes.read', 'CATEGORY_HEAD'), ('report.notes.read', 'STAFF'), ('report.notes.read', 'SYSTEM_ADMIN'),
    ('report.feedback.submit', 'STUDENT'),
    ('report.feedback.read', 'STUDENT'), ('report.feedback.read', 'CATEGORY_HEAD'), ('report.feedback.read', 'SYSTEM_ADMIN'),
    ('report.audit.read', 'SYSTEM_ADMIN'),
    ('department.list', 'CATEGORY_HEAD'), ('department.list', 'SYSTEM_ADMIN'),
    ('routing_rule.read', 'SYSTEM_ADMIN'),
    ('routing_rule.write', 'SYSTEM_ADMIN'),
    ('sla_policy.read', 'SYSTEM_ADMIN'),
    ('sla_policy.write', 'SYSTEM_ADMIN'),
    ('dashboard.read', 'SYSTEM_ADMIN'), ('dashboard.read', 'PRESIDENT')
ON CONFLICT (action, role_code) DO NOTHING;
