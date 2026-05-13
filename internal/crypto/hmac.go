package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"os"
)

var hmacSecret = []byte(os.Getenv("HMAC_SECRET"))

// Генерация HMAC подписи
func GenerateHMAC(data string) string {
	h := hmac.New(sha256.New, hmacSecret)
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Проверка HMAC подписи
func VerifyHMAC(data, signature string) bool {
	expected := GenerateHMAC(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}
