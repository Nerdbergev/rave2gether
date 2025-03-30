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

type Config struct {
	Port       int
	FileDir    string
	YTApiKey   string
	Mode       Operatingmode
	Secret     string
	CoinConfig CoinConfig
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
