package api

import (
	"time"

	"github.com/Nerdbergev/rave2gether/pkg/config"
	"github.com/Nerdbergev/rave2gether/pkg/queue"
)

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

type currentSongResponse struct {
	Name     string        `json:"name"`
	Position time.Duration `json:"position"`
	Length   time.Duration `json:"length"`
	AddedBy  string        `json:"addedby"`
	AddedAt  time.Time     `json:"addedat"`
	Points   int           `json:"points"`
}

type authResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
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

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type modeResponse struct {
	Mode config.Operatingmode `json:"mode"`
}

type allQueuesResponse struct {
	PrepareQueue  []queue.Entry `json:"preparequeue"`
	DownloadQueue []queue.Entry `json:"downloadqueue"`
	PlayQueue     []queue.Entry `json:"playqueue"`
}
