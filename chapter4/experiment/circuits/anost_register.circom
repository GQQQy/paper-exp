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

template AnostRegister() {
    signal input epoch;
    signal input user_key;
    signal input tag_randomness;
    signal input deposit_value;
    signal input deposit_randomness;
    signal input d_min;
    signal output registration_commitment;
    signal output deposit_commitment;
    signal output registration_nullifier;

    component tag_hash = Hash2();
    tag_hash.in[0] <== epoch;
    tag_hash.in[1] <== user_key;

    component reg_commit = Commitment();
    reg_commit.value <== tag_hash.out;
    reg_commit.randomness <== tag_randomness;
    registration_commitment <== reg_commit.commitment;

    component dep_commit = Commitment();
    dep_commit.value <== deposit_value;
    dep_commit.randomness <== deposit_randomness;
    deposit_commitment <== dep_commit.commitment;

    component nullifier_hash = Hash2();
    nullifier_hash.in[0] <== epoch;
    nullifier_hash.in[1] <== tag_hash.out;
    registration_nullifier <== nullifier_hash.out;

    component deposit_ok = GreaterEqThan(128);
    deposit_ok.in[0] <== deposit_value;
    deposit_ok.in[1] <== d_min;
    deposit_ok.out === 1;
}

component main { public [epoch, d_min] } = AnostRegister();
