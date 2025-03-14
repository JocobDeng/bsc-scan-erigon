package util

import (
	"crypto/sha256"
	"encoding/base64"
)

func EncodeContractCode(code []byte) string {
	if len(code) > 0 {
		hasher := sha256.New()
		hasher.Write(code)
		hash := hasher.Sum(nil)
		return base64.StdEncoding.EncodeToString(hash)
	}
	return ""
}
func EncodeContractStrCode(code string) string {
	if code != "" {
		hasher := sha256.New()
		hasher.Write([]byte(code))
		hash := hasher.Sum(nil)
		return base64.StdEncoding.EncodeToString(hash)
	}
	return ""
}
