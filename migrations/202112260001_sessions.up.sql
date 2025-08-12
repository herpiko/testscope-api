DO $$ BEGIN
  CREATE EXTENSION pgcrypto;
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE sessions (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  author_id UUID NOT NULL,
  project_id UUID NOT NULL,
  scenarios TEXT[],
  version TEXT NOT NULL, 
  description TEXT, 
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  status INT DEFAULT 0, /* 0: open, 1: closed, 2: archived */
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  FOREIGN KEY (author_id) REFERENCES users(id) ON UPDATE CASCADE,
  FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE
);

CREATE TABLE tests (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  assignee_id UUID NOT NULL,
  session_id UUID NOT NULL,
  scenario_id UUID NOT NULL,
  steps jsonb,
  status INT DEFAULT 0, /* 0: unassigned, 1: ontest, 2: passed, 3: fail */
  notes TEXT DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES sessions(id) ON UPDATE CASCADE,
  FOREIGN KEY (assignee_id) REFERENCES users(id) ON UPDATE CASCADE
);
