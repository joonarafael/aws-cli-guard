package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
)

const tokenBytes = 32

var tokenRE = regexp.MustCompile(`^[0-9a-f]{64}$`)

func generateToken() (string, error) {
	buf := make([]byte, tokenBytes)

	_, err := rand.Read(buf)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: sessions: random token: %w", err)
	}

	return hex.EncodeToString(buf), nil
}

func validToken(token string) bool {
	return tokenRE.MatchString(token)
}

// ValidTokenPublic reports whether token matches the on-disk shape.
func ValidTokenPublic(token string) bool {
	return validToken(token)
}
