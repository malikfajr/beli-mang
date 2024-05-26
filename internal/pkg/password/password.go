package password

import (
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

var (
	salt int
)

func init() {
	saltStr := os.Getenv("BCRYPT_SALT")
	n, err := strconv.Atoi(saltStr)
	if err != nil {
		salt = 8
	} else {
		salt = n
	}
}

func Hash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), salt)
	if err != nil {
		panic(err)
	}

	return string(hash)
}

func Compare(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}

	return true
}
