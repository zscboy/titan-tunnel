package middleware

import (
	"context"
	"fmt"
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
		ipStr := m.getClientIP(r)
		ip := net.ParseIP(ipStr)
		if ip.IsLoopback() || ip.IsPrivate() {
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

			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
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
			http.Error(w, fmt.Sprintf("Invalid token error:%v, token:'%s',", err, tokenStr), http.StatusUnauthorized)
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

func (m *JwtMiddleware) getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	return ip
}
