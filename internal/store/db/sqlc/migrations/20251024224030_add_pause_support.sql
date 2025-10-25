-- +goose Up
-- +goose StatementBegin

ALTER TABLE game_state
ADD COLUMN pause_time_remaining_ms INTEGER DEFAULT 300000,
ADD COLUMN paused_at TIMESTAMP,
ADD COLUMN pause_deadline TIMESTAMP;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE game_state
DROP COLUMN pause_time_remaining_ms,
DROP COLUMN paused_at,
DROP COLUMN pause_deadline;

-- +goose StatementEnd
