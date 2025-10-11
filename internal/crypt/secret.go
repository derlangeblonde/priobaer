package crypt

import (
	"crypto/rand"
	"log/slog"
)

type Secret []byte

func generateSecret() Secret {
	secret := make([]byte, 32)
	n, err := rand.Reader.Read(secret)

	if err != nil {
		slog.Error("Error generating secret", "err", err)
		panic(err)
	}

	if n != 32 {
		msg := "Could not generate secret in full length"
		slog.Error(msg, "n", n)
		panic(msg)
	}

	return secret
}
