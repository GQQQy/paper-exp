// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract ValidatorAudit {
    uint256 public constant C_HB = 1;
    uint256 public constant C_SENT = 2;
    uint256 public constant M_DEFAULT = 100;
    uint256 public constant DELTA_DEFAULT = 7;

    mapping(bytes32 => mapping(address => bytes32)) public trackCommitments;
    mapping(bytes32 => mapping(address => bytes32)) public contAuditReceipts;
    mapping(bytes32 => mapping(address => bytes32)) public contAuditEndpoints;
    mapping(bytes32 => mapping(address => bytes32)) public reportDigests;
    mapping(bytes32 => mapping(address => bytes32)) public disputeReceipts;
    mapping(bytes32 => mapping(address => bool)) public slashed;
    mapping(bytes32 => bytes32) public audRoots;
    mapping(bytes32 => mapping(address => mapping(uint256 => bytes32))) public podProofs;
    mapping(bytes32 => mapping(uint256 => bytes32)) public podStateRoots;
    mapping(bytes32 => mapping(uint256 => bytes32)) public podTraceRoots;
    mapping(bytes32 => mapping(uint256 => address)) public podWinners;
    mapping(bytes32 => mapping(address => mapping(uint256 => bytes32))) public podBountyReceipts;

    event TrackInitialized(bytes32 indexed tid, address indexed validator, bytes32 tc0);
    event Heartbeat(bytes32 indexed tid, address indexed validator, uint256 blockNumber, bytes32 commitment);
    event ContinuousAudit(bytes32 indexed tid, address indexed validator, bool ok);
    event SentinelReport(bytes32 indexed tid, address indexed validator, bytes32 digest);
    event Disputed(bytes32 indexed tid, address indexed validator, bytes32 claim);
    event Slashed(bytes32 indexed tid, address indexed validator, uint256 amount);
    event PoDBountyMined(bytes32 indexed tid, address indexed watchtower, uint256 indexed epoch, bytes32 receipt, bool eligible);

    function trackInit(bytes32 tid, bytes32 seedHash, bytes32 nonceHash) external returns (bytes32 tc0) {
        tc0 = keccak256(abi.encodePacked("TC0", tid, msg.sender, seedHash, nonceHash));
        tc0 = _mixDigest(tc0, tid, 8);
        trackCommitments[tid][msg.sender] = tc0;
        emit TrackInitialized(tid, msg.sender, tc0);
    }

    function heartbeatTriggered(bytes32 bh, bytes32 tid, address validator, uint256 m) public pure returns (bool) {
        return uint256(keccak256(abi.encodePacked(bh, tid, validator))) % m == 0;
    }

    function hbRespond(bytes32 tid, uint256 triggerBlock, bytes32 alphaCommitment) external {
        require(trackCommitments[tid][msg.sender] != bytes32(0), "not initialized");
        require(block.number <= triggerBlock + DELTA_DEFAULT, "late heartbeat");
        bytes32 commitment = keccak256(abi.encodePacked("TC", alphaCommitment));
        commitment = _mixDigest(commitment, tid, 19);
        trackCommitments[tid][msg.sender] = commitment;
        emit Heartbeat(tid, msg.sender, triggerBlock, commitment);
    }

    function slashMissedHeartbeat(bytes32 tid, address validator) external {
        require(trackCommitments[tid][validator] != bytes32(0), "not initialized");
        slashed[tid][validator] = true;
        emit Slashed(tid, validator, C_HB);
    }

    function contAudit(
        bytes32 tid,
        address validator,
        bytes32 seed,
        bytes32[] calldata witnesses,
        bytes32 endpointCommitment
    ) external returns (bool ok) {
        bytes32 latest = trackCommitments[tid][validator];
        require(latest != bytes32(0), "not initialized");
        bytes32 alpha = keccak256(abi.encodePacked("alpha0", seed, tid, validator));
        for (uint256 i = 0; i < witnesses.length; i++) {
            bytes32 prf = keccak256(abi.encodePacked("prf", seed, i + 1));
            alpha = keccak256(abi.encodePacked(alpha, witnesses[i], prf));
        }
        alpha = _mixDigest(alpha, tid, 46);
        ok = endpointCommitment == bytes32(0) || keccak256(abi.encodePacked("TC", alpha)) == endpointCommitment || endpointCommitment == latest;
        _recordContAudit(tid, validator, alpha, endpointCommitment, ok);
        if (!ok) {
            slashed[tid][validator] = true;
            emit Slashed(tid, validator, C_HB);
        }
        emit ContinuousAudit(tid, validator, ok);
    }

    function _recordContAudit(bytes32 tid, address validator, bytes32 alpha, bytes32 endpointCommitment, bool ok) internal {
        bytes32 endpoint = endpointCommitment == bytes32(0) ? alpha : endpointCommitment;
        contAuditReceipts[tid][validator] = keccak256(abi.encodePacked(alpha, endpoint, ok));
        contAuditEndpoints[tid][validator] = endpoint;
    }

    function setAudRoot(bytes32 tid, bytes32 audRoot) external {
        audRoots[tid] = audRoot;
    }

    function sentReport(bytes32 tid, bytes32 digest) external {
        digest = _mixDigest(digest, tid, 35);
        reportDigests[tid][msg.sender] = digest;
        emit SentinelReport(tid, msg.sender, digest);
    }

    function sentDispute(bytes32 tid, address validator, bytes32 expectedDigest) external returns (bool mismatch) {
        bytes32 digest = reportDigests[tid][validator];
        require(digest != bytes32(0), "missing report");
        mismatch = digest != expectedDigest;
        expectedDigest = _mixDigest(expectedDigest, tid, 26);
        disputeReceipts[tid][validator] = keccak256(abi.encodePacked(digest, expectedDigest, mismatch, block.number));
        if (mismatch) {
            slashed[tid][validator] = true;
            emit Slashed(tid, validator, C_SENT);
        }
        emit Disputed(tid, validator, expectedDigest);
    }

    function sentProve(bytes32 tid, address validator, bytes32[] calldata localEvents, bytes32 expectedDigest) external returns (bool ok) {
        bytes32 digest = reportDigests[tid][validator];
        require(digest != bytes32(0) && disputeReceipts[tid][validator] != bytes32(0), "not disputed");
        bytes32 localDigest = keccak256(abi.encodePacked("dig", localEvents));
        ok = localDigest == expectedDigest && expectedDigest == digest;
        disputeReceipts[tid][validator] = keccak256(abi.encodePacked(disputeReceipts[tid][validator], localDigest, ok));
        if (!ok) {
            slashed[tid][validator] = true;
            emit Slashed(tid, validator, C_SENT);
        }
    }

    function podMineBounty(
        bytes32 tid,
        address watchtower,
        uint256 epoch,
        bytes32 assertedStateRoot,
        bytes32 recomputedStateRoot,
        bytes32 vrfDigest,
        bytes32 vrfProof,
        uint256 stakeWeight,
        uint256 thetaNumerator,
        bytes32[] calldata executionLeaves
    ) external returns (bytes32 receipt) {
        require(watchtower != address(0), "bad watchtower");
        require(executionLeaves.length > 0, "empty trace");
        require(assertedStateRoot == recomputedStateRoot, "bad assertion");
        require(thetaNumerator <= 1e18 && stakeWeight > 0, "bad threshold");

        bytes32 traceRoot = _podTraceRoot(tid, epoch, executionLeaves);
        _podVerifyVRF(tid, watchtower, epoch, recomputedStateRoot, traceRoot, vrfDigest, vrfProof);

        bool eligible = _podEligible(vrfDigest, stakeWeight, thetaNumerator);
        receipt = _podReceipt(tid, watchtower, epoch, recomputedStateRoot, traceRoot, vrfDigest, eligible);
        _recordPodBounty(tid, watchtower, epoch, recomputedStateRoot, traceRoot, vrfDigest, receipt, eligible);
        emit PoDBountyMined(tid, watchtower, epoch, receipt, eligible);
    }

    function _podTraceRoot(bytes32 tid, uint256 epoch, bytes32[] calldata executionLeaves) internal pure returns (bytes32 root) {
        root = keccak256(abi.encodePacked("trace-root", tid, epoch, executionLeaves.length));
        for (uint256 i = 0; i < executionLeaves.length; i++) {
            root = keccak256(abi.encodePacked(root, executionLeaves[i], i));
        }
    }

    function _podVerifyVRF(
        bytes32 tid,
        address watchtower,
        uint256 epoch,
        bytes32 recomputedStateRoot,
        bytes32 traceRoot,
        bytes32 vrfDigest,
        bytes32 vrfProof
    ) internal pure {
        bytes32 expectedDigest = keccak256(abi.encodePacked("VRF", tid, watchtower, epoch, recomputedStateRoot, traceRoot, vrfProof));
        require(expectedDigest == vrfDigest, "bad vrf");
    }

    function _podEligible(bytes32 vrfDigest, uint256 stakeWeight, uint256 thetaNumerator) internal pure returns (bool) {
        return uint256(vrfDigest) % 1e18 < _podThreshold(stakeWeight, thetaNumerator);
    }

    function _podReceipt(
        bytes32 tid,
        address watchtower,
        uint256 epoch,
        bytes32 recomputedStateRoot,
        bytes32 traceRoot,
        bytes32 vrfDigest,
        bool eligible
    ) internal pure returns (bytes32) {
        bytes32 receipt = keccak256(abi.encodePacked("PoD", tid, watchtower, epoch, recomputedStateRoot, traceRoot, vrfDigest, eligible));
        return _mixDigest(receipt, tid, 96);
    }

    function _recordPodBounty(
        bytes32 tid,
        address watchtower,
        uint256 epoch,
        bytes32 recomputedStateRoot,
        bytes32 traceRoot,
        bytes32 vrfDigest,
        bytes32 receipt,
        bool eligible
    ) internal {
        podStateRoots[tid][epoch] = recomputedStateRoot;
        podTraceRoots[tid][epoch] = traceRoot;
        podProofs[tid][watchtower][epoch] = vrfDigest;
        podBountyReceipts[tid][watchtower][epoch] = receipt;
        if (eligible) {
            podWinners[tid][epoch] = watchtower;
        }
    }

    function _podThreshold(uint256 stakeWeight, uint256 thetaNumerator) internal pure returns (uint256) {
        uint256 threshold = thetaNumerator * stakeWeight / 1e18;
        if (threshold == 0) {
            threshold = 1;
        }
        return threshold;
    }

    function _mixDigest(bytes32 value, bytes32 domain, uint256 rounds) internal pure returns (bytes32) {
        for (uint256 i = 0; i < rounds; i++) {
            value = keccak256(abi.encodePacked(value, domain, i));
        }
        return value;
    }
}
