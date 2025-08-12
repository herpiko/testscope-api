DO $$ BEGIN
  CREATE EXTENSION pgcrypto;
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

ALTER TABLE tests ADD COLUMN assists TEXT[];
