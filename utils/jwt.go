package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID uint, role string) (string, error) {
    secret := []byte(os.Getenv("JWT_SECRET"))

    claims := jwt.MapClaims{
        "sub":  userID,
        "role": role,
        "exp":  time.Now().Add(30 * time.Minute).Unix(),
        "iat":  time.Now().Unix(),
        "iss":  "lazzat-backend",
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}
