-- +goose Up
ALTER TABLE posts ADD COLUMN author_id INT REFERENCES users(id);

CREATE TABLE likes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    post_id INT NOT NULL REFERENCES posts(id),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS likes;
ALTER TABLE posts DROP COLUMN IF EXISTS author_id;