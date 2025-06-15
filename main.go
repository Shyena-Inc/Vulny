package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Shyena-Inc/Vulny/controllers"
	"github.com/Shyena-Inc/Vulny/middlewares"
	"github.com/joho/godotenv"
)

const minJWTSecretLength = 32 // Matches controllers/user.go

var (
	mongoClient   *mongo.Client
	mongoDatabase *mongo.Database
	redisClient   *redis.Client
	jobConnection rmq.Connection
	jobQueue      rmq.Queue
	ctx           = context.Background()
	jwtSecret     []byte
)

// scanWorker processes scan jobs asynchronously
type scanWorker struct{}

func (w *scanWorker) Consume(delivery rmq.Delivery) {
	log.Println("Processing scan job:", delivery.Payload())
	// TODO: Implement actual scan job processing logic (e.g., call port scanner or subdomain enumeration)
	time.Sleep(2 * time.Second) // Simulate work (replace with real logic)
	log.Println("Scan job completed:", delivery.Payload())
	if err := delivery.Ack(); err != nil {
		log.Println("Failed to acknowledge scan job:", err)
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Validate JWT secret
	jwtSecret = []byte(getEnv("JWT_SECRET", ""))
	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET is not set")
	}
	if len(jwtSecret) < minJWTSecretLength {
		log.Fatalf("JWT_SECRET is too short (%d bytes, minimum %d bytes)", len(jwtSecret), minJWTSecretLength)
	}
	log.Printf("JWT_SECRET set (%d bytes)", len(jwtSecret))
	controllers.SetJWTSecret(jwtSecret)

	// MongoDB setup
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB Connection Error:", err)
	}
	if err = mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB Ping Error:", err)
	}
	mongoDatabase = mongoClient.Database("vulny")
	controllers.SetUserCollection(mongoDatabase)
	log.Println("MongoDB connected")
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Println("MongoDB Disconnect Error:", err)
		}
	}()

	// Redis setup
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if _, err = redisClient.Ping(ctx).Result(); err != nil {
		log.Fatal("Redis Connection Error:", err)
	}
	log.Println("Redis connected")
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Println("Redis Close Error:", err)
		}
	}()

	// RMQ (Redis message queue) setup
	jobConnection, err = rmq.OpenConnection("vulny_rmq", "tcp", redisAddr, 1, nil)
	if err != nil {
		log.Fatal("RMQ Connection Error:", err)
	}
	defer func() {
		if err := jobConnection.StopAllConsuming(); err != nil {
			log.Println("RMQ StopAllConsuming Error:", err)
		}
	}()

	jobQueue, err = jobConnection.OpenQueue("scanQueue")
	if err != nil {
		log.Fatal("RMQ OpenQueue Error:", err)
	}

	if err = jobQueue.StartConsuming(5, time.Second); err != nil {
		log.Fatal("RMQ StartConsuming Error:", err)
	}

	if _, err = jobQueue.AddConsumer("scanWorker", &scanWorker{}); err != nil {
		log.Fatal("RMQ AddConsumer Error:", err)
	}
	log.Println("Scan worker started")

	// Setup router and middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS configuration
	allowedOrigins := strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ",")
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}).Handler)

	// Rate Limiting middleware (configurable)
	rateLimitRequests := getEnvInt("RATE_LIMIT_REQUESTS", 100)
	rateLimitDuration := time.Duration(getEnvInt("RATE_LIMIT_DURATION_MINUTES", 15)) * time.Minute
	r.Use(middlewares.RateLimitMiddleware(rateLimitRequests, rateLimitDuration))

	// Routes
	r.Route("/api/users", func(r chi.Router) {
		r.Post("/register", controllers.RegisterUser)
		r.Post("/login", controllers.LoginUser)
	})

	r.Route("/api/scans", func(r chi.Router) {
		r.Use(middlewares.AuthenticateJWT(jwtSecret))
		r.Post("/", controllers.StartScan)
		r.Get("/", controllers.GetAllScans)
		r.Get("/{id}", controllers.GetScanByID)
	})

	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middlewares.AuthenticateJWT(jwtSecret))
		r.Use(middlewares.AuthorizeRoles("admin"))
		// TODO: Implement admin routes as needed
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Welcome to Vulny - Web Vulnerability Scanner API"}`))
	})

	// Setup server with graceful shutdown
	port := getEnv("PORT", "3000")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Starting Vulny API server on port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server Error:", err)
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server...")

	// Give the server 10 seconds to finish in-progress requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown Error:", err)
	}
	log.Println("Server gracefully stopped")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}