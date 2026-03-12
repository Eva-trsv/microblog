-- +goose Up
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_likes_user_id ON likes(user_id);
CREATE INDEX idx_likes_post_id ON likes(post_id);

-- +goose Down
DROP INDEX IF EXISTS idx_posts_author_id;
DROP INDEX IF EXISTS idx_likes_user_id;
DROP INDEX IF EXISTS idx_likes_post_id;