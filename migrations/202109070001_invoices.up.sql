CREATE TABLE invoices (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id TEXT NOT NULL DEFAULT '',
  user_id TEXT NOT NULL DEFAULT '',
  email_address TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL,
  amount numeric NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  items json,

  /* Populated on callback event */
  paid_amount TEXT NOT NULL DEFAULT '',
  payment_method TEXT NOT NULL DEFAULT '',
  payment_channel TEXT NOT NULL DEFAULT '',
  payment_destination TEXT NOT NULL DEFAULT '',
  paid_at TIMESTAMP,

  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE TABLE products (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  amount numeric NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);

INSERT INTO products (id, name, amount) VALUES ('4b6613a8-3c64-41e4-b474-924360caa824', 'standard', 100000);
