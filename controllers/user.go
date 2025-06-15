package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Shyena-Inc/Vulny/models"
)

// Constants for input validation and JWT
const (
	minUsernameLength = 3
	minPasswordLength = 8
	maxUsernameLength = 50
	maxEmailLength    = 100
	bcryptCost        = bcrypt.DefaultCost
	tokenExpiration   = 24 * time.Hour
	tokenIssuer       = "vulny-api"
	minJWTSecretLength = 32 // Recommended for HS256
)

var (
	userCollection *mongo.Collection
	jwtSecret      []byte
)

// SetUserCollection sets the MongoDB user collection
func SetUserCollection(db *mongo.Database) {
	userCollection = db.Collection("users")
}

// SetJWTSecret sets the JWT secret for signing tokens
func SetJWTSecret(secret []byte) {
	if len(secret) == 0 {
		log.Println("Warning: JWT secret is empty")
	} else if len(secret) < minJWTSecretLength {
		log.Printf("Warning: JWT secret is too short (%d bytes, minimum %d bytes)", len(secret), minJWTSecretLength)
	} else {
		log.Printf("JWT secret set (%d bytes)", len(secret))
	}
	jwtSecret = secret
}

// userResponse defines the structure for user API responses
type userResponse struct {
	Message string      `json:"message"`
	User    userData    `json:"user"`
	Token   string      `json:"token"`
}

type userData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// reqBody is used for user registration and login requests
type reqBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterUser handles POST /api/users/register
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize and validate input
	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Password = strings.TrimSpace(body.Password)

	if err := validateRegisterInput(body); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check for existing user
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := userCollection.CountDocuments(ctx, bson.M{"$or": []bson.M{
		{"email": body.Email},
		{"username": body.Username},
	}})
	if err != nil {
		sendError(w, "Failed to check user existence", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		sendError(w, "Username or email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashPw, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcryptCost)
	if err != nil {
		sendError(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Username:  body.Username,
		Email:     body.Email,
		Password:  string(hashPw),
		Role:      "user",
		CreatedAt: time.Now(),
	}

	// Insert user
	res, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		sendError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	userID := res.InsertedID.(primitive.ObjectID)
	token, err := generateJWT(userID.Hex(), user.Role)
	if err != nil {
		log.Printf("RegisterUser: Failed to generate token for user %s: %v", userID.Hex(), err)
		sendError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	sendUserResponse(w, "User registered successfully", user, token, http.StatusCreated)
}

// LoginUser handles POST /api/users/login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize and validate input
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Password = strings.TrimSpace(body.Password)
	if body.Email == "" || body.Password == "" {
		sendError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Find user
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": body.Email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		sendError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		sendError(w, "Failed to query user", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		sendError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := generateJWT(user.ID.Hex(), user.Role)
	if err != nil {
		log.Printf("LoginUser: Failed to generate token for user %s: %v", user.ID.Hex(), err)
		sendError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	sendUserResponse(w, "Login successful", user, token, http.StatusOK)
}

// validateRegisterInput checks registration input for validity
func validateRegisterInput(body reqBody) error {
	if len(body.Username) < minUsernameLength || len(body.Username) > maxUsernameLength {
		return fmt.Errorf("username must be between %d and %d characters", minUsernameLength, maxUsernameLength)
	}
	if len(body.Password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}
	if len(body.Email) > maxEmailLength || !isValidEmail(body.Email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// isValidEmail provides a simple email format validation
func isValidEmail(email string) bool {
	// Basic check for presence of "@" and "."
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")
	return at > 0 && dot > at+1 && dot < len(email)-1
}

// generateJWT creates a JWT token for a user
func generateJWT(userID, role string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", fmt.Errorf("JWT secret is not set")
	}
	if len(jwtSecret) < minJWTSecretLength {
		return "", fmt.Errorf("JWT secret is too short (%d bytes, minimum %d bytes)", len(jwtSecret), minJWTSecretLength)
	}
	claims := models.JWTClaims{
		ID:   userID,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Failed to sign JWT for user %s: %v", userID, err)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	log.Printf("Generated JWT for user %s with role %s", userID, role)
	return signedToken, nil
}

// sendUserResponse sends a standardized user response
func sendUserResponse(w http.ResponseWriter, message string, user models.User, token string, status int) {
	resp := userResponse{
		Message: message,
		User: userData{
			ID:       user.ID.Hex(),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
		Token: token,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// sendError sends an error response with a consistent format
func sendError(w http.ResponseWriter, message string, status int) {
	resp := map[string]string{"error": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}