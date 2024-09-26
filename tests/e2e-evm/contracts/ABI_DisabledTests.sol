// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./ABI_BasicTests.sol";

//
// Disabled contract that is payable and callable with any calldata (receive + fallback)
//
contract NoopDisabledMock is NoopReceivePayableFallback{
    // solc-ignore-next-line func-mutability
    function noopNonpayable() external {
      mockRevert();
    }
    function noopPayable() external payable {
      mockRevert();
    }
    // solc-ignore-next-line func-mutability
    function noopView() external view {
      mockRevert();
    }
    function noopPure() external pure {
      mockRevert();
    }
    receive() external payable {
      mockRevert();
    }
    fallback() external payable {
      mockRevert();
    }

    //
    // Mimic revert + revert reason
    //
    function mockRevert() private pure {
      revert("call not allowed to disabled contract");
    }
}
