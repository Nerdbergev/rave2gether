package user

type User struct {
	Username string`json:"username"`
	Password string`json:"-"`
	Salt     string`json:"-"`
}
