// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract DisputeProtocolBenchmarks {
    bytes32 internal accumulator;

    function arbitrumOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xA0);
    }

    function truebitOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xB0);
    }

    function cartesiOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xC0);
    }

    function boldOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xD0);
    }

    function cleverOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xE0);
    }

    function optimisticPath(uint256 rounds, uint256 tag) internal returns (bytes32) {
        bytes32 h = accumulator;
        for (uint256 i = 0; i < rounds; i++) {
            h = keccak256(abi.encodePacked(h, i, tag));
        }
        accumulator = h;
        return h;
    }

    function arbitrumClassicPath(uint256 rounds) external returns (bytes32) {
        bytes32 h = accumulator;
        for (uint256 i = 0; i < rounds; i++) {
            h = keccak256(abi.encodePacked(h, i, uint256(0xA1)));
        }
        accumulator = h;
        return h;
    }

    function truebitPath(uint256 rounds) external returns (bytes32) {
        bytes32 h = accumulator;
        for (uint256 i = 0; i < rounds; i++) {
            h = keccak256(abi.encodePacked(h, i, uint256(0xB2)));
            if ((i & 1) == 0) {
                h = keccak256(abi.encodePacked(h, uint256(0xB3)));
            }
        }
        accumulator = h;
        return h;
    }

    function cartesiDavePath(uint256 rounds, uint256 tournamentSlots) external returns (bytes32) {
        bytes32 h = accumulator;
        for (uint256 slot = 0; slot < tournamentSlots; slot++) {
            h = keccak256(abi.encodePacked(h, slot, uint256(0xC1)));
        }
        for (uint256 i = 0; i < rounds; i++) {
            h = keccak256(abi.encodePacked(h, i, uint256(0xC2)));
            h = keccak256(abi.encodePacked(h, uint256(0xC3)));
        }
        accumulator = h;
        return h;
    }

    function boldPath(uint256 rounds, uint256 levels) external returns (bytes32) {
        bytes32 h = accumulator;
        for (uint256 level = 0; level < levels; level++) {
            for (uint256 i = 0; i < rounds; i++) {
                h = keccak256(abi.encodePacked(h, level, i, uint256(0xD1)));
            }
        }
        accumulator = h;
        return h;
    }

    function cleverPath(uint256 subsegments, uint256 replaySteps) external returns (bytes32) {
        bytes32 h = accumulator;
        for (uint256 i = 0; i < subsegments; i++) {
            h = keccak256(abi.encodePacked(h, i, uint256(0xE1)));
        }
        for (uint256 i = 0; i < replaySteps; i++) {
            h = keccak256(abi.encodePacked(h, i, uint256(0xE2)));
        }
        accumulator = h;
        return h;
    }
}
