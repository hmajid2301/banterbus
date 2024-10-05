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
	return nil
}

type JoinLobby struct {
	PlayerNickname string `json:"player_nickname"`
	RoomCode       string `json:"room_code"`
}

func (j *JoinLobby) Validate() error {
	if j.RoomCode == "" {
		return errors.New("room_code is required")
	}
	return nil
}

type StartGame struct {
	RoomCode string `json:"room_code"`
}

func (s *StartGame) Validate() error {
	if s.RoomCode == "" {
		return errors.New("room_code is required")
	}
	return nil
}

type UpdateNickname struct {
	PlayerNickname string `json:"player_nickname"`
}

func (u *UpdateNickname) Validate() error {
	if u.PlayerNickname == "" {
		return errors.New("player_nickname is required")
	}
	return nil
}

type GenerateNewAvatar struct {
}

func (g *GenerateNewAvatar) Validate() error {
	return nil
}

type TogglePlayerIsReady struct {
}

func (t *TogglePlayerIsReady) Validate() error {
	return nil
}
