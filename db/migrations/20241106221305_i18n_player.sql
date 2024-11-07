-- +goose Up
-- +goose StatementBegin
ALTER TABLE players ADD COLUMN locale TEXT NOT NULL DEFAULT 'en-GB';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE players DROP COLUMN locale;
-- +goose StatementEnd
