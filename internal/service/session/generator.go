package session

import (
	"crypto/rand"
	"encoding/hex"
)

// CryptoSIDGenerator generates cryptographically secure SIDs.
type CryptoSIDGenerator struct{}

// Generate creates 128-bit SID encoded as hex string.
func (CryptoSIDGenerator) Generate() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
