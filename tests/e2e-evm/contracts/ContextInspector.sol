// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/**
 * @title A contract to inspect the msg context.
 * @notice This contract is used to test the expected msg.sender and msg.value
 *         of a contract call in various scenarios.
 */
contract ContextInspector {
    /**
     * @dev Emitted when the emitMsgSender() function is called.
     */
    event MsgSender(address sender);

    /**
     * @dev Emitted when the emitMsgValue() function is called.
     *
     * Note that `value` may be zero.
     */
    event MsgValue(uint256 value);

    function emitMsgSender() external {
        emit MsgSender(msg.sender);
    }

    function emitMsgValue() external payable {
        emit MsgValue(msg.value);
    }

    /**
     * @dev Returns the current msg.sender. This is primarily used for testing
     *      staticcall as events are not emitted.
     */
    function getMsgSender() external view returns (address) {
        return msg.sender;
    }
}
