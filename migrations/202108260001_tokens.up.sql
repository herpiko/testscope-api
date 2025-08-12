CREATE TABLE tokens (
  key TEXT NOT NULL,
  user_id TEXT NOT NULL,
  email_address TEXT NOT NULL,
  auth_provider TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
