package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	DbRootDir     string
	SessionMaxAge time.Duration
	Port          int
}

func ParseConfig(getenv func(string) string) (Config, error) {
	config := Config{}

	dbRootDir := getenv("PRIOBAER_DB_ROOT_DIR")

	if dbRootDir == "" {
		return config, errors.New("PRIOBAER_DB_ROOT_DIR not set")
	}

	config.DbRootDir = dbRootDir

	sessionMaxAge, err := GetInt(getenv, "PRIOBAER_SESSION_MAX_AGE")

	if err != nil {
		return config, err
	}

	config.SessionMaxAge = time.Second * time.Duration(sessionMaxAge)

	port, err := GetInt(getenv, "PRIOBAER_PORT")

	if err != nil {
		return config, err
	}

	config.Port = port

	return config, nil
}

func GetInt(getenv func(string) string, key string) (int, error) {
	sessionMaxAgeString := getenv(key)

	if sessionMaxAgeString == "" {
		return 0, fmt.Errorf("%s not set", key)
	}

	return strconv.Atoi(sessionMaxAgeString)
}
