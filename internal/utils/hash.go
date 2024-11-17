package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashContent(content string) string {
	hasher := sha256.New()
	hasher.Write([]byte(content))
	return hex.EncodeToString(hasher.Sum(nil))
}
