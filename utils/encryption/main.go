/*
Package encryption holds all logic and functions related
to hancle encryption, mostly used at the keys package.
*/

package encryption

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func GetSecretHash(username, secret string) string {
	data := username
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	hash := h.Sum(nil)
	secretHash := base64.StdEncoding.EncodeToString(hash)
	return secretHash
}

func GetSHA256(message []byte) []byte {
	w := sha256.New()
	w.Write(message)
	return w.Sum(nil)
}
