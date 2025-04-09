package config

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml"
)

type Operatingmode int

const (
	Simple Operatingmode = iota
	Voting
	UserVoting
	UserCoins
)

type CoinConfig struct {
	InitialCoins int
	PerVoteCoins int
	PerAddCoins  int
	MaximumCoins int
	RegenTime    int
}

type UserConfig struct {
	UserConfigDir          string
	AllowUserRegistration  bool
	ActivateUsersByDefault bool
}

type Config struct {
	Port       int
	FileDir    string
	YTApiKey   string
	Mode       Operatingmode
	Secret     string
	CoinConfig CoinConfig
	UserConfig UserConfig
}

func LoadConfig(filepath string) (Config, error) {

	res := Config{
		Port:    8081,
		FileDir: "/tmp/rave2gether/music",
		Mode:    Simple,
		CoinConfig: CoinConfig{
			InitialCoins: 10,
			PerVoteCoins: 1,
			PerAddCoins:  2,
			MaximumCoins: 100,
			RegenTime:    60,
		},
		UserConfig: UserConfig{
			UserConfigDir:          "/tmp/rave2gether/users",
			AllowUserRegistration:  true,
			ActivateUsersByDefault: true,
		},
	}
	file, err := os.Open(filepath)
	if err != nil {
		return res, errors.New("Error opening file: " + err.Error())
	}
	defer file.Close()
	decoder := toml.NewDecoder(file)
	err = decoder.Decode(&res)
	if err != nil {
		return res, errors.New("Error decoding file: " + err.Error())
	}
	if res.Secret == "" {
		return res, errors.New("secret is empty")
	}
	return res, nil
}

func (c *Config) SaveConfig(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New("Error opening file: " + err.Error())
	}
	defer file.Close()
	encoder := toml.NewEncoder(file)
	err = encoder.Encode(c)
	if err != nil {
		return errors.New("Error encoding file: " + err.Error())
	}
	return nil
}
