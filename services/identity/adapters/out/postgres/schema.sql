-- Service: identity (owner: Gayley Choden)
-- Database: identity (PostgreSQL 16). Only the identity service connects to it.
-- Holds user accounts, hashed passwords and each user's roles. The roles go
-- into the token; the authorization service decides what each role may do.

CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL,
    full_name     TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,              -- bcrypt hash, never the plain password
    roles         TEXT[]      NOT NULL DEFAULT ARRAY['STUDENT'], -- jurwa.common.v1.Role names
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_format CHECK (position('@' IN email) > 1),
    CONSTRAINT users_roles_not_empty CHECK (cardinality(roles) > 0)
);

-- Emails are unique regardless of upper or lower case.
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uq ON users (lower(email));
