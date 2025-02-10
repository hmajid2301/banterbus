-- +goose Up
-- +goose StatementBegin
ALTER TABLE fibbing_it_answers ADD CONSTRAINT fibbing_it_answers_player_round_unique UNIQUE (player_id, round_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE fibbing_it_answers DROP CONSTRAINT fibbing_it_answers_player_round_unique;
-- +goose StatementEnd
