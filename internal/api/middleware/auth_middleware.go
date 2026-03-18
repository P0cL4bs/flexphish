package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"flexphish/internal/api/handlers"
	"flexphish/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(jwtService *auth.JWTService) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handlers.JSONResponse(w, http.StatusUnauthorized, map[string]string{
					"error": "missing authorization token",
				})
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwtService.ValidateToken(tokenStr)
			if err != nil {
				handlers.JSONResponse(w, http.StatusUnauthorized, map[string]string{
					"error": "invalid token",
				})
				return
			}

			userIDUint64, err := strconv.ParseUint(claims.UserID, 10, 64)
			if err != nil {
				http.Error(w, "invalid user id", http.StatusUnauthorized)
				handlers.JSONResponse(w, http.StatusUnauthorized, map[string]string{
					"error": "invalid user data",
				})
				return
			}
			ctx := context.WithValue(r.Context(), "userID", int64(userIDUint64))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
