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
        bytes32[] memory bhs = new bytes32[](10);
        bytes32[] memory taus = new bytes32[](10);
        uint256[] memory steps = new uint256[](10);
        for (uint256 i = 0; i < 10; i++) {
            bhs[i] = keccak256(abi.encodePacked("bh", i));
            taus[i] = keccak256(abi.encodePacked("tau", i));
            steps[i] = i + 1;
        }
        audit.contAudit(tid, address(this), seed, nonce, bhs, taus, steps, bytes32(0));
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

    function testPoDBaselineGas() public {
        audit.podBaseline(tid, address(this), 7, keccak256("watchtower"));
    }

    function testHeartbeatTriggerPure() public view {
        audit.heartbeatTriggered(keccak256("bh"), tid, address(this), 100);
    }

    function testFullAuditFlow() public {
        audit.trackInit(tid, keccak256(abi.encodePacked(seed)), keccak256(abi.encodePacked(nonce)));
        audit.hbRespond(tid, block.number, keccak256("alpha"));
        bytes32[] memory bhs = new bytes32[](1);
        bytes32[] memory taus = new bytes32[](1);
        uint256[] memory steps = new uint256[](1);
        bhs[0] = keccak256("bh");
        taus[0] = keccak256("tau");
        steps[0] = 1;
        audit.contAudit(tid, address(this), seed, nonce, bhs, taus, steps, bytes32(0));
        bytes32 dig = keccak256("dig");
        audit.setAudRoot(tid, keccak256("aud-root"));
        audit.sentReport(tid, dig);
    }
}
