package config

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Port     int
	MusicDir string
	YTApiKey string
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
