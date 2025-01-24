package user

type userright int

const (
	Unprivileged userright = iota
	Moderator
	Admin
)

type User struct {
	Username string    `json:"username"`
	Password string    `json:"-"`
	Salt     string    `json:"-"`
	Right    userright `json:"-"`
	Coins    int       `json:"coins"`
}
