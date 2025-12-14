CREATE TABLE oauth_clients (
  id          UUID PRIMARY KEY,
  name        TEXT NOT NULL,
  secret      TEXT NOT NULL,
  grant_types TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE device_codes (
  device_code  TEXT PRIMARY KEY,
  user_code    TEXT NOT NULL UNIQUE,
  client_id    UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
  user_id      UUID REFERENCES users(id),
  scope        TEXT[],
  status       TEXT NOT NULL DEFAULT 'pending',
  expires_at   TIMESTAMPTZ NOT NULL,
  interval_sec INT NOT NULL DEFAULT 5,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_device_codes_status ON device_codes(status);
CREATE INDEX idx_device_codes_expires_at ON device_codes(expires_at);

CREATE TABLE oauth_tokens (
  id             UUID PRIMARY KEY,
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  client_id      UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
  access_token   TEXT NOT NULL UNIQUE,
  refresh_token  TEXT UNIQUE,
  scope          TEXT[],
  expires_at     TIMESTAMPTZ NOT NULL,
  revoked_at     TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_oauth_tokens_access_token ON oauth_tokens(access_token);
CREATE INDEX idx_oauth_tokens_refresh_token ON oauth_tokens(refresh_token);
CREATE INDEX idx_oauth_tokens_user ON oauth_tokens(user_id);
