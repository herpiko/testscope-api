CREATE TABLE users (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  email_address TEXT NOT NULL,
  full_name TEXT NOT NULL DEFAULT '',
  user_name TEXT NOT NULL DEFAULT '',
  role TEXT NOT NULL DEFAULT 'USER',
  subscription_type TEXT NOT NULL DEFAULT 'free',
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);

ALTER TABLE users ADD CONSTRAINT unique_users_email_address UNIQUE (email_address); 

INSERT INTO users (email_address, role) VALUES ('herpiko@gmail.com', 'ADMIN');
