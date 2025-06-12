// middlewares/auth.go
package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Shyena-Inc/Vulny/models"
)

type contextKey string

const ContextUserKey = contextKey("user")

// AuthenticateJWT validates JWT token and attaches user info to context
func AuthenticateJWT(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims := &models.JWTClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			userID, err := primitive.ObjectIDFromHex(claims.ID)
			if err != nil {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}
			user := &models.User{
				ID:   userID,
				Role: claims.Role,
			}

			ctx := context.WithValue(r.Context(), ContextUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// AuthorizeRoles allows only given roles to access route
func AuthorizeRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]struct{})
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			userRaw := r.Context().Value(ContextUserKey)
			if userRaw == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user := userRaw.(*models.User)
			if _, ok := roleSet[user.Role]; !ok {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// FromRequest returns user info extracted by authentication middleware from request context
func UserFromRequest(r *http.Request) *models.User {
	userRaw := r.Context().Value(ContextUserKey)
	if userRaw == nil {
		return nil
	}
	user, ok := userRaw.(*models.User)
	if !ok {
		return nil
	}
	return user
}