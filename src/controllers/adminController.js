// src/controllers/adminController.js
const User = require("../models/User");
const Scan = require("../models/Scan");
const Plugin = require("../models/Plugin"); // Plugin model for managing scanner modules

// Get all users
const getAllUsers = async (req, res, next) => {
  try {
    const users = await User.find().select("-password").sort({ createdAt: -1 });
    return res.json({ users });
  } catch (err) {
    next(err);
  }
};

// Get basic scan stats
const getScanStats = async (req, res, next) => {
  try {
    const totalScans = await Scan.countDocuments();
    const scansByStatus = await Scan.aggregate([
      {
        $group: {
          _id: "$status",
          count: { $sum: 1 },
        },
      },
    ]);

    const scansByUser = await Scan.aggregate([
      {
        $group: {
          _id: "$user",
          count: { $sum: 1 },
        },
      },
      { $limit: 10 },
    ]);

    return res.json({
      totalScans,
      scansByStatus,
      topUsersByScan: scansByUser,
    });
  } catch (err) {
    next(err);
  }
};

// Plugin management controllers
// Assuming plugins have structure: name, enabled(boolean), config(object), createdAt, updatedAt
const getPlugins = async (req, res, next) => {
  try {
    const plugins = await Plugin.find().sort({ createdAt: -1 });
    return res.json({ plugins });
  } catch (err) {
    next(err);
  }
};
const addPlugin = async (req, res, next) => {
  try {
    const { name, config } = req.body;
    if (!name)
      return res.status(400).json({ error: "Plugin name is required" });

    const existing = await Plugin.findOne({ name });
    if (existing)
      return res.status(409).json({ error: "Plugin name already exists" });

    const plugin = new Plugin({ name, config: config || {}, enabled: true });
    await plugin.save();

    return res.status(201).json({ message: "Plugin added", plugin });
  } catch (err) {
    next(err);
  }
};

const updatePluginStatus = async (req, res, next) => {
  try {
    const pluginId = req.params.id;
    const { enabled } = req.body;
    if (typeof enabled !== "boolean")
      return res.status(400).json({ error: "Enabled boolean value required" });
    const plugin = await Plugin.findById(pluginId);
    if (!plugin) return res.status(404).json({ error: "Plugin not found" });

    plugin.enabled = enabled;
    await plugin.save();

    return res.json({ message: "Plugin status updated", plugin });
  } catch (err) {
    next(err);
  }
};

const deletePlugin = async (req, res, next) => {
  try {
    const pluginId = req.params.id;
    const plugin = await Plugin.findById(pluginId);
    if (!plugin) return res.status(404).json({ error: "Plugin not found" });

    await plugin.remove();

    return res.json({ message: "Plugin deleted" });
  } catch (err) {
    next(err);
  }
};

module.exports = {
  getAllUsers,
  getScanStats,
  getPlugins,
  addPlugin,
  updatePluginStatus,
  deletePlugin,
};
