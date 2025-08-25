-- +goose Up
-- +goose StatementBegin
-- First, remove duplicates keeping the most recent record based on created_at
DELETE FROM fibbing_it_answers a1
USING fibbing_it_answers a2
WHERE
    a1.player_id = a2.player_id
    AND a1.round_id = a2.round_id
    AND a1.created_at < a2.created_at;

-- Handle case where created_at is identical (keep the one with higher id)
DELETE FROM fibbing_it_answers a1
USING fibbing_it_answers a2
WHERE
    a1.player_id = a2.player_id
    AND a1.round_id = a2.round_id
    AND a1.created_at = a2.created_at
    AND a1.id < a2.id;

-- Add the unique constraint if it doesn't already exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fibbing_it_answers_player_round_unique' 
        AND table_name = 'fibbing_it_answers'
    ) THEN
        ALTER TABLE fibbing_it_answers ADD CONSTRAINT fibbing_it_answers_player_round_unique UNIQUE (player_id, round_id);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE fibbing_it_answers DROP CONSTRAINT IF EXISTS fibbing_it_answers_player_round_unique;
-- +goose StatementEnd
