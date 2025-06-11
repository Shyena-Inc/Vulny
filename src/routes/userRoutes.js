// src/routes/userRoutes.js
const express = require("express");
const router = express.Router();
const {
  registerUser,
  loginUser,
  getCurrentUser,
} = require("../controllers/userController");
const { authenticateJWT } = require("../middlewares/authMiddleware");

// Register new user
router.post("/register", registerUser);

// Login user
router.post("/login", loginUser);

// Get current user info
router.get("/me", authenticateJWT, getCurrentUser);

module.exports = router;
