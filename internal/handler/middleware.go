// internal/handler/middleware.go
package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go_marketplace_v1/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	userIDKey   ctxKey = "userID"
	userRoleKey ctxKey = "userRole"
)

// cGwC3QHEwc5Y
// AuthMiddleware проверяет JWT и кладёт userID из sub в контекст запроса.
func AuthMiddleware(secret []byte) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			log.Printf("m-ware auth: header=%q", auth)
			if !strings.HasPrefix(auth, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "требуется токен")
				return
			}
			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			claims := service.CustomClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				writeError(w, http.StatusUnauthorized, "недействительный токен")
				return
			}

			id, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "некорректный токен")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, id)
			ctx = context.WithValue(ctx, userRoleKey, claims.Role)
			next(w, r.WithContext(ctx))
		}
	}
}

// UserIDFromCtx достаёт id пользователя, положенный AuthMiddleware.
func UserIDFromCtx(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
func UserRoleFromCtx(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleKey).(string)
	return role, ok
}

// проверить роль  "admin"
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role, ok := UserRoleFromCtx(r.Context())
		if !ok || role != "admin" {
			writeError(w, http.StatusForbidden, "доступ запрещен: требуются права администратора")
			return
		}
		// Если роль "admin", передаем управление следующему хэндлеру
		next(w, r)
	}
}
