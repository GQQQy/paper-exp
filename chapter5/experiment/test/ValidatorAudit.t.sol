// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../src/ValidatorAudit.sol";

contract ValidatorAuditTest {
    ValidatorAudit internal audit;
    bytes32 internal tid = keccak256("chapter5-task");
    bytes32 internal seed = keccak256("seed");
    bytes32 internal nonce = keccak256("nonce");

    function setUp() public {
        audit = new ValidatorAudit();
    }

    function testTrackInitGas() public {
        audit.trackInit(tid, keccak256(abi.encodePacked(seed)), keccak256(abi.encodePacked(nonce)));
    }

    function testHBRespondGas() public {
        audit.trackInit(tid, keccak256(abi.encodePacked(seed)), keccak256(abi.encodePacked(nonce)));
        audit.hbRespond(tid, block.number, keccak256("alpha"));
    }

    function testContAuditGas() public {
        audit.trackInit(tid, keccak256(abi.encodePacked(seed)), keccak256(abi.encodePacked(nonce)));
        bytes32[] memory witnesses = new bytes32[](10);
        for (uint256 i = 0; i < 10; i++) {
            witnesses[i] = keccak256(abi.encodePacked("audit-witness", i));
        }
        audit.contAudit(tid, address(this), seed, witnesses, bytes32(0));
    }

    function testSentReportGas() public {
        audit.sentReport(tid, keccak256("dig"));
    }

    function testDisputeGas() public {
        audit.trackInit(tid, keccak256(abi.encodePacked(seed)), keccak256(abi.encodePacked(nonce)));
        audit.sentReport(tid, keccak256("bad"));
        audit.sentDispute(tid, address(this), keccak256("good"));
    }

    function testSentProveGas() public {
        audit.trackInit(tid, keccak256(abi.encodePacked(seed)), keccak256(abi.encodePacked(nonce)));
        bytes32[] memory events = new bytes32[](64);
        for (uint256 i = 0; i < events.length; i++) {
            events[i] = keccak256(abi.encodePacked("event", i));
        }
        bytes32 digest = keccak256(abi.encodePacked("dig", events));
        audit.sentReport(tid, digest);
        audit.sentDispute(tid, address(this), keccak256("other"));
        audit.sentProve(tid, address(this), events, digest);
    }

    function testPoDMineBountyGas() public {
        uint256 epoch = 7;
        bytes32 asserted = keccak256("asserted-state-root");
        bytes32[] memory leaves = new bytes32[](260);
        bytes32 traceRoot = keccak256(abi.encodePacked("trace-root", tid, epoch, leaves.length));
        for (uint256 i = 0; i < leaves.length; i++) {
            leaves[i] = keccak256(abi.encodePacked("execution-leaf", i));
            traceRoot = keccak256(abi.encodePacked(traceRoot, leaves[i], i));
        }
        bytes32 proof = keccak256("watchtower-vrf-proof");
        bytes32 digest = keccak256(abi.encodePacked("VRF", tid, address(this), epoch, asserted, traceRoot, proof));
        audit.podMineBounty(tid, address(this), epoch, asserted, asserted, digest, proof, 1e18, 9e17, leaves);
    }

    function testHeartbeatTriggerPure() public view {
        audit.heartbeatTriggered(keccak256("bh"), tid, address(this), 100);
    }

    function testFullAuditFlow() public {
        audit.trackInit(tid, keccak256(abi.encodePacked(seed)), keccak256(abi.encodePacked(nonce)));
        audit.hbRespond(tid, block.number, keccak256("alpha"));
        bytes32[] memory witnesses = new bytes32[](1);
        witnesses[0] = keccak256("audit-witness");
        audit.contAudit(tid, address(this), seed, witnesses, bytes32(0));
        bytes32 dig = keccak256("dig");
        audit.setAudRoot(tid, keccak256("aud-root"));
        audit.sentReport(tid, dig);
    }
}
