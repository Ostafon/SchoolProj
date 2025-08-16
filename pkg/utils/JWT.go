package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

func SignToken(userId int, username, role string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpires := os.Getenv("JWT_EXPIRES_IN")

	claims := jwt.MapClaims{
		"userId":   userId,
		"username": username,
		"role":     role,
	}
	if jwtExpires != "" {
		duration, err := time.ParseDuration(jwtExpires)
		if err != nil {
			return "", ErrorHandler(err, "Internal error,expired")
		}
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration))
	} else {
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", ErrorHandler(err, "Internal error,token")
	}
	return signedToken, nil
}
