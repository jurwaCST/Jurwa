-- Service: notification (owner: Gayley Choden)
-- Database: notification (PostgreSQL 16). Only this service connects to it.

CREATE TABLE IF NOT EXISTS notifications (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_user_id UUID        NOT NULL,
    channel           TEXT        NOT NULL,
    event_type        TEXT        NOT NULL,          -- e.g. status.changed, sla.lapsed, digest
    report_id         UUID,                          -- NULL for digests
    subject           TEXT        NOT NULL,
    body              TEXT        NOT NULL,
    delivery_status   TEXT        NOT NULL DEFAULT 'PENDING',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at           TIMESTAMPTZ,
    read_at           TIMESTAMPTZ,
    saved_seq         BIGINT      GENERATED ALWAYS AS IDENTITY, -- save order, breaks ties on created_at
    CONSTRAINT notifications_channel_chk CHECK (channel IN ('EMAIL', 'IN_APP')),
    CONSTRAINT notifications_status_chk  CHECK (delivery_status IN ('PENDING', 'SENT', 'FAILED'))
);

CREATE INDEX IF NOT EXISTS notifications_recipient_idx
    ON notifications (recipient_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS notifications_pending_idx
    ON notifications (created_at) WHERE delivery_status = 'PENDING';
