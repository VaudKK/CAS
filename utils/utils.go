package utils

import (
	"crypto/sha256"
	"fmt"
)

func HashFile(data []byte) string {
	hashObject := sha256.New()
	hashObject.Write(data)
	return fmt.Sprintf("%x", hashObject.Sum(nil))
}