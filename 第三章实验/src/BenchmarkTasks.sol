// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract BenchmarkTasks {
    function fibonacci(uint256 n) external pure returns (uint256) {
        uint256 a = 0;
        uint256 b = 1;
        for (uint256 i = 0; i < n; i++) {
            uint256 c = a + b;
            a = b;
            b = c;
        }
        return a;
    }

    function polyChain(uint256 n, uint256 x) external pure returns (uint256) {
        uint256 y = x;
        for (uint256 i = 0; i < n; i++) {
            y = addmod(mulmod(y, y + 17, type(uint256).max), 31 * i + 7, type(uint256).max);
        }
        return y;
    }

    function sortLarge(uint256 n) external pure returns (bytes32) {
        uint256[] memory values = new uint256[](n);
        for (uint256 i = 0; i < n; i++) {
            values[i] = uint256(keccak256(abi.encodePacked(i, n))) % 1_000_000;
        }
        for (uint256 width = 1; width < n; width *= 2) {
            for (uint256 left = 0; left < n; left += 2 * width) {
                uint256 mid = left + width;
                uint256 right = left + 2 * width;
                if (mid > n) mid = n;
                if (right > n) right = n;
                for (uint256 i = left; i < mid; i++) {
                    for (uint256 j = mid; j < right; j++) {
                        if (values[j] < values[i]) {
                            uint256 tmp = values[i];
                            values[i] = values[j];
                            values[j] = tmp;
                        }
                    }
                }
            }
        }
        return keccak256(abi.encode(values));
    }

    function dpLarge(uint256 n) external pure returns (uint256) {
        uint256[] memory dp = new uint256[](n + 2);
        dp[0] = 1;
        dp[1] = 1;
        for (uint256 i = 2; i <= n; i++) {
            dp[i] = addmod(dp[i - 1], dp[i - 2] + i, type(uint256).max);
        }
        return dp[n];
    }
}
