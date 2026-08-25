-- +goose Up
ALTER TABLE redirect_map 
ADD user_id SERIAL;

ALTER TABLE redirect_map
ADD CONSTRAINT fk_users
FOREIGN KEY (user_id)
REFERENCES users(id);

-- +goose Down
ALTER TABLE redirect_map DROP CONSTRAINT fk_users;
ALTER TABLE redirect_map DROP COLUMN user_id;
