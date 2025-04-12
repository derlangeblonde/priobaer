package main 

import (
	"context"
	"os"

	"github.com/jonboulle/clockwork"
	"softbaer.dev/ass/internal/app/server"
)

func main() {
	ctx := context.Background()
	err := server.Run(ctx, os.Getenv, clockwork.NewRealClock())

	if err != nil {
		panic(err)
	}
}
