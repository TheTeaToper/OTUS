-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    description TEXT,
    user_id VARCHAR(100) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time);

-- +goose Down
DROP TABLE IF EXISTS events;
