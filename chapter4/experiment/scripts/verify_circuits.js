#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const circomlibjs = require("circomlibjs");

const ROOT = path.resolve(__dirname, "..");
const BUILD = path.join(ROOT, "build", "circuits");
const WITNESS = path.join(ROOT, "build", "witness");
const SNARKJS = path.join(ROOT, "node_modules", ".bin", "snarkjs");

function writeJson(file, data) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, `${JSON.stringify(data, null, 2)}\n`);
}

function run(cmd, args) {
  execFileSync(cmd, args, {
    cwd: ROOT,
    stdio: "inherit",
  });
}

function readWitness(name) {
  return JSON.parse(fs.readFileSync(path.join(WITNESS, `${name}.witness.json`), "utf8"));
}

function assertEqual(actual, expected, label) {
  if (actual.toString() !== expected.toString()) {
    throw new Error(`${label} mismatch\nactual:   ${actual}\nexpected: ${expected}`);
  }
}

async function main() {
  fs.mkdirSync(WITNESS, { recursive: true });

  const poseidon = await circomlibjs.buildPoseidon();
  const F = poseidon.F;
  const hash2 = (a, b) => F.toObject(poseidon([BigInt(a), BigInt(b)])).toString();

  const params = {
    epoch: "4",
    user_key: "123456789",
    tag_randomness: "98765",
    deposit_value: "64",
    deposit_randomness: "24680",
    d_min: "32",
    index: "7",
    stake_value: "32",
    stake_randomness: "13579",
    change_randomness: "97531",
  };

  const tag = hash2(params.epoch, params.user_key);
  const registrationCommitment = hash2(tag, params.tag_randomness);
  const depositCommitment = hash2(params.deposit_value, params.deposit_randomness);
  const registrationNullifier = hash2(params.epoch, tag);

  const registerInput = {
    epoch: params.epoch,
    user_key: params.user_key,
    tag_randomness: params.tag_randomness,
    deposit_value: params.deposit_value,
    deposit_randomness: params.deposit_randomness,
    d_min: params.d_min,
  };
  writeJson(path.join(WITNESS, "anost_register.input.json"), registerInput);

  const stakeCommitment = hash2(params.stake_value, params.stake_randomness);
  const changeValue = (BigInt(params.deposit_value) - BigInt(params.stake_value)).toString();
  const changeCommitment = hash2(changeValue, params.change_randomness);
  const stakeNullifier = hash2(hash2(params.epoch, tag), params.index);

  const stakeInput = {
    epoch: params.epoch,
    tag,
    index: params.index,
    deposit_value: params.deposit_value,
    deposit_commitment: depositCommitment,
    stake_value: params.stake_value,
    deposit_randomness: params.deposit_randomness,
    stake_randomness: params.stake_randomness,
    change_randomness: params.change_randomness,
  };
  writeJson(path.join(WITNESS, "anost_stake.input.json"), stakeInput);

  const merklePath = Array.from({ length: 20 }, (_, i) => (1000 + i).toString());
  const merkleIndex = Array.from({ length: 20 }, (_, i) => (i % 2).toString());
  let merkleRoot = registrationCommitment;
  for (let i = 0; i < merklePath.length; i += 1) {
    const left = merkleIndex[i] === "1" ? merklePath[i] : merkleRoot;
    const right = merkleIndex[i] === "1" ? merkleRoot : merklePath[i];
    merkleRoot = hash2(left, right);
  }
  const credentialNullifier = hash2(stakeNullifier, "112233");

  const credentialInput = {
    epoch: params.epoch,
    tag,
    index: params.index,
    registration_commitment: registrationCommitment,
    merkle_root: merkleRoot,
    stake_value: params.stake_value,
    stake_randomness: params.stake_randomness,
    merkle_path: merklePath,
    merkle_index: merkleIndex,
  };
  writeJson(path.join(WITNESS, "anost_credential.input.json"), credentialInput);

  const circuits = [
    {
      name: "anost_register",
      wasm: path.join(BUILD, "anost_register_js", "anost_register.wasm"),
      generator: path.join(BUILD, "anost_register_js", "generate_witness.js"),
      r1cs: path.join(BUILD, "anost_register.r1cs"),
    },
    {
      name: "anost_stake",
      wasm: path.join(BUILD, "anost_stake_js", "anost_stake.wasm"),
      generator: path.join(BUILD, "anost_stake_js", "generate_witness.js"),
      r1cs: path.join(BUILD, "anost_stake.r1cs"),
    },
    {
      name: "anost_credential",
      wasm: path.join(BUILD, "anost_credential_js", "anost_credential.wasm"),
      generator: path.join(BUILD, "anost_credential_js", "generate_witness.js"),
      r1cs: path.join(BUILD, "anost_credential.r1cs"),
    },
  ];

  for (const circuit of circuits) {
    const input = path.join(WITNESS, `${circuit.name}.input.json`);
    const wtns = path.join(WITNESS, `${circuit.name}.wtns`);
    const exported = path.join(WITNESS, `${circuit.name}.witness.json`);
    run(process.execPath, [circuit.generator, circuit.wasm, input, wtns]);
    run(SNARKJS, ["wtns", "check", circuit.r1cs, wtns]);
    run(SNARKJS, ["wtns", "export", "json", wtns, exported]);
  }

  const regWitness = readWitness("anost_register");
  assertEqual(regWitness[1], registrationCommitment, "registration_commitment");
  assertEqual(regWitness[2], depositCommitment, "deposit_commitment");
  assertEqual(regWitness[3], registrationNullifier, "registration_nullifier");

  const stakeWitness = readWitness("anost_stake");
  assertEqual(stakeWitness[1], stakeCommitment, "stake_commitment");
  assertEqual(stakeWitness[2], changeCommitment, "change_commitment");
  assertEqual(stakeWitness[3], stakeNullifier, "stake_nullifier");

  const credWitness = readWitness("anost_credential");
  assertEqual(credWitness[1], credentialNullifier, "credential_nullifier");
  assertEqual(credWitness[2], stakeCommitment, "credential stake_commitment");

  const summary = {
    tooling: {
      circom: "2.1.8",
      snarkjs: "0.7.6",
      poseidon: "circomlib/circomlibjs Poseidon(2)",
    },
    params,
    commitments: {
      tag,
      registration_commitment: registrationCommitment,
      deposit_commitment: depositCommitment,
      registration_nullifier: registrationNullifier,
      stake_commitment: stakeCommitment,
      change_commitment: changeCommitment,
      stake_nullifier: stakeNullifier,
      merkle_root: merkleRoot,
      credential_nullifier: credentialNullifier,
    },
    checks: circuits.map((c) => ({
      circuit: c.name,
      r1cs: path.relative(ROOT, c.r1cs),
      witness: path.relative(ROOT, path.join(WITNESS, `${c.name}.wtns`)),
      status: "PASS",
    })),
  };
  writeJson(path.join(WITNESS, "circuit_verification_summary.json"), summary);
  console.log(JSON.stringify(summary, null, 2));
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
