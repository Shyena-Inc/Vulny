// src/scan-engine/plugins/subdomainEnum.js
// Dummy subdomain enumeration plugin

async function subdomainEnum(targetURL) {
  const urlObj = new URL(targetURL);
  const domain = urlObj.hostname;

  // Simulate subdomains found
  const foundSubdomains = [`dev.${domain}`, `test.${domain}`];

  return foundSubdomains;
}

module.exports = subdomainEnum;
