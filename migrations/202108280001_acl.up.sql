CREATE TABLE access_control_lists (
  object_id TEXT NOT NULL,
  object_type TEXT NOT NULL,
  user_id TEXT NOT NULL,
  access TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE TABLE parent_childs (
  parent TEXT NOT NULL,
  child TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
