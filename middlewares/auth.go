package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Shyena-Inc/Vulny/models"
)

type contextKey string

const (
	ContextUserKey = contextKey("user")
	tokenIssuer    = "vulny-api" // Must match controllers/user.go
)

// AuthenticateJWT validates JWT token and attaches user info to context
func AuthenticateJWT(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			log.Printf("Auth Header: %s", authHeader)
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				sendError(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims := &models.JWTClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return secret, nil
			})

			if err != nil {
				log.Printf("JWT Parse Error: %v", err)
				if errors.Is(err, jwt.ErrTokenExpired) {
					sendError(w, "Token expired", http.StatusUnauthorized)
					return
				}
				sendError(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				log.Printf("Invalid token: token is not valid")
				sendError(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if claims.Issuer != tokenIssuer {
				log.Printf("Invalid token: issuer mismatch, got %s, expected %s", claims.Issuer, tokenIssuer)
				sendError(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			userID, err := primitive.ObjectIDFromHex(claims.ID)
			if err != nil {
				log.Printf("Invalid user ID in token: %v", err)
				sendError(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}

			user := &models.User{
				ID:   userID,
				Role: claims.Role,
			}

			ctx := context.WithValue(r.Context(), ContextUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthorizeRoles allows only given roles to access route
func AuthorizeRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]struct{})
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRaw := r.Context().Value(ContextUserKey)
			if userRaw == nil {
				sendError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user, ok := userRaw.(*models.User)
			if !ok {
				log.Printf("Invalid user data in context")
				sendError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if _, ok := roleSet[user.Role]; !ok {
				log.Printf("Forbidden: user role %s not allowed", user.Role)
				sendError(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserFromRequest returns user info extracted by authentication middleware from request context
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

// sendError sends a standardized error response
func sendError(w http.ResponseWriter, message string, status int) {
	resp := map[string]string{"error": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}