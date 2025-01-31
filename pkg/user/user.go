package user

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Userright int

const (
	Unprivileged Userright = iota
	Moderator
	Admin
)

const bcryptCost = 15

type User struct {
	Username string    `json:"username"`
	Password string    `json:"-"`
	Right    Userright `json:"right"`
	Coins    int       `json:"coins"`
}

type UserDB struct {
	users map[string]User
	mutex sync.Mutex
}

// Hash password with bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", errors.New("Error hashing password: " + err.Error())
	}
	return string(hash), nil
}

// Verify password with bcrypt
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (u *User) SetPassword(password string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return checkPasswordHash(password, u.Password)
}

func (u *User) SetRight(right Userright) {
	u.Right = right
}

func (u *User) SetCoins(coins int) {
	u.Coins = coins
}

func (u *User) AddCoins(coins int) {
	u.Coins += coins
}

func (u *User) RemoveCoins(coins int) {
	u.Coins -= coins
}

func (u *User) GetCoins() int {
	return u.Coins
}

func (u *User) GetRight() Userright {
	return u.Right
}

func (u *User) IsModerator() bool {
	return u.GetRight() >= Moderator
}

func (u *User) IsAdmin() bool {
	return u.GetRight() >= Admin
}

func (udb *UserDB) DoesUserExist(username string) bool {
	udb.mutex.Lock()
	defer udb.mutex.Unlock()
	_, ok := udb.users[username]
	return ok
}

func (udb *UserDB) GetUser(username string) (User, error) {
	udb.mutex.Lock()
	defer udb.mutex.Unlock()
	u, ok := udb.users[username]
	if !ok {
		return User{}, errors.New("User does not exist")
	}
	return u, nil
}

func (udb *UserDB) AddUser(username string, password string) error {
	if udb.DoesUserExist(username) {
		return errors.New("User already exists")
	}
	u := User{Username: username}
	err := u.SetPassword(password)
	if err != nil {
		return errors.New("Error setting password: " + err.Error())
	}
	udb.mutex.Lock()
	udb.users[username] = u
	udb.mutex.Unlock()
	return nil
}

func (udb *UserDB) RemoveUser(username string) error {
	if !udb.DoesUserExist(username) {
		return errors.New("User does not exist")
	}
	udb.mutex.Lock()
	delete(udb.users, username)
	udb.mutex.Unlock()
	return nil
}

func (udb *UserDB) UpdateUser(username string, password string, right Userright) error {
	if !udb.DoesUserExist(username) {
		return errors.New("User does not exist")
	}
	u := User{Username: username}
	err := u.SetPassword(password)
	if err != nil {
		return errors.New("Error setting password: " + err.Error())
	}
	u.SetRight(right)
	udb.mutex.Lock()
	udb.users[username] = u
	udb.mutex.Unlock()
	return nil
}

func (udb *UserDB) ListUsers() []User {
	udb.mutex.Lock()
	defer udb.mutex.Unlock()
	res := make([]User, 0, len(udb.users))
	for _, u := range udb.users {
		res = append(res, u)
	}
	return res
}

func (udb *UserDB) SetUserCoins(username string, coins int) error {
	if !udb.DoesUserExist(username) {
		return errors.New("User does not exist")
	}
	udb.mutex.Lock()
	u := udb.users[username]
	u.SetCoins(coins)
	udb.users[username] = u
	udb.mutex.Unlock()
	return nil
}

func (udb *UserDB) SaveToFile(filename string) error {
	udb.mutex.Lock()
	defer udb.mutex.Unlock()
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New("Error opening file: " + err.Error())
	}
	defer f.Close()

	for _, user := range udb.users {
		line := fmt.Sprintf("%s:%s:%s\n", user.Username, user.Password, strconv.Itoa(int(user.Right)))
		_, err := f.WriteString(line)
		if err != nil {
			return errors.New("Error writing to file: " + err.Error())
		}
	}

	return nil
}

func (udb *UserDB) LoadFromFile(filename string) error {
	udb.mutex.Lock()
	defer udb.mutex.Unlock()
	f, err := os.Open(filename)
	if err != nil {
		return errors.New("Error opening file: " + err.Error())
	}
	defer f.Close()

	udb.users = make(map[string]User)

	var username, password string
	var right int
	line := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line++
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) != 3 {
			return errors.New("Wrong user:pass:right format in line " + strconv.Itoa(line))
		}
		username = parts[0]
		password = parts[1]
		right, err = strconv.Atoi(parts[2])
		if err != nil {
			return errors.New("Error parsing right in line " + strconv.Itoa(line) + ": " + err.Error())
		}
		u := User{Username: username, Password: password, Right: Userright(right)}
		udb.users[username] = u
	}

	return nil
}
