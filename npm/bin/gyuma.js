#!/usr/bin/env node

// gyuma バイナリを実行するラッパースクリプト

const { spawn } = require("child_process");
const path = require("path");
const os = require("os");

const binaryName = os.platform() === "win32" ? "gyuma.exe" : "gyuma";
const binaryPath = path.join(__dirname, binaryName);

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
});

child.on("error", (err) => {
  if (err.code === "ENOENT") {
    console.error("gyuma binary not found. Try reinstalling the package:");
    console.error("  npm uninstall gyuma && npm install gyuma");
  } else {
    console.error(`Failed to run gyuma: ${err.message}`);
  }
  process.exit(1);
});

child.on("exit", (code) => {
  process.exit(code ?? 0);
});
