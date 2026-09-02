-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS links (
  id SERIAL PRIMARY KEY,
  slug CHAR(6) UNIQUE NOT NULL, 
  redirect_url TEXT NOT NULL,
  click_count INTEGER NOT NULL DEFAULT 0, 
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), 
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  click_limit INTEGER,
  expires TIMESTAMPTZ,
  owner_id INTEGER NOT NULL,
  FOREIGN KEY (owner_id) REFERENCES users(id) 
);

CREATE TABLE IF NOT EXISTS click_events (
  id SERIAL PRIMARY KEY, 
  link_id INTEGER NOT NULL,
  clicked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  FOREIGN KEY (link_id) REFERENCES links(id)
);

-- +goose Down
DROP TABLE IF EXISTS click_events;
DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS users;
