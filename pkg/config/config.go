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
	var res Config
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
