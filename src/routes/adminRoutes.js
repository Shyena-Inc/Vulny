// src/routes/adminRoutes.js
const express = require("express");
const router = express.Router();
const {
  authenticateJWT,
  authorizeRole,
} = require("../middlewares/authMiddleware");
const {
  getAllUsers,
  getScanStats,
  getPlugins,
  addPlugin,
  updatePluginStatus,
  deletePlugin,
} = require("../controllers/adminController");

router.use(authenticateJWT, authorizeRole("admin"));

// Get all users
router.get("/users", getAllUsers);

// Get scan stats
router.get("/scan-stats", getScanStats);

// Manage scan plugins
router.get("/plugins", getPlugins);
router.post("/plugins", addPlugin);
router.patch("/plugins/:id/status", updatePluginStatus);
router.delete("/plugins/:id", deletePlugin);

module.exports = router;
