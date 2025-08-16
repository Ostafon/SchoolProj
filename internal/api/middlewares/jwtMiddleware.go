package middlewares

import (
	"WebProject/pkg/utils"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("Bearer")
		jwtSecret := os.Getenv("JWT_SECRET")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		parsedToken, err := jwt.Parse(token.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

		if parsedToken.Valid {
			log.Println("Token Valid")
		} else {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if ok {
			fmt.Println(claims["userId"], claims["exp"])
		} else {
			fmt.Println(err)
		}

		ctx := context.WithValue(r.Context(), utils.ContextKey("role"), claims["role"].(string))
		ctx.Value("role")
		ctx = context.WithValue(r.Context(), utils.ContextKey("expiresAt"), claims["exp"])
		ctx = context.WithValue(r.Context(), utils.ContextKey("username"), claims["username"])
		ctx = context.WithValue(r.Context(), utils.ContextKey("userId"), claims["userId"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})

}
