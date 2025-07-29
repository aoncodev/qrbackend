package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware validates token and sets user info in context
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing or invalid token"})
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        secret := []byte(os.Getenv("JWT_SECRET"))

        token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
            if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return secret, nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
            return
        }

        claims := token.Claims.(jwt.MapClaims)
        c.Set("userID", uint(claims["sub"].(float64)))
        c.Set("userRole", claims["role"].(string))
        c.Next()
    }
}

// RequireRole restricts access to a specific role
func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("userRole")
        if userRole != role {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Access denied"})
            return
        }
        c.Next()
    }
}
