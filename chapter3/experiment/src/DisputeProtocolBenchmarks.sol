// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract DisputeProtocolBenchmarks {
    bytes32 internal accumulator;
    mapping(bytes32 => bytes32) public submittedClaims;
    mapping(bytes32 => bytes32) public activeChallenges;
    mapping(bytes32 => bytes32) public localizedSegments;
    mapping(bytes32 => bytes32) public finalVerdicts;

    function arbitrumOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xA0, "arbitrum-classic-submit");
    }

    function truebitOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xB0, "truebit-task-submit");
    }

    function cartesiOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xC0, "cartesi-tournament-claim");
    }

    function boldOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xD0, "bold-assertion-tree");
    }

    function cleverOptimisticPath(uint256 rounds) external returns (bytes32) {
        return optimisticPath(rounds, 0xE0, "clever-sliced-task");
    }

    function optimisticPath(uint256 rounds, uint256 tag, bytes32 label) internal returns (bytes32) {
        bytes32 claim = _submitClaim(label, tag, rounds);
        bytes32 root = _commitCheckpoints(claim, rounds, tag);
        return _acceptUnchallenged(claim, root, tag);
    }

    function arbitrumClassicPath(uint256 rounds) external returns (bytes32) {
        bytes32 claim = _submitClaim("arbitrum-classic-assertion", 0xA1, rounds);
        bytes32 challenge = _openChallenge(claim, "one-vs-one-bisection", 0xA2);
        bytes32 segment = _bisectInterval(challenge, rounds, 0xA3);
        _recordLocalization(challenge, segment, "single-step-location");
        bytes32 proof = _oneStepProof(segment, 19, 0xA4);
        return _adjudicate(challenge, proof, "avm-one-step-proof");
    }

    function truebitPath(uint256 rounds) external returns (bytes32) {
        bytes32 claim = _submitClaim("truebit-solver-commitment", 0xB1, rounds);
        bytes32 challenge = _openChallenge(claim, "solver-verifier-game", 0xB2);
        bytes32 segment = _solverVerifierBisection(challenge, rounds, 0xB3);
        _recordLocalization(challenge, segment, "instruction-index-location");
        bytes32 proof = _oneStepProof(segment, 17, 0xB4);
        return _adjudicate(challenge, proof, "truebit-final-judge");
    }

    function cartesiDavePath(uint256 rounds, uint256 tournamentSlots) external returns (bytes32) {
        bytes32 tournament = _setupTournament("cartesi-dave-tournament", tournamentSlots, 0xC1);
        bytes32 challenge = _openChallenge(tournament, "cartesi-claim-counterclaim", 0xC2);
        bytes32 segment = _cartesiBisection(challenge, rounds, 0xC3);
        _recordLocalization(challenge, segment, "machine-step-location");
        bytes32 proof = _oneStepProof(segment, 23, 0xC4);
        return _adjudicate(challenge, proof, "dave-referee");
    }

    function boldPath(uint256 rounds, uint256 levels) external returns (bytes32) {
        bytes32 root = _submitClaim("bold-top-level-assertion", 0xD1, rounds * levels);
        bytes32 challenge = _openChallenge(root, "bold-parallel-claim-edges", 0xD2);
        bytes32 segment = _boldMultiLevelSearch(challenge, rounds, levels, 0xD3);
        _recordLocalization(challenge, segment, "lowest-level-edge");
        bytes32 proof = _oneStepProof(segment, 13 + levels, 0xD4);
        return _adjudicate(challenge, proof, "bold-confirmation");
    }

    function cleverPath(uint256 subsegments, uint256 replaySteps) external returns (bytes32) {
        bytes32 claim = _submitClaim("clever-two-layer-slice-root", 0xE1, subsegments);
        bytes32 challenge = _openChallenge(claim, "outer-inner-slice-challenge", 0xE2);
        bytes32 segment = _cleverLocateSegment(challenge, subsegments, 0xE3);
        _recordLocalization(challenge, segment, "bounded-subsegment");
        bytes32 proof = _versegReplay(segment, replaySteps, 0xE4);
        return _adjudicate(challenge, proof, "clever-verseg");
    }

    function _submitClaim(bytes32 label, uint256 tag, uint256 size) internal returns (bytes32 claim) {
        claim = keccak256(abi.encodePacked(label, accumulator, msg.sender, tag, size));
        submittedClaims[claim] = keccak256(abi.encodePacked("state-root", claim, size));
        accumulator = claim;
    }

    function _acceptUnchallenged(bytes32 claim, bytes32 root, uint256 tag) internal returns (bytes32 verdict) {
        verdict = keccak256(abi.encodePacked("optimistic-accept", claim, root, tag));
        finalVerdicts[claim] = verdict;
        accumulator = verdict;
    }

    function _openChallenge(bytes32 claim, bytes32 label, uint256 tag) internal returns (bytes32 challenge) {
        challenge = keccak256(abi.encodePacked(label, claim, tag, block.number));
        activeChallenges[claim] = challenge;
        accumulator = challenge;
    }

    function _recordLocalization(bytes32 challenge, bytes32 segment, bytes32 label) internal {
        localizedSegments[challenge] = keccak256(abi.encodePacked(label, challenge, segment));
    }

    function _adjudicate(bytes32 challenge, bytes32 proof, bytes32 label) internal returns (bytes32 verdict) {
        verdict = keccak256(abi.encodePacked(label, challenge, proof));
        finalVerdicts[challenge] = verdict;
        accumulator = verdict;
    }

    function _commitCheckpoints(bytes32 h, uint256 rounds, uint256 tag) internal pure returns (bytes32) {
        for (uint256 i = 0; i < rounds; i++) {
            h = keccak256(abi.encodePacked("checkpoint", h, i, tag));
        }
        return h;
    }

    function _bisectInterval(bytes32 h, uint256 rounds, uint256 tag) internal pure returns (bytes32) {
        for (uint256 i = 0; i < rounds; i++) {
            h = keccak256(abi.encodePacked("bisect", h, i, tag));
        }
        return h;
    }

    function _solverVerifierBisection(bytes32 h, uint256 rounds, uint256 tag) internal pure returns (bytes32) {
        for (uint256 i = 0; i < rounds; i++) {
            h = keccak256(abi.encodePacked("solver-step", h, i, tag));
            if ((i & 1) == 0) {
                h = keccak256(abi.encodePacked("verifier-response", h, tag));
            }
        }
        return h;
    }

    function _setupTournament(bytes32 label, uint256 tournamentSlots, uint256 tag) internal returns (bytes32 h) {
        h = _submitClaim(label, tag, tournamentSlots);
        for (uint256 slot = 0; slot < tournamentSlots; slot++) {
            h = keccak256(abi.encodePacked("tournament-match", h, slot, tag));
        }
        localizedSegments[h] = keccak256(abi.encodePacked("tournament-winner", h));
    }

    function _cartesiBisection(bytes32 h, uint256 rounds, uint256 tag) internal pure returns (bytes32) {
        for (uint256 i = 0; i < rounds; i++) {
            h = keccak256(abi.encodePacked("machine-claim", h, i, tag));
            h = keccak256(abi.encodePacked("counter-claim", h, tag));
        }
        return h;
    }

    function _boldMultiLevelSearch(bytes32 h, uint256 rounds, uint256 levels, uint256 tag) internal pure returns (bytes32) {
        for (uint256 level = 0; level < levels; level++) {
            h = keccak256(abi.encodePacked("bold-level-open", h, level, tag));
            for (uint256 i = 0; i < rounds; i++) {
                h = keccak256(abi.encodePacked("bold-edge", h, level, i, tag));
            }
            h = keccak256(abi.encodePacked("bold-level-confirm", h, level, tag));
        }
        return h;
    }

    function _cleverLocateSegment(bytes32 h, uint256 subsegments, uint256 tag) internal pure returns (bytes32) {
        for (uint256 i = 0; i < subsegments; i++) {
            h = keccak256(abi.encodePacked("slice-localize", h, i, tag));
        }
        return h;
    }

    function _versegReplay(bytes32 h, uint256 replaySteps, uint256 tag) internal pure returns (bytes32) {
        for (uint256 i = 0; i < replaySteps; i++) {
            h = keccak256(abi.encodePacked("verseg-step", h, i, tag));
        }
        return h;
    }

    function _oneStepProof(bytes32 h, uint256 replaySteps, uint256 tag) internal pure returns (bytes32) {
        for (uint256 i = 0; i < replaySteps; i++) {
            h = keccak256(abi.encodePacked("single-step", h, i, tag));
        }
        return h;
    }
}
