// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../src/AnoStBenchmarks.sol";

contract AnoStBenchmarksTest {
    AnoStBenchmarks bench;
    bytes proof;

    constructor() {
        bench = new AnoStBenchmarks();
        proof = new bytes(256);
    }

    function testPublicReg() public {
        bench.publicReg(bytes32(uint256(1)));
    }

    function testPublicStake() public view {
        bench.publicStake(bytes32(uint256(2)));
    }

    function testCandidateDeclare() public view {
        bench.candidateDeclare(bytes32(uint256(3)), 32 ether);
    }

    function testAnonyReg() public {
        bench.anonyReg(bytes32(uint256(4)), bytes32(uint256(5)), bytes32(uint256(6)), proof);
    }

    function testAnonyStake() public {
        bench.anonyStake(bytes32(uint256(7)), bytes32(uint256(8)), bytes32(uint256(9)), proof);
    }

    function testPresentCred() public {
        bytes32[] memory path = new bytes32[](20);
        for (uint256 i = 0; i < path.length; i++) path[i] = bytes32(i + 10);
        bench.presentCred(bytes32(uint256(100)), path, 32 ether, proof);
    }

    function testElectCTWR() public view {
        uint256[] memory stakes = new uint256[](64);
        for (uint256 i = 0; i < stakes.length; i++) stakes[i] = 32 ether + (i % 8) * 8 ether;
        bench.electCTWR(stakes, 20, 256 ether, 42);
    }
}
