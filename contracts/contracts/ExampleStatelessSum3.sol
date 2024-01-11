pragma solidity ^0.8.0;

import "hardhat/console.sol";

address constant PRECOMPILED_STATELESS_SUM3_CONTRACT_ADDRESS = address(0x0b);

contract ExampleStatelessSum3 {
    function sum3(uint256 a, uint256 b, uint256 c) public view returns (bytes memory) {
        (bool ok, bytes memory data) = address(PRECOMPILED_STATELESS_SUM3_CONTRACT_ADDRESS).staticcall(abi.encode(a,b,c));
        require(ok, "call to precompiled contract failed");

        return data;
    }
}
