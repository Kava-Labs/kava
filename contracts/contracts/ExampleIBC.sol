pragma solidity ^0.8.0;

address constant PRECOMPILED_IBC_CONTRACT_ADDRESS = address(0x0300000000000000000000000000000000000002);

interface IBC {
    function ibcTransfer(
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata denom,
        uint256 amount,
        string calldata sender,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp
    ) external;
}

contract ExampleIBC {
    // TODO(yevhenii): uncomment and implement
    // function ibcTransfer() public {
    //     IBC(PRECOMPILED_IBC_CONTRACT_ADDRESS).ibcTransfer();
    // }

    function ibcTransferCall(
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata denom,
        uint256 amount,
        string calldata sender,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp
    ) public {
        bytes memory input = abi.encodeWithSelector(
            IBC.ibcTransfer.selector,
            sourcePort,
            sourceChannel,
            denom,
            amount,
            sender,
            receiver,
            revisionNumber,
            revisionHeight,
            timeoutTimestamp
        );
        (bool success, bytes memory output) = PRECOMPILED_IBC_CONTRACT_ADDRESS.call(input);
        require(success, string.concat("call to precompile failed, output: ", string(output)));
    }
}
