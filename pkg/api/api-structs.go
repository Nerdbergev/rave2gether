package api

import "time"

type errorResponse struct {
	Httpstatus   string `json:"httpstatus"`
	Errormessage string `json:"errormessage"`
	RequestURL   string `json:"requesturl"`
}

type addSongRequest struct {
	Queries []string `json:"queries"`
}

type voteRequest struct {
	Upvote bool `json:"upvote"`
}

type positionResponse struct {
	Position time.Duration `json:"position"`
	Length   time.Duration `json:"length"`
}

type authResponse struct {
	Token string `json:"token"`
}

type addUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Right    int    `json:"right"`
}

type passwordChangeRequest struct {
	NewPassword string `json:"newpassword"`
}

type coinsResponse struct {
	Coins int `json:"coins"`
}
