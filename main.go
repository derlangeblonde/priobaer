package main

import (
	"context"
	"os"

	"github.com/jonboulle/clockwork"
	"softbaer.dev/ass/cmd"
)

func main() {
	ctx := context.Background()
	err := cmd.Run(ctx, os.Getenv, clockwork.NewRealClock())

	if err != nil {
		panic(err)
	}
}
