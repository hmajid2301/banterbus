-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    game_name TEXT NOT NULL,
    host_player TEXT NOT NULL,
    room_state TEXT NOT NULL,
    room_code TEXT NOT NULL,
    FOREIGN KEY (host_player) REFERENCES players (id)
);

CREATE TABLE IF NOT EXISTS players (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avatar BLOB NOT NULL,
    nickname TEXT NOT NULL,
    is_ready BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS rooms_players (
    room_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, player_id),
    FOREIGN KEY (room_id) REFERENCES rooms (id),
    FOREIGN KEY (player_id) REFERENCES players (id)
);

CREATE TABLE IF NOT EXISTS game_state (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    room_id TEXT NOT NULL,
    submit_deadline TIMESTAMP NOT NULL,
    state TEXT NOT NULL,
    FOREIGN KEY (room_id) REFERENCES rooms (id)
);

CREATE TABLE questions (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    game_name TEXT NOT NULL,
    round TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    question TEXT NOT NULL,
    language_code TEXT NOT NULL,
    group_id TEXT NOT NULL,
    FOREIGN KEY (group_id) REFERENCES questions_groups (id)
);

CREATE TABLE questions_groups (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    group_name TEXT NOT NULL,
    group_type TEXT NOT NULL
);

CREATE TABLE fibbing_it_rounds (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    round_type TEXT NOT NULL,
    round INT NOT NULL,
    fibber_question_id TEXT NOT NULL,
    normal_question_id TEXT NOT NULL,
    room_id TEXT NOT NULL,
    game_state_id TEXT NOT NULL,
    FOREIGN KEY (fibber_question_id) REFERENCES questions (id),
    FOREIGN KEY (normal_question_id) REFERENCES questions (id),
    FOREIGN KEY (room_id) REFERENCES rooms (id),
    FOREIGN KEY (game_state_id) REFERENCES game_state (id)
);

CREATE TABLE fibbing_it_answers (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    answer TEXT NOT NULL,
    player_id TEXT NOT NULL,
    round_id TEXT NOT NULL,
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id),
    FOREIGN KEY (player_id) REFERENCES players(id)
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

CREATE TABLE fibbing_it_voting (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    votes INTEGER DEFAULT 0,
    player_id TEXT NOT NULL,
    round_id TEXT NOT NULL,
    FOREIGN KEY (player_id) REFERENCES players (id),
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rooms;
DROP TABLE players;
DROP TABLE rooms_players;
DROP TABLE game_state;
DROP TABLE questions;
DROP TABLE questions_groups;
DROP TABLE fibbing_it_rounds;
DROP TABLE fibbing_it_answers;
DROP TABLE fibbing_it_player_roles;
DROP TABLE fibbing_it_voting;
-- +goose StatementEnd
