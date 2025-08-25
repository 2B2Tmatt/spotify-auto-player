package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateRandomString(length int) (string, error) {
	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	randString := base64.RawURLEncoding.EncodeToString(randBytes)
	return randString, nil
}

func CodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func NewPKCE() (string, string, error) {
	v, err := GenerateRandomString(64)
	if err != nil {
		return "", "", err
	}
	return v, CodeChallenge(v), nil
}
