-- +goose Up

CREATE INDEX IF NOT EXISTS idx_posts_user_id
ON posts(user_id);

CREATE INDEX IF NOT EXISTS idx_likes_post_id
ON likes(post_id);

CREATE INDEX IF NOT EXISTS idx_likes_user_id
ON likes(user_id);


CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_posts_content_trgm
ON posts
USING GIN (content gin_trgm_ops);



-- +goose Down

DROP INDEX IF EXISTS idx_posts_user_id;
DROP INDEX IF EXISTS idx_likes_post_id;
DROP INDEX IF EXISTS idx_likes_user_id;
DROP INDEX IF EXISTS idx_posts_content_trgm;