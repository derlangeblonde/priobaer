package cmd

import (
	"errors"
	"strconv"
	"time"
)

type Config struct {
	DbRootDir string
	SessionMaxAge time.Duration 
}

func ParseConfig(getenv func(string) string) (Config, error) {
	config := Config{}

	dbRootDir := getenv("DB_ROOT_DIR")

	if dbRootDir == "" {
		return config, errors.New("DB_ROOT_DIR not set")
	}

	config.DbRootDir = dbRootDir

	sessionMaxAgeString := getenv("SESSION_MAX_AGE_MILLI_SECONDS")

	if sessionMaxAgeString == "" {
		return config, errors.New("SESSION_MAX_AGE_MILLI_SECONDS not set")
	}

	sessionMaxAgeMilliSeconds, err := strconv.Atoi(sessionMaxAgeString)

	if err != nil {
		return config, err
	}

	config.SessionMaxAge = time.Millisecond * time.Duration(sessionMaxAgeMilliSeconds) 

	return config, nil
}
