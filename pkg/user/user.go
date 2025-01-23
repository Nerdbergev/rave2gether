package user

type userright int

const (
	Admin userright = iota
	Moderator
	Unprivileged
)

type User struct {
	Username string    `json:"username"`
	Password string    `json:"-"`
	Salt     string    `json:"-"`
	Right    userright `json:"right"`
}
