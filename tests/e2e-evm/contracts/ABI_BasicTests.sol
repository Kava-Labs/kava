// solhint-disable one-contract-per-file
// solhint-disable no-empty-blocks
// solhint-disable payable-fallback

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

//
// Low level caller
//
contract Caller {
    function functionCall(address payable to, bytes calldata data) external payable {
        (bool success, bytes memory result) = to.call{value: msg.value}(data);

        if (!success) {
            // solhint-disable-next-line gas-custom-errors
            if (result.length == 0) revert("reverted with no reason");

            // solhint-disable-next-line no-inline-assembly
            assembly {
                revert(add(32, result), mload(result))
            }
        }
    }

    // TODO: Callcode

    function functionDelegateCall(address to, bytes calldata data) external {
        // solhint-disable-next-line avoid-low-level-calls
        (bool success, bytes memory result) = to.delegatecall(data);

        if (!success) {
            // solhint-disable-next-line gas-custom-errors
            if (result.length == 0) revert("reverted with no reason");

            // solhint-disable-next-line no-inline-assembly
            assembly {
                revert(add(32, result), mload(result))
            }
        }
    }

    function functionStaticCall(address to, bytes calldata data) external view {
        (bool success, bytes memory result) = to.staticcall(data);

        if (!success) {
            // solhint-disable-next-line gas-custom-errors
            if (result.length == 0) revert("reverted with no reason");

            // solhint-disable-next-line no-inline-assembly
            assembly {
                revert(add(32, result), mload(result))
            }
        }
    }
}

//
// High level caller
//
contract NoopCaller {
    NoopNoReceiveNoFallback private target;

    constructor(NoopNoReceiveNoFallback _target) {
        target = _target;
    }

    // call
    function noopNonpayable() external {
        target.noopNonpayable();
    }
    // call
    function noopPayable() external payable {
        target.noopPayable{value: msg.value}();
    }
    // staticcall (readonly)
    function noopView() external view {
        target.noopView();
    }
    // staticcall (readonly)
    function noopPure() external view {
        target.noopPure();
    }
}

//
// Normal noop functions only with nonpayable, payable, view, and pure modifiers
//
interface NoopNoReceiveNoFallback {
    function noopNonpayable() external;
    function noopPayable() external payable;
    function noopView() external view;
    function noopPure() external pure;
}
contract NoopNoReceiveNoFallbackMock is NoopNoReceiveNoFallback {
    function noopNonpayable() external {}
    function noopPayable() external payable {}
    function noopView() external view {}
    function noopPure() external pure {}
}

//
// Added receive function (always payable)
//
interface NoopReceiveNoFallback is NoopNoReceiveNoFallback {
    receive() external payable;
}
contract NoopReceiveNoFallbackMock is NoopReceiveNoFallback, NoopNoReceiveNoFallbackMock {
    receive() external payable {}
}

//
// Added receive function and payable fallback
//
interface NoopReceivePayableFallback is NoopNoReceiveNoFallback {
    receive() external payable;
    fallback() external payable;
}
contract NoopReceivePayableFallbackMock is NoopReceivePayableFallback, NoopNoReceiveNoFallbackMock {
    receive() external payable {}
    fallback() external payable {}
}

//
// Added receive function and non-payable fallback
//
interface NoopReceiveNonpayableFallback is NoopNoReceiveNoFallback {
    receive() external payable;
    fallback() external;
}
contract NoopReceiveNonpayableFallbackMock is NoopReceiveNonpayableFallback, NoopNoReceiveNoFallbackMock {
    receive() external payable {}
    fallback() external {}
}

//
// Added payable fallback and no receive function
//
// solc-ignore-next-line missing-receive
interface NoopNoReceivePayableFallback is NoopNoReceiveNoFallback {
    fallback() external payable;
}
// solc-ignore-next-line missing-receive
contract NoopNoReceivePayableFallbackMock is NoopNoReceivePayableFallback, NoopNoReceiveNoFallbackMock {
    fallback() external payable {}
}

//
// Added non-payable fallback and no receive function
//
interface NoopNoReceiveNonpayableFallback is NoopNoReceiveNoFallback {
    fallback() external;
}
contract NoopNoReceiveNonpayableFallbackMock is NoopNoReceiveNonpayableFallback, NoopNoReceiveNoFallbackMock {
    fallback() external {}
}
