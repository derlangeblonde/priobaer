package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log/slog"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Secret []byte

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

func Encrypt(plaintext string, secret Secret) (string, error) {
	logger := slog.With("Func", "Encrypt")
	block, err := aes.NewCipher(secret)
	if err != nil {
		logger.Error("Error creating AES cipher", "err", err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("Error creating gcm block", "err", err)
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Error("error generating nonce", "err", err)
		return "", err
	}

	cipherbytes := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(append(nonce, cipherbytes...)), nil
}

func Decrypt(ciphertext string, secret Secret) (string, error) {
	logger := slog.With("Func", "Decrypt")
	cipherbytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		logger.Error("Error decoding from base64", "err", err)
		return "", err
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		logger.Error("Error creating cipher", "err", err)
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("Error creating gcm block", "err", err)
		return "", err
	}

	nonce, cipherbytes := cipherbytes[:aesgcm.NonceSize()], cipherbytes[aesgcm.NonceSize():]

	plaintext, err := aesgcm.Open(nil, nonce, cipherbytes, nil)
	if err != nil {
		logger.Error("Error in final decrypt", "err", err)
		return "", err
	}

	return string(plaintext), nil
}

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
