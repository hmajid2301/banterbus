-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS game_state (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    room_id TEXT NOT NULL,
    FOREIGN KEY (room_id) REFERENCES rooms (id)
);

CREATE TABLE fibbing_it_rounds (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    round_type TEXT NOT NULL,
    round INT NOT NULL,
    fibber_question TEXT NOT NULL,
    normal_question TEXT NOT NULL,
    game_state_id TEXT NOT NULL,
    FOREIGN KEY (game_state_id) REFERENCES game_state (id)
);

CREATE TABLE fibbing_it_answers (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    answer TEXT NOT NULL,
    round_id TEXT NOT NULL,
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id)
);

CREATE TABLE fibbing_it_player_roles (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    player_role TEXT NOT NULL,
    round_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id),
    FOREIGN KEY (player_id) REFERENCES players (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE fibbing_it_player_roles;
DROP TABLE fibbing_it_answers;
DROP TABLE fibbing_it_rounds;
DROP TABLE game_state;
-- +goose StatementEnd
