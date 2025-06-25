-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    email text UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE users;
