// src/app.js
require("dotenv").config();
const express = require("express");
const mongoose = require("mongoose");
const { Queue, Worker } = require("bullmq");
const morgan = require("morgan");
const rateLimit = require("express-rate-limit");
const helmet = require("helmet");
const cors = require("cors");
const { runScan } = require("./scan-engine/index");

const userRoutes = require("./routes/userRoutes");
const scanRoutes = require("./routes/scanRoutes");
const adminRoutes = require("./routes/adminRoutes");
const { errorHandler } = require("./middlewares/errorHandler");
const { authenticateJWT } = require("./middlewares/authMiddleware");

const app = express();

// Security Headers
app.use(helmet());

// CORS
app.use(cors());

// Logger
app.use(morgan("dev"));

// Body parsers
app.use(express.json());
app.use(express.urlencoded({ extended: false }));

// Rate Limiting
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100,
  message: "Too many requests from this IP, please try again later.",
});
app.use(limiter);

// index scan engine
// (removed duplicate scanWorker declaration)

// MongoDB Connection
const MONGODB_URI =
  process.env.MONGODB_URI || "mongodb://localhost:27017/vulny";
mongoose
  .connect(MONGODB_URI, {
    useNewUrlParser: true,
    useUnifiedTopology: true,
  })
  .then(() => console.log("MongoDB connected successfully"))
  .catch((err) => {
    console.error("MongoDB connection error:", err);
    process.exit(1);
  });

// Redis Connection config
const REDIS_HOST = process.env.REDIS_HOST || "127.0.0.1";
const REDIS_PORT = process.env.REDIS_PORT || 6379;
const redisConnection = {
  host: REDIS_HOST,
  port: REDIS_PORT,
};

// BullMQ Scan Queue
const scanQueue = new Queue("scanQueue", {
  connection: redisConnection,
});

// Worker processing scan jobs (basic stub, scan logic goes in services/scan-engine/)
const scanWorker = new Worker(
  "scanQueue",
  async (job) => {
    console.log(`Processing scan job id ${job.id}:`, job.data);
    // Here call the scan engine modules with job.data to perform scans

    // Simulate scan result for now
    return {
      status: "completed",
      vulnerabilities: [],
      scannedAt: new Date(),
    };
  },
  { connection: redisConnection }
);

// Events for worker
scanWorker.on("completed", (job, result) => {
  console.log(`Scan job ${job.id} finished:`, result);
});

scanWorker.on("failed", (job, err) => {
  console.error(`Scan job ${job.id} failed:`, err);
});

// Routes
app.use("/api/users", userRoutes);
app.use("/api/scans", authenticateJWT, scanRoutes);
app.use("/api/admin", authenticateJWT, adminRoutes);

// Root endpoint
app.get("/", (req, res) => {
  res.json({ message: "Welcome to Vulny - Web Vulnerability Scanner API" });
});

// 404 Handler
app.use((req, res, next) => {
  res.status(404).json({ error: "Route not found" });
});

// Centralized Error Handler
app.use(errorHandler);

// Start Server
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Vulny API server running on port ${PORT}`);
});

module.exports = { app, scanQueue };
