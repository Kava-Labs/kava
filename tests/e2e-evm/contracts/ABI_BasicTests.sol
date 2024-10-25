// solhint-disable one-contract-per-file
// solhint-disable no-empty-blocks
// solhint-disable payable-fallback

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

//
// Low level caller
//
contract Caller {
    /**
     * @dev Call a function via CALL with the current msg.value
     */
    function functionCall(address payable to, bytes calldata data) external payable {
        this.functionCallWithValue(to, msg.value, data);
    }

    /**
     * @dev Call a function via CALL with a specific value that may be different
     *      from the current msg.value
     */
    function functionCallWithValue(address payable to, uint256 value, bytes calldata data) public payable {
        (bool success, bytes memory result) = to.call{value: value}(data);

        if (!success) {
            // solhint-disable-next-line gas-custom-errors
            if (result.length == 0) revert("reverted with no reason");

            // solhint-disable-next-line no-inline-assembly
            assembly {
                // Bubble up errors: revert(pointer, length of revert reason)
                // - result is a dynamic array, so the first 32 bytes is the
                //   length of the array.
                // - add(32, result) skips the length of the array and points to
                //   the start of the data.
                // - mload(result) reads 32 bytes from the memory location,
                //   which is the length of the revert reason.
                revert(add(32, result), mload(result))
            }
        }
    }

    /**
     * @dev Call a contract function via CALLCODE with the current msg.value
     */
    function functionCallCode(address to, bytes calldata data) external payable {
        this.functionCallCodeWithValue(to, msg.value, data);
    }

    /**
     * @dev Call a contract function via CALLCODE with a specific value that may
     *      be different from the current msg.value
     */
    function functionCallCodeWithValue(address to, uint256 value, bytes calldata data) external payable {
        // solhint-disable-next-line no-inline-assembly
        assembly {
            // Copy the calldata to memory, as callcode uses memory pointers.
            // offset is where the actual data starts in the calldata. Copy the
            // data to memory starting at 0.
            // Note: We are taking full control of memory as we do not return
            // to high-level Solidity code. This would be not memory safe as it
            // may exceed the scratch space in the first 64 bytes from 0.
            // This should be safe to still do, similar to the OpenZeppelin
            // proxy contract, overwriting the full memory scratch pad at
            // target 0 AND never returning to high-level Solidity code.
            calldatacopy(0, data.offset, data.length)

            // callcode(g, a, v, in, insize, out, outsize)
            // returns 0 on error (eg. out of gas) and 1 on success
            let result := callcode(
                gas(), // gas
                to, // to address
                value, // value to send
                0, // in - pointer to start of input, 0 since we copied the data to 0
                data.length, // insize - size of the input
                0, // out
                0 // outsize - 0 since we don't know the size of the output
            )

            // Copy the returned data to memory.
            // returndatacopy(t, f, s)
            // - t: target location in memory
            // - f: source location in return data
            // - s: size
            // Note: Same memory safety notes as above with calldatacopy()
            returndatacopy(0, 0, returndatasize())

            switch result
            // 0 on error
            case 0 {
                revert(0, returndatasize())
            }
            // 1 on success
            case 1 {
                return(0, returndatasize())
            }
            // Invalid result
            default {
                revert(0, 0)
            }
        }
    }

    /**
     * @dev Call a contract function via DELEGATECALL with the current msg.value
     *      and current msg.sender. DELEGATECALL cannot specify a different
     *      value.
     */
    function functionDelegateCall(address to, bytes calldata data) external payable {
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

    /**
     * @dev Call a contract function via STATICCALL with the current msg.value
     *      and current msg.sender.
     * @return The result of the static call in bytes.
     */
    function functionStaticCall(address to, bytes calldata data) external view returns (bytes memory) {
        (bool success, bytes memory result) = to.staticcall(data);

        if (!success) {
            // solhint-disable-next-line gas-custom-errors
            if (result.length == 0) revert("reverted with no reason");

            // solhint-disable-next-line no-inline-assembly
            assembly {
                revert(add(32, result), mload(result))
            }
        }

        return result;
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
