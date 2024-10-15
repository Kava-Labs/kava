// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract TargetContract {
    uint256 public storageValue;

    function getMsgInfo() external payable returns (address, uint256) {
        return (msg.sender, msg.value);
    }

    function setStorageValue(uint256 value) external {
        storageValue = value;
    }
}
