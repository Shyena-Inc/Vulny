// src/scan-engine/index.js
// This is the main scan engine coordinator module.
// It will be called by BullMQ worker to perform scan logic.

const Scan = require("../models/Scan");
const portScan = require("./plugins/portScan");
const subdomainEnum = require("./plugins/subdomainEnum");
const dirBruteForce = require("./plugins/dirBruteForce");
const vulnChecks = require("./plugins/vulnChecks");

// Run scan
async function runScan(scanId) {
  // Load scan document
  const scan = await Scan.findById(scanId);
  if (!scan) throw new Error("Scan not found");

  // Mark scan as running
  scan.status = "running";
  scan.completedAt = undefined;
  await scan.save();

  const vulnerabilities = [];

  try {
    // 1. Port Scan (optional)
    const ports = await portScan(scan.targetURL);

    // 2. Subdomain Enumeration
    const subdomains = await subdomainEnum(scan.targetURL);

    // 3. Directory Brute Force
    const directories = await dirBruteForce(scan.targetURL, scan.config.depth);

    // 4. Vulnerability Checks
    const vulns = await vulnChecks(scan.targetURL);

    // Aggregate vulnerabilities
    vulnerabilities.push(...vulns);

    // Save results
    scan.vulnerabilities = vulnerabilities;
    scan.status = "completed";
    scan.completedAt = new Date();
    await scan.save();

    return scan;
  } catch (error) {
    scan.status = "failed";
    scan.completedAt = new Date();
    await scan.save();
    throw error;
  }
}

module.exports = { runScan };
