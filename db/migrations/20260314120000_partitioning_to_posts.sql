-- +goose Up

ALTER TABLE users ADD COLUMN created_at TIMESTAMPTZ DEFAULT now();

ALTER TABLE posts RENAME TO posts_old;

CREATE TABLE posts (
    id SERIAL,
    author_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    like_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE posts_2025
  PARTITION OF posts
  FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
  
CREATE INDEX idx_posts_author_id
ON posts(id);

INSERT INTO posts (id, author_id, content, like_count, created_at)
SELECT id, author_id, content, like_count, created_at
FROM posts_old;

SELECT setval('posts_id_seq', (SELECT MAX(id) FROM posts));

DROP TABLE posts_old;

ALTER TABLE posts ADD CONSTRAINT posts_author_id_fkey 
    FOREIGN KEY (author_id) REFERENCES users(id);


-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS created_at;

ALTER TABLE posts RENAME TO posts_partitioned;

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    author_id INT NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    like_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO posts (id, author_id, content, like_count, created_at)
SELECT id, author_id, content, like_count, created_at
FROM posts_partitioned;

SELECT setval('posts_id_seq', (SELECT MAX(id) FROM posts));

DROP TABLE posts_partitioned CASCADE;