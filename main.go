package main

import (
	"os"

	"softbaer.dev/ass/cmd"
)

func main() {
	err := cmd.Run(os.Getenv)

	if err != nil {
		panic(err)
	}
}
