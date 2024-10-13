-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = OFF;

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

CREATE TABLE fibbing_it_rounds_new(
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    round_type TEXT NOT NULL,
    round INT NOT NULL,
    fibber_question_id TEXT NOT NULL,
    normal_question_id TEXT NOT NULL,
    game_state_id TEXT NOT NULL,
    FOREIGN KEY (fibber_question_id) REFERENCES questions (id),
    FOREIGN KEY (normal_question_id) REFERENCES questions (id),
    FOREIGN KEY (game_state_id) REFERENCES game_state (id)
);

INSERT INTO fibbing_it_rounds_new SELECT * FROM fibbing_it_rounds;
DROP TABLE fibbing_it_rounds;
ALTER TABLE fibbing_it_rounds_new RENAME TO fibbing_it_rounds;

CREATE TABLE fibbing_it_answers_new (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    answer TEXT NOT NULL,
    player_id TEXT NOT NULL,
    round_id TEXT NOT NULL,
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id)
    FOREIGN KEY (player_id) REFERENCES players(id)
);

INSERT INTO fibbing_it_answers_new (id, created_at, updated_at, answer, round_id)
SELECT id, created_at, updated_at, answer, round_id
FROM fibbing_it_answers;

DROP TABLE fibbing_it_answers;
ALTER TABLE fibbing_it_answers_new RENAME TO fibbing_it_answers;

PRAGMA foreign_keys = ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE questions;
DROP TABLE questions_groups;

PRAGMA foreign_keys = OFF;

CREATE TABLE fibbing_it_rounds_old (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    round_type TEXT NOT NULL,
    round INT NOT NULL,
    fibber_question TEXT NOT NULL,
    normal_question TEXT NOT NULL,
    game_state TEXT NOT NULL,
    FOREIGN KEY (game_state) REFERENCES fibbing_it_rounds (game_state)
);

INSERT INTO fibbing_it_rounds_old SELECT * FROM fibbing_it_rounds;
DROP TABLE fibbing_it_rounds;
ALTER TABLE fibbing_it_rounds_old RENAME TO fibbing_it_rounds;


CREATE TABLE fibbing_it_answers_new (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    answer TEXT NOT NULL,
    round_id TEXT NOT NULL,
    FOREIGN KEY (round_id) REFERENCES fibbing_it_rounds (id)
);

INSERT INTO fibbing_it_answers_new (id, created_at, updated_at, answer, round_id)
SELECT id, created_at, updated_at, answer, round_id
FROM fibbing_it_answers;

DROP TABLE fibbing_it_answers;
ALTER TABLE fibbing_it_answers_new RENAME TO fibbing_it_answers;

PRAGMA foreign_keys = ON;
-- +goose StatementEnd
