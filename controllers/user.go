// controllers/user.go
package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Shyena-Inc/Vulny/models"
	"github.com/Shyena-Inc/Vulny/services"
)

var (
	userCollection *mongo.Collection
	jwtSecret      []byte
)

func init() {
	jwtSecret = []byte(services.JWTSecret)
}

// SetUserCollection sets the user collection
func SetUserCollection(db *mongo.Database) {
	userCollection = db.Collection("users")
}

// RegisterUser handler for POST /api/users/register
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Password = strings.TrimSpace(body.Password)

	if len(body.Username) < 3 || len(body.Password) < 8 || !services.IsValidEmail(body.Email) {
		http.Error(w, "Invalid input data (username, email or password)", http.StatusBadRequest)
		return
	}

	// Check user existence
	count, err := userCollection.CountDocuments(context.Background(), bson.M{"$or": []bson.M{
		{"email": body.Email},
		{"username": body.Username},
	}})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "User with username or email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashPw, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Username:  body.Username,
		Email:     body.Email,
		Password:  string(hashPw),
		Role:      "user",
		CreatedAt: time.Now(),
	}

	res, err := userCollection.InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	userID := res.InsertedID.(primitive.ObjectID)

	// Generate JWT token
	token, err := generateJWT(userID.Hex(), user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":       userID.Hex(),
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"token": token,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// LoginUser handler for POST /api/users/login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Password = strings.TrimSpace(body.Password)
	if body.Email == "" || body.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var user models.User
	err := userCollection.FindOne(context.Background(), bson.M{"email": body.Email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.Hex(), user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"token": token,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Internal token generation helper
func generateJWT(userID, role string) (string, error) {
	claims := &models.JWTClaims{  // <- pointer here
		ID:   userID,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
