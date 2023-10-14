const { execSync } = require("node:child_process");

const isThisHostAlive = (host) => {
  try {
    execSync(`ping -c 5 -s 1000 ${host}`, {
      stdio: "inherit",
    });
    return true;
  } catch (err) {
    console.log(`Failed to ping ${host}`, err);
    return false;
  }
};

exports.isThisHostAlive = isThisHostAlive;
