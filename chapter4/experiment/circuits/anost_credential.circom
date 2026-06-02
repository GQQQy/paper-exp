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

template MerklePath(levels) {
    signal input leaf;
    signal input path_elements[levels];
    signal input path_indices[levels];
    signal output root;
    signal cur[levels + 1];
    cur[0] <== leaf;
    signal left[levels];
    signal right[levels];
    component h[levels];

    for (var i = 0; i < levels; i++) {
        path_indices[i] * (path_indices[i] - 1) === 0;

        left[i] <== cur[i] + path_indices[i] * (path_elements[i] - cur[i]);
        right[i] <== path_elements[i] + path_indices[i] * (cur[i] - path_elements[i]);
        h[i] = Hash2();
        h[i].in[0] <== left[i];
        h[i].in[1] <== right[i];
        cur[i + 1] <== h[i].out;
    }
    root <== cur[levels];
}

template AnostCredential(levels) {
    signal input epoch;
    signal input tag;
    signal input index;
    signal input registration_commitment;
    signal input merkle_root;
    signal input stake_value;
    signal input stake_randomness;
    signal input merkle_path[levels];
    signal input merkle_index[levels];
    signal output credential_nullifier;
    signal output stake_commitment;

    component path = MerklePath(levels);
    path.leaf <== registration_commitment;
    for (var i = 0; i < levels; i++) {
        path.path_elements[i] <== merkle_path[i];
        path.path_indices[i] <== merkle_index[i];
    }
    path.root === merkle_root;

    component stake_hash = Hash2();
    stake_hash.in[0] <== stake_value;
    stake_hash.in[1] <== stake_randomness;
    stake_commitment <== stake_hash.out;

    component n1 = Hash2();
    n1.in[0] <== epoch;
    n1.in[1] <== tag;
    component n2 = Hash2();
    n2.in[0] <== n1.out;
    n2.in[1] <== index;
    component n3 = Hash2();
    n3.in[0] <== n2.out;
    n3.in[1] <== 112233;
    credential_nullifier <== n3.out;
}

component main { public [epoch, merkle_root, stake_value] } = AnostCredential(20);
