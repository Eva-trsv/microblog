-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    author TEXT NOT NULL,
    content TEXT NOT NULL,
    like_count INTEGER NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE posts;
DROP TABLE users;