-- +goose Up
-- +goose StatementBegin
ALTER TABLE players ADD COLUMN is_ready BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE players DROP COLUMN is_ready;
-- +goose StatementEnd
