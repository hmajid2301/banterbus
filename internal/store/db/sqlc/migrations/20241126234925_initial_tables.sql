-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avatar TEXT NOT NULL,
    nickname TEXT NOT NULL,
    is_ready BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    game_name TEXT NOT NULL,
    host_player UUID NOT NULL,
    room_state TEXT NOT NULL,
    room_code TEXT NOT NULL,
    FOREIGN KEY (host_player) REFERENCES players (id)
);

CREATE TABLE IF NOT EXISTS rooms_players (
    room_id UUID NOT NULL,
    player_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, player_id),
    FOREIGN KEY (room_id) REFERENCES rooms (id),
    FOREIGN KEY (player_id) REFERENCES players (id)
);

CREATE TABLE IF NOT EXISTS game_state (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    room_id UUID NOT NULL,
    submit_deadline TIMESTAMP NOT NULL,
    state TEXT NOT NULL,
    FOREIGN KEY (room_id) REFERENCES rooms (id)
);

CREATE TABLE IF NOT EXISTS questions_groups (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    group_name TEXT NOT NULL,
    group_type TEXT NOT NULL
);

CREATE TABLE questions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    game_name TEXT NOT NULL,
    round TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    question TEXT NOT NULL,
    language_code TEXT NOT NULL,
    group_id UUID NOT NULL,
    FOREIGN KEY (group_id) REFERENCES questions_groups (id)
);

CREATE TABLE fibbing_it_rounds (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    round_type TEXT NOT NULL,
    round INT NOT NULL,
    fibber_question_id UUID NOT NULL,
    normal_question_id UUID NOT NULL,
    game_state_id UUID NOT NULL,
    FOREIGN KEY (fibber_question_id) REFERENCES questions (id),
    FOREIGN KEY (normal_question_id) REFERENCES questions (id),
    FOREIGN KEY (game_state_id) REFERENCES game_state (id)
);

CREATE TABLE fibbing_it_answers (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    answer TEXT NOT NULL,
    player_id UUID NOT NULL,
    round_id UUID NOT NULL,
    is_ready BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id),
    FOREIGN KEY (player_id) REFERENCES players (id)
);

CREATE TABLE fibbing_it_player_roles (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    player_role TEXT NOT NULL,
    round_id UUID NOT NULL,
    player_id UUID NOT NULL,
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id),
    FOREIGN KEY (player_id) REFERENCES players (id)
);

CREATE TABLE fibbing_it_votes (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    player_id UUID NOT NULL,
    voted_for_player_id UUID NOT NULL,
    round_id UUID NOT NULL,
    is_ready BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (player_id) REFERENCES players (id),
    FOREIGN KEY (voted_for_player_id) REFERENCES players (id),
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id),
    UNIQUE (player_id, round_id)
);

CREATE TABLE fibbing_it_scores (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    player_id UUID NOT NULL,
    score INT NOT NULL,
    round_id UUID NOT NULL,
    FOREIGN KEY (player_id) REFERENCES players (id),
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id),
    UNIQUE (player_id, round_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS fibbing_it_votes;
DROP TABLE IF EXISTS fibbing_it_player_roles;
DROP TABLE IF EXISTS fibbing_it_answers;
DROP TABLE IF EXISTS fibbing_it_scoring;
DROP TABLE IF EXISTS fibbing_it_rounds;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS questions_groups;
DROP TABLE IF EXISTS game_state;
DROP TABLE IF EXISTS rooms_players;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS players;

-- +goose StatementEnd
