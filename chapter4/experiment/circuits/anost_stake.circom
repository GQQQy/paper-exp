pragma circom 2.1.8;

include "../node_modules/circomlib/circuits/poseidon.circom";
include "../node_modules/circomlib/circuits/comparators.circom";

template Hash2() {
    signal input in[2];
    signal output out;

    component h = Poseidon(2);
    h.inputs[0] <== in[0];
    h.inputs[1] <== in[1];
    out <== h.out;
}

template Commitment() {
    signal input value;
    signal input randomness;
    signal output commitment;

    component h = Hash2();
    h.in[0] <== value;
    h.in[1] <== randomness;
    commitment <== h.out;
}

template AnostStake() {
    signal input epoch;
    signal input tag;
    signal input index;
    signal input deposit_value;
    signal input deposit_commitment;
    signal input stake_value;
    signal input deposit_randomness;
    signal input stake_randomness;
    signal input change_randomness;
    signal output stake_commitment;
    signal output change_commitment;
    signal output stake_nullifier;

    signal change_value;
    change_value <== deposit_value - stake_value;

    component source_commit = Commitment();
    source_commit.value <== deposit_value;
    source_commit.randomness <== deposit_randomness;
    source_commit.commitment === deposit_commitment;

    component stake_ok = GreaterEqThan(128);
    stake_ok.in[0] <== deposit_value;
    stake_ok.in[1] <== stake_value;
    stake_ok.out === 1;

    component stake_commit = Commitment();
    stake_commit.value <== stake_value;
    stake_commit.randomness <== stake_randomness;
    stake_commitment <== stake_commit.commitment;

    component change_commit = Commitment();
    change_commit.value <== change_value;
    change_commit.randomness <== change_randomness;
    change_commitment <== change_commit.commitment;

    component n1 = Hash2();
    n1.in[0] <== epoch;
    n1.in[1] <== tag;
    component n2 = Hash2();
    n2.in[0] <== n1.out;
    n2.in[1] <== index;
    stake_nullifier <== n2.out;
}

component main { public [epoch, deposit_commitment] } = AnostStake();
