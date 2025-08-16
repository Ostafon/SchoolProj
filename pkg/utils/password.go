package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

func VerifyPassword(existPass, checkPass string) error {
	parts := strings.Split(existPass, ".")
	if len(parts) != 2 {
		err := errors.New("password format error")
		return ErrorHandler(err, "Invalid hash format")
	}
	saltBase64 := parts[0]
	hashBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return ErrorHandler(err, "Invalid salt format")
	}
	hashPass, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return ErrorHandler(err, "Invalid password format")
	}

	hash := argon2.IDKey([]byte(checkPass), salt, 1, 64*1024, 4, 32)
	fmt.Printf("hash: %x\n", hash)
	fmt.Printf("hashPass: %x", hashPass)

	if len(hash) != len(hashPass) {
		return ErrorHandler(err, "Invalid hash length")
	}

	if subtle.ConstantTimeCompare(hash, []byte(hashPass)) != 1 {
		return ErrorHandler(err, "Invalid password")
	}
	return nil
}

func PasswordHashing(password string) (error, string) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return ErrorHandler(err, "Error generating random salt"), ""
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)
	encodedPass := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	return err, encodedPass
}
