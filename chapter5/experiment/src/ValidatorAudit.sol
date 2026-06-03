// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract ValidatorAudit {
    struct TrackState {
        bytes32 tc0;
        bytes32 latestCommitment;
        uint256 latestBlock;
        bool initialized;
        bool slashed;
    }

    struct SentState {
        bytes32 digest;
        bool reported;
        bool disputed;
        bool proved;
    }

    uint256 public constant C_HB = 1;
    uint256 public constant C_SENT = 2;
    uint256 public constant M_DEFAULT = 100;
    uint256 public constant DELTA_DEFAULT = 7;

    mapping(bytes32 => mapping(address => TrackState)) public tracks;
    mapping(bytes32 => mapping(address => SentState)) public reports;
    mapping(bytes32 => bytes32) public audRoots;
    mapping(bytes32 => mapping(address => mapping(uint256 => bytes32))) public podProofs;

    event TrackInitialized(bytes32 indexed tid, address indexed validator, bytes32 tc0);
    event Heartbeat(bytes32 indexed tid, address indexed validator, uint256 blockNumber, bytes32 commitment);
    event ContinuousAudit(bytes32 indexed tid, address indexed validator, bool ok);
    event SentinelReport(bytes32 indexed tid, address indexed validator, bytes32 digest);
    event Disputed(bytes32 indexed tid, address indexed validator, bytes32 claim);
    event Slashed(bytes32 indexed tid, address indexed validator, uint256 amount);
    event PoDAttested(bytes32 indexed tid, address indexed validator, uint256 indexed epoch, bytes32 proof);

    function trackInit(bytes32 tid, bytes32 seedHash, bytes32 nonceHash) external returns (bytes32 tc0) {
        tc0 = keccak256(abi.encodePacked("TC0", tid, msg.sender, seedHash, nonceHash));
        tracks[tid][msg.sender] = TrackState({
            tc0: tc0,
            latestCommitment: tc0,
            latestBlock: block.number,
            initialized: true,
            slashed: false
        });
        emit TrackInitialized(tid, msg.sender, tc0);
    }

    function heartbeatTriggered(bytes32 bh, bytes32 tid, address validator, uint256 m) public pure returns (bool) {
        return uint256(keccak256(abi.encodePacked(bh, tid, validator))) % m == 0;
    }

    function hbRespond(bytes32 tid, uint256 triggerBlock, bytes32 alphaCommitment) external {
        TrackState storage state = tracks[tid][msg.sender];
        require(state.initialized, "not initialized");
        require(block.number <= triggerBlock + DELTA_DEFAULT, "late heartbeat");
        state.latestCommitment = keccak256(abi.encodePacked("TC", alphaCommitment));
        state.latestBlock = triggerBlock;
        emit Heartbeat(tid, msg.sender, triggerBlock, state.latestCommitment);
    }

    function slashMissedHeartbeat(bytes32 tid, address validator) external {
        TrackState storage state = tracks[tid][validator];
        require(state.initialized, "not initialized");
        state.slashed = true;
        emit Slashed(tid, validator, C_HB);
    }

    function contAudit(
        bytes32 tid,
        address validator,
        bytes32 seed,
        bytes32 nonce,
        bytes32[] calldata blockHashes,
        bytes32[] calldata taus,
        uint256[] calldata steps,
        bytes32 endpointCommitment
    ) external returns (bool ok) {
        TrackState storage state = tracks[tid][validator];
        require(state.initialized, "not initialized");
        require(blockHashes.length == taus.length && taus.length == steps.length, "length mismatch");
        bytes32 alpha = keccak256(abi.encodePacked("alpha0", seed, nonce, tid, validator));
        for (uint256 i = 0; i < steps.length; i++) {
            bytes32 prf = keccak256(abi.encodePacked("prf", seed, steps[i]));
            alpha = keccak256(abi.encodePacked(alpha, blockHashes[i], prf, taus[i]));
        }
        ok = keccak256(abi.encodePacked("TC", alpha)) == endpointCommitment || endpointCommitment == state.latestCommitment;
        if (!ok) {
            state.slashed = true;
            emit Slashed(tid, validator, C_HB);
        }
        emit ContinuousAudit(tid, validator, ok);
    }

    function setAudRoot(bytes32 tid, bytes32 audRoot) external {
        audRoots[tid] = audRoot;
    }

    function sentReport(bytes32 tid, bytes32 digest) external {
        reports[tid][msg.sender] = SentState({digest: digest, reported: true, disputed: false, proved: false});
        emit SentinelReport(tid, msg.sender, digest);
    }

    function sentDispute(bytes32 tid, address validator, bytes32 expectedDigest) external returns (bool mismatch) {
        SentState storage report = reports[tid][validator];
        require(report.reported, "missing report");
        report.disputed = true;
        mismatch = report.digest != expectedDigest;
        if (mismatch) {
            tracks[tid][validator].slashed = true;
            emit Slashed(tid, validator, C_SENT);
        }
        emit Disputed(tid, validator, expectedDigest);
    }

    function sentProve(bytes32 tid, address validator, bytes32[] calldata localEvents, bytes32 expectedDigest) external returns (bool ok) {
        SentState storage report = reports[tid][validator];
        require(report.reported && report.disputed, "not disputed");
        bytes32 digest = keccak256(abi.encodePacked("dig", localEvents));
        ok = digest == expectedDigest && expectedDigest == report.digest;
        report.proved = ok;
        if (!ok) {
            tracks[tid][validator].slashed = true;
            emit Slashed(tid, validator, C_SENT);
        }
    }

    function podBaseline(bytes32 tid, address validator, uint256 epoch, bytes32 watchtowerProof) external returns (bytes32) {
        require(validator != address(0), "bad validator");
        bytes32 acc = keccak256(abi.encodePacked(tid, validator, epoch, watchtowerProof));
        for (uint256 i = 0; i < 12; i++) {
            acc = keccak256(abi.encodePacked(acc, i));
        }
        podProofs[tid][validator][epoch] = acc;
        emit PoDAttested(tid, validator, epoch, acc);
        return acc;
    }
}
