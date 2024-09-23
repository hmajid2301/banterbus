package websockets

import "errors"

type CreateRoom struct {
	GameName       string `json:"game_name"`
	PlayerNickname string `json:"player_nickname"`
}

func (c *CreateRoom) Validate() error {
	if c.GameName == "" {
		return errors.New("game_name is required")
	}
	if c.PlayerNickname == "" {
		return errors.New("player_nickname is required")
	}
	return nil
}

type JoinLobby struct {
	PlayerNickname string `json:"player_nickname"`
	RoomCode       string `json:"room_code"`
}

func (j *JoinLobby) Validate() error {
	if j.PlayerNickname == "" {
		return errors.New("player_nickname is required")
	}
	if j.RoomCode == "" {
		return errors.New("room_code is required")
	}
	return nil
}

type StartGame struct {
	RoomCode string `json:"room_code"`
	PlayerID string `json:"player_id"`
}

func (s *StartGame) Validate() error {
	if s.RoomCode == "" {
		return errors.New("room_code is required")
	}
	if s.PlayerID == "" {
		return errors.New("player_id is required")
	}
	return nil
}

type UpdateNickname struct {
	PlayerNickname string `json:"player_nickname"`
	PlayerID       string `json:"player_id"`
}

func (u *UpdateNickname) Validate() error {
	if u.PlayerNickname == "" {
		return errors.New("player_nickname is required")
	}
	if u.PlayerID == "" {
		return errors.New("player_id is required")
	}
	return nil
}

type GenerateNewAvatar struct {
	PlayerID string `json:"player_id"`
}

func (g *GenerateNewAvatar) Validate() error {
	if g.PlayerID == "" {
		return errors.New("player_id is required")
	}
	return nil
}

type TogglePlayerIsReady struct {
	PlayerID string `json:"player_id"`
}

func (t *TogglePlayerIsReady) Validate() error {
	if t.PlayerID == "" {
		return errors.New("player_id is required")
	}
	return nil
}
