-- Service: identity (owner: Gayley Choden)
-- Database: identity (PostgreSQL 16). Only the identity service connects to it.
-- Holds user accounts and hashed passwords. Roles live in the authorization service.

CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL,
    full_name     TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,              -- bcrypt hash, never the plain password
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_format CHECK (position('@' IN email) > 1)
);

-- Emails are unique regardless of upper or lower case.
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uq ON users (lower(email));
