// src/models/Scan.js
const mongoose = require("mongoose");

const VulnerabilitySchema = new mongoose.Schema({
  type: { type: String, required: true },
  parameter: { type: String, required: true },
  severity: {
    type: String,
    enum: ["low", "medium", "high", "critical"],
    required: true,
  },
  description: { type: String },
  remediation: { type: String },
});

const ScanSchema = new mongoose.Schema(
  {
    user: {
      type: mongoose.Schema.Types.ObjectId,
      ref: "User",
      required: true,
    },
    targetURL: {
      type: String,
      required: true,
      trim: true,
    },
    config: {
      depth: { type: Number, default: 1 },
      headers: { type: Object, default: {} },
    },
    status: {
      type: String,
      enum: ["pending", "running", "completed", "cancelled", "failed"],
      default: "pending",
      required: true,
    },
    vulnerabilities: [VulnerabilitySchema],
    createdAt: { type: Date, default: Date.now },
    completedAt: { type: Date },
  },
  { timestamps: true }
);

module.exports = mongoose.model("Scan", ScanSchema);
