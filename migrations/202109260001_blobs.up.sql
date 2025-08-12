CREATE TABLE blobs (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  filename TEXT NOT NULL,
  content_type TEXT NOT NULL,
  size NUMERIC NOT NULL,
  chunks TEXT DEFAULT '',
  bucket TEXT NOT NULL,

  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);

