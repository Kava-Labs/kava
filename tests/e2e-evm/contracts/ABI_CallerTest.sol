// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract TargetContract {
    function getMsgInfo() external payable returns (address, uint256) {
        return (msg.sender, msg.value);
    }
}
