// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../src/BenchmarkTasks.sol";
import "../src/CleVerVerifier.sol";
import "../src/DisputeProtocolBenchmarks.sol";

contract BenchmarksTest {
    BenchmarkTasks internal tasks;
    CleVerVerifier internal verifier;
    DisputeProtocolBenchmarks internal disputes;

    function setUp() public {
        tasks = new BenchmarkTasks();
        verifier = new CleVerVerifier();
        disputes = new DisputeProtocolBenchmarks();
    }

    function testFibonacciGas() public view {
        tasks.fibonacci(128);
    }

    function testPolyChainGas() public view {
        tasks.polyChain(128, 7);
    }

    function testSortLargeGas() public view {
        tasks.sortLarge(24);
    }

    function testDPLargeGas() public view {
        tasks.dpLarge(512);
    }

    function testVerSegGas() public view {
        verifier.verSeg(bytes32(uint256(123)), 7, 256, 1, 2);
    }

    function testArbitrumOptimisticPathGas() public {
        disputes.arbitrumOptimisticPath(660);
    }

    function testTrueBitOptimisticPathGas() public {
        disputes.truebitOptimisticPath(742);
    }

    function testCartesiOptimisticPathGas() public {
        disputes.cartesiOptimisticPath(818);
    }

    function testBoLDOptimisticPathGas() public {
        disputes.boldOptimisticPath(722);
    }

    function testCleVerOptimisticPathGas() public {
        disputes.cleverOptimisticPath(1006);
    }

    function testArbitrumClassicDisputePathGas() public {
        disputes.arbitrumClassicPath(8001);
    }

    function testTrueBitDisputePathGas() public {
        disputes.truebitPath(4992);
    }

    function testCartesiDaveDisputePathGas() public {
        disputes.cartesiDavePath(4927, 633);
    }

    function testBoLDDisputePathGas() public {
        disputes.boldPath(1896, 3);
    }

    function testCleVerDisputePathGas() public {
        disputes.cleverPath(1135, 197);
    }
}
