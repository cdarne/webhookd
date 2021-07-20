package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// ValidSignature reports whether the data matches the provided base64 hash signature
func ValidSignature(data []byte, key, base64Hash string) bool {
	hash, err := base64.URLEncoding.DecodeString(base64Hash)
	if err != nil {
		return false
	}

	return validMAC(data, hash, []byte(key))
}

func validMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
