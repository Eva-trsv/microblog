-- +goose Up

CREATE TABLE IF NOT EXISTS post_like_counters (
    post_id INT PRIMARY KEY,
    like_count INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS processed_events (
    event_id TEXT PRIMARY KEY
);

-- +goose Down

DROP TABLE IF EXISTS processed_events;
DROP TABLE IF EXISTS post_like_counters;
