-- Service: location-taxonomy (owner: Sonam Choki)
-- Database: location_taxonomy (PostgreSQL 16). Only this service connects to it.
-- The list students pick a location from: building (hostel or block), wing, area.

CREATE TABLE IF NOT EXISTS locations (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    category_code TEXT        NOT NULL,              -- category-taxonomy code
    building      TEXT        NOT NULL,              -- hostel or block name
    wing          TEXT        NOT NULL DEFAULT '',   -- '' when not applicable
    area          TEXT        NOT NULL DEFAULT '',   -- '' when not applicable
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT locations_uq UNIQUE (category_code, building, wing, area)
);

CREATE INDEX IF NOT EXISTS locations_category_idx ON locations (category_code) WHERE is_active;
