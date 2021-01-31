package rediskey

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashGuildID(guildID string) string {
	return genericHash(guildID)
}

func genericHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
