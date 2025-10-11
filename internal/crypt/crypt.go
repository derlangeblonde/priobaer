package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log/slog"
)

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
