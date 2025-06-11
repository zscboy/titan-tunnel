package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type JwtMiddleware struct {
	SecretKey string
}

func NewJwtMiddleware(secret string) *JwtMiddleware {
	return &JwtMiddleware{
		SecretKey: secret,
	}
}

func (m *JwtMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isLocalIP(r) {
			next(w, r)
			return
		}

		tokenStr := r.URL.Query().Get("token")
		if len(tokenStr) == 0 {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) == 0 {
				http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				http.Error(w, "Invalid Authorization Header", http.StatusUnauthorized)
				return
			}
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.SecretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Optionally: extract claims and put into context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), "jwtClaims", claims)
			r = r.WithContext(ctx)
		}

		next(w, r)
	}
}

func isLocalIP(r *http.Request) bool {
	localIP := "127.0.0.1"
	ip := r.Header.Get("X-Forwarded-For")
	if len(ip) != 0 {
		return ip == localIP
	}

	ip = r.Header.Get("X-Real-IP")
	if len(ip) != 0 {
		return ip == localIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if ip == "127.0.0.1" || ip == "::1" {
			return true
		}
	}

	return false

}
