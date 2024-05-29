pragma solidity ^0.8.0;

// Noop is an interface of precompile
interface Noop {
    // noop does nothing
    function noop() external view;
}

// NoopCaller is a contract which interacts with noop precompile either by using Noop interface or
// by using low-level calls like: call, static_call, etc...
// It's helpful in EOA -> NoopCaller -> Precompile test scenarios, meaning: EOA calls NoopCaller,
// NoopCaller calls precompile.
contract NoopCaller {
    event Event(string indexed indexedParam, string param);

    address private target;

    constructor(address _target) {
        target = _target;
    }

    // noop calls noop method of precompile
    function noop() external view {
        return Noop(target).noop();
    }

    // noop_static_call calls noop method of precompile by using static_call opcode
    function noop_static_call() external view returns (bytes memory) {
        bytes memory input = abi.encodeWithSelector(Noop.noop.selector);

        (bool ok, bytes memory data) = address(target).staticcall(input);
        require(ok, "call to precompiled contract failed");

        return data;
    }

    function emitEvent() external {
        emit Event("test-indexed-param", "test-param");
    }
}
