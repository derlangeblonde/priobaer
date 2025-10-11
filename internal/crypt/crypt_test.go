package crypt

import (
	"testing"

	"github.com/matryer/is"
)

func TestEncryptDecryptRoundtripIsConsistent(t *testing.T) {
	is := is.New(t)

	secret := generateSecret()

	testcases := []string{
		"hallo",
		"mit ümläute",
		"malen nach zahlen 134",
	}

	for _, plaintext := range testcases {
		ciphertext, err := Encrypt(plaintext, secret)
		is.NoErr(err) // want encryption to be successful
		plaintext2, err := Decrypt(ciphertext, secret)
		is.NoErr(err) // want decryption to be successful

		is.Equal(plaintext, plaintext2) // want plaintext to stay the same through a en- decrypt roundtrip
	}
}

func BenchmarkCryptRoundtrip(t *testing.B) {
	secret := generateSecret()
	plaintext := "einmittellangerstring"

	for t.Loop() {
		ciphertext, _ := Encrypt(plaintext, secret)
		Decrypt(ciphertext, secret)
	}

}
