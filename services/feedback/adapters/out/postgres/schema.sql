-- Service: feedback (owner: Gayley Choden)
-- Database: feedback (PostgreSQL 16). Only this service connects to it.
-- A student rates the resolution of their own report once it is Resolved (FR-10).

CREATE TABLE IF NOT EXISTS feedback (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id       UUID        NOT NULL,
    student_user_id UUID        NOT NULL,
    rating          SMALLINT    NOT NULL,
    comment         TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT feedback_report_uq   UNIQUE (report_id),          -- one rating per report
    CONSTRAINT feedback_rating_chk  CHECK (rating BETWEEN 1 AND 5),
    CONSTRAINT feedback_comment_len CHECK (char_length(comment) <= 500)
);

CREATE INDEX IF NOT EXISTS feedback_created_idx ON feedback (created_at);
