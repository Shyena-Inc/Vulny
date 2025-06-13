// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
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
	"github.com/Shyena-Inc/Vulny/services"
	"github.com/joho/godotenv"
)

var (
	mongoClient   *mongo.Client
	mongoDatabase *mongo.Database
	redisClient   *redis.Client
	jobConnection rmq.Connection
	jobQueue      rmq.Queue
	ctx           = context.Background()
	jwtSecret     []byte
)

// Start worker to process scan jobs asynchronously
type scanWorker struct{}

func (w *scanWorker) Consume(delivery rmq.Delivery) {
    log.Println("Processing scan job:", delivery.Payload())
    // Implement scan job processing logic here
    time.Sleep(2 * time.Second) // Simulate some work
    log.Println("Scan job completed:", delivery.Payload())
    delivery.Ack()
}

func main() {
	_ = godotenv.Load()
	jwtSecret = []byte(services.JWTSecret)


	// MongoDB setup
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB Connection Error:", err)
	}
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB Ping Error:", err)
	}
	mongoDatabase = mongoClient.Database("vulny")
	controllers.SetUserCollection(mongoDatabase)
	log.Println("MongoDB connected.")

	// Redis setup
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Redis Connection Error:", err)
	}
	log.Println("Redis connected.")

	// RMQ (Redis message queue) setup
	jobConnection, err = rmq.OpenConnection("vulny_rmq", "tcp", redisAddr, 1, nil)

	if err != nil {
		log.Fatal("RMQ Connection Error:", err)
	}
	jobQueue, err = jobConnection.OpenQueue("scanQueue")
	if err != nil {
    	log.Fatal("RMQ OpenQueue Error:", err)
	}

	jobQueue.StartConsuming(5, time.Second)

	_, err = jobQueue.AddConsumer("scanWorker", &scanWorker{})
	if err != nil {
		log.Fatal("RMQ AddConsumer Error:", err)
	}
	log.Println("Scan worker started.")
	

	// Setup router and middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Change to specific origins in production
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}).Handler)

	// Rate Limiting middleware (simple IP-based)
	r.Use(middlewares.RateLimitMiddleware(100, 15*time.Minute))

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
		// r.Delete("/{id}", controllers.CancelScan) // CancelScan handler not implemented
	})

	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middlewares.AuthenticateJWT(jwtSecret))
		r.Use(middlewares.AuthorizeRoles("admin"))
		//r.Get("/users", controllers.GetAllUsers)
		//r.Get("/scan-stats", controllers.GetScanStats)
		//r.Get("/plugins", controllers.GetPlugins)
		//r.Post("/plugins", controllers.AddPlugin)
		//r.Patch("/plugins/{id}/status", controllers.UpdatePluginStatus)
		//r.Delete("/plugins/{id}", controllers.DeletePlugin)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Welcome to Vulny - Web Vulnerability Scanner API"}`))
	})

	port := getEnv("PORT", "3000")
	log.Println("Starting Vulny API server on port", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}

func getEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}