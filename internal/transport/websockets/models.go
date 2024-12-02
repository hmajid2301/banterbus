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

type KickPlayer struct {
	RoomCode             string `json:"room_code"`
	PlayerNicknameToKick string `json:"player_nickname_to_kick"`
}

func (k *KickPlayer) Validate() error {
	if k.RoomCode == "" {
		return errors.New("room_code is required")
	}

	if k.PlayerNicknameToKick == "" {
		return errors.New("player_nickname_to_kick is required")
	}
	return nil
}

type SubmitAnswer struct {
	Answer string
}

func (s *SubmitAnswer) Validate() error {
	if s.Answer == "" {
		return errors.New("answer is required")
	}
	return nil
}

type SubmitVote struct {
	VotedPlayerNickname string `json:"voted_player_nickname"`
}

func (s *SubmitVote) Validate() error {
	if s.VotedPlayerNickname == "" {
		return errors.New("player nickname is required")
	}

	return nil
}

type ToggleAnswerIsReady struct {
}

func (t *ToggleAnswerIsReady) Validate() error {
	return nil
}
