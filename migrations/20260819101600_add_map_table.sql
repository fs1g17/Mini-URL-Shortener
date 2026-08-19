-- +goose Up
CREATE TABLE IF NOT EXISTS redirect_map (
  slug varchar(24) not null primary key,
  redirect_url text not null
);

-- +goose Down
DROP TABLE IF EXISTS redirect_map;