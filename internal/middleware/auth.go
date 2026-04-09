package middleware

import (
	"net/http"
	"strings"

	"somewebproject/internal/auth"
	"somewebproject/internal/repository"
)

func NewAuthMiddleware(secret string, users repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := bearerToken(r.Header.Get("Authorization"))
			if tokenString == "" {
				writeAuthError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			claims, err := auth.ParseToken(tokenString, secret)
			if err != nil || claims.TokenType != auth.TokenTypeAccess {
				writeAuthError(w, http.StatusUnauthorized, "invalid access token")
				return
			}

			userID, err := auth.PrincipalIDFromClaims(claims)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			user, err := users.FindByID(r.Context(), userID)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "user not found")
				return
			}

			if user.IsBlocked {
				writeAuthError(w, http.StatusForbidden, "user is blocked")
				return
			}

			ctx := auth.WithPrincipal(r.Context(), auth.Principal{ID: user.ID, Email: user.Email, Role: user.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
