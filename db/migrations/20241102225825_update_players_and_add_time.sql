-- +goose Up
-- +goose StatementBegin
ALTER TABLE fibbing_it_rounds ADD COLUMN submit_deadline TIMESTAMP;
CREATE TABLE fibbing_it_rounds_new(
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    round_type TEXT NOT NULL,
    round INT NOT NULL,
    submit_deadline TIMESTAMP NOT NULL,
    fibber_question_id TEXT NOT NULL,
    normal_question_id TEXT NOT NULL,
    room_id TEXT NOT NULL,
    FOREIGN KEY (fibber_question_id) REFERENCES questions (id),
    FOREIGN KEY (normal_question_id) REFERENCES questions (id),
    FOREIGN KEY (room_id) REFERENCES rooms (id)
);

INSERT INTO fibbing_it_rounds_new SELECT * FROM fibbing_it_rounds;
DROP TABLE fibbing_it_rounds;
ALTER TABLE fibbing_it_rounds_new RENAME TO fibbing_it_rounds;

DROP TABLE IF EXISTS game_state;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE fibbing_it_rounds DROP COLUMN submit_deadline;

CREATE TABLE IF NOT EXISTS game_state (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    room_id TEXT NOT NULL,
    FOREIGN KEY (room_id) REFERENCES rooms (id)
);

CREATE TABLE fibbing_it_rounds_new AS
SELECT id, created_at, updated_at, round_type, round, fibber_question_id, normal_question_id, room_id, NULL AS game_state_id
FROM fibbing_it_rounds;

ALTER TABLE fibbing_it_rounds_new ADD CONSTRAINT fk_game_state_id FOREIGN KEY (game_state_id) REFERENCES game_state (id);

DROP TABLE fibbing_it_rounds;

ALTER TABLE fibbing_it_rounds_new RENAME TO fibbing_it_rounds;
-- +goose StatementEnd
