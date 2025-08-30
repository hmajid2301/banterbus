-- +goose Up
-- +goose StatementBegin

CREATE INDEX IF NOT EXISTS idx_rooms_room_code ON rooms (room_code);
CREATE INDEX IF NOT EXISTS idx_rooms_players_room_id ON rooms_players (room_id);
CREATE INDEX IF NOT EXISTS idx_rooms_players_player_id ON rooms_players (player_id);
CREATE INDEX IF NOT EXISTS idx_game_state_room_id ON game_state (room_id);
CREATE INDEX IF NOT EXISTS idx_fibbing_it_rounds_game_state_id ON fibbing_it_rounds (game_state_id);
CREATE INDEX IF NOT EXISTS idx_fibbing_it_votes_round_id ON fibbing_it_votes (round_id);
CREATE INDEX IF NOT EXISTS idx_questions_game_name_round_type_enabled
ON questions (game_name, round_type) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_fibbing_it_answers_player_round ON fibbing_it_answers (player_id, round_id);
CREATE INDEX IF NOT EXISTS idx_fibbing_it_votes_player_round ON fibbing_it_votes (player_id, round_id);
CREATE INDEX IF NOT EXISTS idx_fibbing_it_votes_voted_for_player_id ON fibbing_it_votes (
    voted_for_player_id
);
CREATE INDEX IF NOT EXISTS idx_rooms_room_state ON rooms (room_state);
CREATE INDEX IF NOT EXISTS idx_game_state_state ON game_state (state);
CREATE INDEX IF NOT EXISTS idx_fibbing_it_rounds_game_state_round
ON fibbing_it_rounds (game_state_id, round DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_fibbing_it_rounds_game_state_round;
DROP INDEX IF EXISTS idx_game_state_state;
DROP INDEX IF EXISTS idx_rooms_room_state;
DROP INDEX IF EXISTS idx_fibbing_it_votes_voted_for_player_id;
DROP INDEX IF EXISTS idx_fibbing_it_votes_player_round;
DROP INDEX IF EXISTS idx_fibbing_it_answers_player_round;
DROP INDEX IF EXISTS idx_questions_game_name_round_type_enabled;
DROP INDEX IF EXISTS idx_fibbing_it_votes_round_id;
DROP INDEX IF EXISTS idx_fibbing_it_rounds_game_state_id;
DROP INDEX IF EXISTS idx_game_state_room_id;
DROP INDEX IF EXISTS idx_rooms_players_player_id;
DROP INDEX IF EXISTS idx_rooms_players_room_id;
DROP INDEX IF EXISTS idx_rooms_room_code;

-- +goose StatementEnd
