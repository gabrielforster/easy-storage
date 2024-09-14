package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	// "fmt"
	// "strings"
)

func Sign(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

