package crypt

import (
	"encoding/base64"
	"log/slog"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const userSecretKey = "secret"

func GetSecret(ctx *gin.Context) Secret {
	session := sessions.Default(ctx)
	secretInBase64 := session.Get(userSecretKey).(string)
	secret, err := base64.StdEncoding.DecodeString(secretInBase64)

	if err != nil {
		slog.Error("Error decoding secret", "err", err)
		panic(err)
	}

	return secret
}

func SetNewSecret(ctx *gin.Context) {
	session := sessions.Default(ctx)
	secret := generateSecret()
	secretInBase64 := base64.StdEncoding.EncodeToString(secret)
	session.Set(userSecretKey, secretInBase64)
}
