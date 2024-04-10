pragma solidity ^0.8.0;

address constant PRECOMPILED_MUL3_CONTRACT_ADDRESS = address(0x0300000000000000000000000000000000000000);

// IMul3 is an interface of precompile
interface IMul3 {
    // calcMul3 calculates result and saves to EVM storage
    function calcMul3(uint256 a, uint256 b, uint256 c) external;
    // getMul3 gets previously stored result from storage
    function getMul3() external view returns (uint256 result);
}

// Mul3Caller is a contract which interacts with mul3 precompile either by using IMul3 interface or
// by using low-level calls like: call, static_call, etc...
// It's helpful in EOA -> Mul3Caller -> Precompile test scenarios, meaning: EOA calls Mul3Caller,
// Mul3Caller calls precompile.
contract Mul3Caller {
    // calcMul3 simply calls calcMul3 method of precompile
    function calcMul3(uint256 a, uint256 b, uint256 c) public {
        IMul3(PRECOMPILED_MUL3_CONTRACT_ADDRESS).calcMul3(a, b, c);
    }

    // getMul3 simply calls getMul3 method of precompile
    function getMul3() public view returns (uint256 result) {
        return IMul3(PRECOMPILED_MUL3_CONTRACT_ADDRESS).getMul3();
    }

    // calcMul3Call calls calcMul3 method of precompile by using call opcode
    function calcMul3Call(uint256 a, uint256 b, uint256 c) public returns (bytes memory) {
        bytes memory input = abi.encodeWithSelector(IMul3.calcMul3.selector, a, b, c);

        (bool ok, bytes memory data) = address(PRECOMPILED_MUL3_CONTRACT_ADDRESS).call(input);
        require(ok, "call to precompiled contract failed");

        return data;
    }

    // getMul3StaticCall calls getMul3 method of precompile by using static_call opcode
    function getMul3StaticCall() public view returns (bytes memory) {
        bytes memory input = abi.encodeWithSelector(IMul3.getMul3.selector);

        (bool ok, bytes memory data) = address(PRECOMPILED_MUL3_CONTRACT_ADDRESS).staticcall(input);
        require(ok, "call to precompiled contract failed");

        return data;
    }
}
