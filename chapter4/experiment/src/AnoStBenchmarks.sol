// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract AnoStBenchmarks {
    mapping(bytes32 => bool) public regNullifiers;
    mapping(bytes32 => bool) public stakeNullifiers;
    mapping(bytes32 => bool) public credNullifiers;
    bytes32[] public registry;

    function publicReg(bytes32 identityCommitment) external {
        registry.push(identityCommitment);
    }

    function publicStake(bytes32 stakeCommitment) external pure returns (bytes32) {
        return keccak256(abi.encodePacked(stakeCommitment));
    }

    function candidateDeclare(bytes32 candidateId, uint256 stakeValue) external pure returns (bytes32) {
        require(stakeValue >= 32 ether, "low stake");
        return keccak256(abi.encodePacked(candidateId, stakeValue));
    }

    function anonyReg(bytes32 regCommitment, bytes32 depCommitment, bytes32 regNullifier, bytes calldata proof) external {
        require(!regNullifiers[regNullifier], "dup reg");
        _verifyGroth16Like(proof, 2);
        regNullifiers[regNullifier] = true;
        registry.push(keccak256(abi.encodePacked(regCommitment, depCommitment)));
    }

    function anonyStake(bytes32 stakeCommitment, bytes32 changeCommitment, bytes32 stakeNullifier, bytes calldata proof) external {
        require(!stakeNullifiers[stakeNullifier], "dup stake");
        _verifyGroth16Like(proof, 2);
        stakeNullifiers[stakeNullifier] = true;
        bytes32 sink = keccak256(abi.encodePacked(stakeCommitment, changeCommitment));
        require(sink != bytes32(0), "bad commitment");
    }

    function presentCred(bytes32 credNullifier, bytes32[] calldata merklePath, uint256 stakeValue, bytes calldata proof) external {
        require(stakeValue >= 32 ether, "low stake");
        require(!credNullifiers[credNullifier], "dup cred");
        _verifyGroth16Like(proof, 3);
        bytes32 acc = credNullifier;
        for (uint256 i = 0; i < merklePath.length; i++) {
            acc = keccak256(abi.encodePacked(acc, merklePath[i]));
        }
        credNullifiers[credNullifier] = true;
        require(acc != bytes32(0), "bad path");
    }

    function electCTWR(uint256[] calldata stakes, uint256 committeeSize, uint256 sCap, uint256 seed) external pure returns (bytes32) {
        uint256 n = stakes.length;
        bool[] memory selected = new bool[](n);
        bytes32 transcript;
        for (uint256 round = 0; round < committeeSize; round++) {
            uint256 total;
            for (uint256 i = 0; i < n; i++) {
                if (!selected[i]) total += stakes[i] < sCap ? stakes[i] : sCap;
            }
            uint256 draw = uint256(keccak256(abi.encodePacked(seed, round))) % total;
            uint256 prefix;
            for (uint256 i = 0; i < n; i++) {
                if (selected[i]) continue;
                prefix += stakes[i] < sCap ? stakes[i] : sCap;
                if (draw < prefix) {
                    selected[i] = true;
                    transcript = keccak256(abi.encodePacked(transcript, i));
                    break;
                }
            }
        }
        return transcript;
    }

    function _verifyGroth16Like(bytes calldata proof, uint256 pairings) internal pure {
        require(proof.length >= 192, "short proof");
        bytes32 acc;
        for (uint256 i = 0; i < pairings * 80; i++) {
            acc = keccak256(abi.encodePacked(acc, proof, i));
        }
        require(acc != bytes32(0), "invalid proof");
    }
}
