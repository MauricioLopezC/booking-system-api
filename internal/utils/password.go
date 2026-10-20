package utils

import "golang.org/x/crypto/bcrypt"

// PasswordCost bcrypt cost: ~250ms per hash. Hashes created with another cost still verify.
const PasswordCost = 12

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
