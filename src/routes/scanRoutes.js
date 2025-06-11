// src/routes/scanRoutes.js
const express = require("express");
const router = express.Router();
const {
  startScan,
  getAllScans,
  getScanById,
  cancelScan,
} = require("../controllers/scanController");

// Start scan - POST /api/scans
router.post("/", startScan);

// Get all scans (for logged-in user)
router.get("/", getAllScans);

// Get scan by ID
router.get("/:id", getScanById);

// Cancel scan by ID
router.delete("/:id", cancelScan);

module.exports = router;
