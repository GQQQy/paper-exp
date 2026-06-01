// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract CleVerVerifier {
    enum Verdict {
        AcceptChallenger,
        AcceptExecutor,
        BothWrong
    }

    function commit(bytes memory stateData) public pure returns (bytes32) {
        return keccak256(stateData);
    }

    function verSeg(
        bytes32 startState,
        uint256 startValue,
        uint256 steps,
        uint256 executorClaim,
        uint256 challengerClaim
    ) external pure returns (Verdict) {
        uint256 value = uint256(startState) ^ startValue;
        for (uint256 i = 0; i < steps; i++) {
            value = uint256(keccak256(abi.encodePacked(value, i)));
        }
        if (value == challengerClaim) {
            return Verdict.AcceptChallenger;
        }
        if (value == executorClaim) {
            return Verdict.AcceptExecutor;
        }
        return Verdict.BothWrong;
    }
}
