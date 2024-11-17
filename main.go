package main

import (
	"context"
	"os"

	"softbaer.dev/ass/cmd"
)

func main() {
	ctx := context.Background()
	err := cmd.Run(ctx, os.Getenv)

	if err != nil {
		panic(err)
	}
}
