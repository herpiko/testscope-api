DO $$ BEGIN
  CREATE EXTENSION pgcrypto;
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE scenarios (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  scope_id UUID NOT NULL,
  project_id UUID NOT NULL,
  name TEXT NOT NULL,
  steps jsonb,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  FOREIGN KEY (scope_id) REFERENCES scopes(id) ON UPDATE CASCADE,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE
);
