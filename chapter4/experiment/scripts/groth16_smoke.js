#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const ROOT = path.resolve(__dirname, "..");
const BUILD = path.join(ROOT, "build", "circuits");
const WITNESS = path.join(ROOT, "build", "witness");
const ZK = path.join(ROOT, "build", "zk");
const SNARKJS = path.join(ROOT, "node_modules", ".bin", "snarkjs");
const PTAU = path.join(ZK, "pot13_final.ptau");

function run(args) {
  try {
    return execFileSync(SNARKJS, args, {
      cwd: ROOT,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
      maxBuffer: 64 * 1024 * 1024,
    });
  } catch (err) {
    if (err.stdout) process.stderr.write(err.stdout);
    if (err.stderr) process.stderr.write(err.stderr);
    throw err;
  }
}

function writeJson(file, value) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function main() {
  if (!fs.existsSync(PTAU)) {
    throw new Error(`Missing ${path.relative(ROOT, PTAU)}. Run powersoftau phase2 first.`);
  }

  const circuits = ["anost_register", "anost_stake", "anost_credential"];
  const results = [];

  for (const name of circuits) {
    const r1cs = path.join(BUILD, `${name}.r1cs`);
    const witness = path.join(WITNESS, `${name}.wtns`);
    const zkey0 = path.join(ZK, `${name}_0000.zkey`);
    const zkeyFinal = path.join(ZK, `${name}_final.zkey`);
    const vkey = path.join(ZK, `${name}_verification_key.json`);
    const proof = path.join(ZK, `${name}_proof.json`);
    const publicSignals = path.join(ZK, `${name}_public.json`);

    run(["groth16", "setup", r1cs, PTAU, zkey0]);
    run([
      "zkey",
      "contribute",
      zkey0,
      zkeyFinal,
      "--name=chapter4-local-smoke",
      "-e=chapter4 deterministic local smoke entropy",
    ]);
    run(["zkey", "verify", r1cs, PTAU, zkeyFinal]);
    run(["zkey", "export", "verificationkey", zkeyFinal, vkey]);

    const proveStart = Date.now();
    run(["groth16", "prove", zkeyFinal, witness, proof, publicSignals]);
    const provingTimeMs = Date.now() - proveStart;

    const verifyStart = Date.now();
    const verifyOutput = run(["groth16", "verify", vkey, publicSignals, proof]);
    const verificationTimeMs = Date.now() - verifyStart;
    if (!verifyOutput.includes("OK")) {
      throw new Error(`${name} Groth16 verify did not report OK`);
    }

    results.push({
      circuit: name,
      r1cs: path.relative(ROOT, r1cs),
      zkey: path.relative(ROOT, zkeyFinal),
      proof: path.relative(ROOT, proof),
      public: path.relative(ROOT, publicSignals),
      proving_time_ms: provingTimeMs,
      verification_time_ms: verificationTimeMs,
      status: "PASS",
      note: "Local smoke-test Groth16 parameters, not a production trusted setup.",
    });
  }

  const summary = {
    ptau: path.relative(ROOT, PTAU),
    generated_at: new Date().toISOString(),
    results,
  };
  writeJson(path.join(ZK, "groth16_smoke_summary.json"), summary);
  console.log(JSON.stringify(summary, null, 2));
}

main();
