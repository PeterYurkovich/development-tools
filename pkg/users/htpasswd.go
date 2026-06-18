package users

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func GenerateHtpasswdData(count int, password string) ([]byte, error) {
	hash, err := generateBcryptHash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bcrypt hash: %w", err)
	}

	var builder strings.Builder
	for index := 1; index <= count; index++ {
		username := fmt.Sprintf("user%d", index)
		builder.WriteString(fmt.Sprintf("%s:%s\n", username, hash))
	}

	return []byte(builder.String()), nil
}

func generateBcryptHash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
