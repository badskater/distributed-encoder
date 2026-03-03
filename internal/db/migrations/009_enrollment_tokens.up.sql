CREATE TABLE enrollment_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token       TEXT NOT NULL UNIQUE,
    created_by  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_by     UUID REFERENCES agents(id) ON DELETE SET NULL,
    used_at     TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_enrollment_tokens_token ON enrollment_tokens(token);
CREATE INDEX idx_enrollment_tokens_expires ON enrollment_tokens(expires_at);
