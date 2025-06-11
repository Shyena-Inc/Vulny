// src/controllers/scanController.js
const Scan = require("../models/Scan");
const { scanQueue } = require("../app"); // The BullMQ queue from app.js
const mongoose = require("mongoose");

// Input validation helper (basic)
const isValidUrl = (url) => {
  try {
    const u = new URL(url);
    return ["http:", "https:"].includes(u.protocol);
  } catch {
    return false;
  }
};

// Start Scan Controller
const startScan = async (req, res, next) => {
  try {
    const userId = req.user._id;
    const { targetURL, config } = req.body;

    if (!targetURL || !isValidUrl(targetURL)) {
      return res.status(400).json({ error: "Invalid or missing targetURL" });
    }

    // Additional input validation on config (depth, headers)
    const validatedConfig = {
      depth:
        config && typeof config.depth === "number" && config.depth > 0
          ? config.depth
          : 1,
      headers:
        config && typeof config.headers === "object" ? config.headers : {},
    };

    // Create Scan doc with status 'pending'
    const scanDoc = new Scan({
      user: userId,
      targetURL,
      config: validatedConfig,
      status: "pending",
      vulnerabilities: [],
    });
    await scanDoc.save();

    // Add scan job to BullMQ queue
    await scanQueue.add("vulny-scan", { scanId: scanDoc._id.toString() });

    return res.status(202).json({
      message: "Scan job accepted",
      scanId: scanDoc._id,
      status: scanDoc.status,
      targetURL: scanDoc.targetURL,
    });
  } catch (err) {
    next(err);
  }
};

// Get all scans for logged-in user
const getAllScans = async (req, res, next) => {
  try {
    const userId = req.user._id;

    // Admin can get all scans, users only their own
    if (req.user.role === "admin") {
      const scans = await Scan.find().sort({ createdAt: -1 }).limit(100);
      return res.json({ scans });
    }

    const scans = await Scan.find({ user: userId })
      .sort({ createdAt: -1 })
      .limit(100);
    return res.json({ scans });
  } catch (err) {
    next(err);
  }
};

// Get scan by ID - only owner or admin allowed
const getScanById = async (req, res, next) => {
  try {
    const scanId = req.params.id;

    if (!mongoose.Types.ObjectId.isValid(scanId)) {
      return res.status(400).json({ error: "Invalid scan ID" });
    }

    const scan = await Scan.findById(scanId);
    if (!scan) return res.status(404).json({ error: "Scan not found" });

    if (req.user.role !== "admin" && !scan.user.equals(req.user._id)) {
      return res.status(403).json({ error: "Access denied" });
    }

    return res.json({ scan });
  } catch (err) {
    next(err);
  }
};

// Cancel scan by ID - only pending or running scans can be cancelled by owner or admin
const cancelScan = async (req, res, next) => {
  try {
    const scanId = req.params.id;

    if (!mongoose.Types.ObjectId.isValid(scanId)) {
      return res.status(400).json({ error: "Invalid scan ID" });
    }

    const scan = await Scan.findById(scanId);
    if (!scan) return res.status(404).json({ error: "Scan not found" });

    if (req.user.role !== "admin" && !scan.user.equals(req.user._id)) {
      return res.status(403).json({ error: "Access denied" });
    }

    // Only pending or running scans can be cancelled
    if (!["pending", "running"].includes(scan.status)) {
      return res
        .status(400)
        .json({ error: "Only pending or running scans can be cancelled" });
    }

    // Update scan status to cancelled
    scan.status = "cancelled";
    scan.completedAt = new Date();
    await scan.save();

    // Remove related job from queue if exists
    const jobs = await scanQueue.getJobs(["waiting", "active", "delayed"]);
    for (const job of jobs) {
      if (job.data.scanId === scan._id.toString()) {
        await job.remove();
        break;
      }
    }

    return res.json({ message: "Scan cancelled", scanId: scan._id });
  } catch (err) {
    next(err);
  }
};

module.exports = {
  startScan,
  getAllScans,
  getScanById,
  cancelScan,
};
