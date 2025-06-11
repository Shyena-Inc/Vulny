// src/controllers/userController.js
const jwt = require("jsonwebtoken");
const User = require("../models/User");

const JWT_SECRET = process.env.JWT_SECRET || "your_jwt_secret";
const JWT_EXPIRES_IN = "24h";

// Generate JWT token
const generateToken = (userId, role) =>
  jwt.sign({ id: userId, role: role }, JWT_SECRET, {
    expiresIn: JWT_EXPIRES_IN,
  });

// Register User Controller
const registerUser = async (req, res, next) => {
  try {
    const { username, email, password, role } = req.body;

    if (!username || !email || !password) {
      return res
        .status(400)
        .json({ error: "Username, email and password are required" });
    }

    // Prevent users from registering as admin directly
    if (role && role === "admin") {
      return res
        .status(403)
        .json({ error: "Cannot assign admin role during registration" });
    }

    const userExists = await User.findOne({ $or: [{ email }, { username }] });
    if (userExists)
      return res.status(409).json({ error: "Username or email already taken" });

    const user = new User({ username, email, password, role: "user" });
    await user.save();

    const token = generateToken(user._id, user.role);
    return res.status(201).json({
      message: "User registered successfully",
      user: {
        id: user._id,
        username: user.username,
        email: user.email,
        role: user.role,
      },
      token,
    });
  } catch (err) {
    next(err);
  }
};

// Login User Controller
const loginUser = async (req, res, next) => {
  try {
    const { email, password } = req.body;

    if (!email || !password)
      return res.status(400).json({ error: "Email and password are required" });

    const user = await User.findOne({ email });
    if (!user)
      return res.status(401).json({ error: "Invalid email or password" });

    const isMatch = await user.comparePassword(password);
    if (!isMatch)
      return res.status(401).json({ error: "Invalid email or password" });

    const token = generateToken(user._id, user.role);

    return res.json({
      message: "Login successful",
      user: {
        id: user._id,
        username: user.username,
        email: user.email,
        role: user.role,
      },
      token,
    });
  } catch (err) {
    next(err);
  }
};

// Get Current User Info Controller
const getCurrentUser = (req, res) => {
  if (!req.user) return res.status(401).json({ error: "Unauthorized" });
  return res.json({
    id: req.user._id,
    username: req.user.username,
    email: req.user.email,
    role: req.user.role,
  });
};

module.exports = {
  registerUser,
  loginUser,
  getCurrentUser,
};
