// src/models/Plugin.js
const mongoose = require("mongoose");

const PluginSchema = new mongoose.Schema(
  {
    name: { type: String, unique: true, required: true, trim: true },
    enabled: { type: Boolean, default: true },
    config: { type: Object, default: {} },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date },
  },
  { timestamps: true }
);

module.exports = mongoose.model("Plugin", PluginSchema);
