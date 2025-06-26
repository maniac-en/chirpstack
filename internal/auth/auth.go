// Package auth provides auth mechanism for the app, like generating/checking
// password hashes.
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	ErrPasswordTooLong          string = "password too long"
	ErrIncorrectEmailOrPassword string = "incorrect email or password"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		if err.Error() == bcrypt.ErrPasswordTooLong.Error() {
			return "", fmt.Errorf(ErrPasswordTooLong)
		} else {
			return "", err
		}
	}
	return string(hash), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err.Error() == bcrypt.ErrMismatchedHashAndPassword.Error() {
			return fmt.Errorf(ErrIncorrectEmailOrPassword)
		} else {
			return err
		}
	}
	return nil
}
