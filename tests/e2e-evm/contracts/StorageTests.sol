// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/**
 * @title A basic contract with a storage value.
 * @notice This contract is used to test storage reads and writes, primarily
 *         for testing storage behavior in delegateCall.
 */
contract StorageBasic {
    uint256 public storageValue;

    function setStorageValue(uint256 value) external {
        storageValue = value;
    }
}
