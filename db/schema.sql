CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    game_name TEXT NOT NULL,
    host_player TEXT NOT NULL,
    room_code TEXT NOT NULL,
    FOREIGN KEY (host_player) REFERENCES players (id)
);


CREATE TABLE IF NOT EXISTS players (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avatar BLOB NOT NULL,
    nickname TEXT NOT NULL
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
