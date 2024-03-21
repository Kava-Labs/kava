pragma solidity ^0.8.0;

address constant PRECOMPILED_IBC_CONTRACT_ADDRESS = address(0x0300000000000000000000000000000000000002);

interface IBC {
    function transferKava(
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp,
        string calldata memo
    ) external;

    function transferCosmosDenom(
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata denom,
        uint256 amount,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp,
        string calldata memo
    ) external;

    function transferERC20(
        string calldata sourcePort,
        string calldata sourceChannel,
        uint256 amount,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp,
        string calldata memo,
        string calldata kavaERC20Address
    ) external;
}

contract ExampleIBC {
    // TODO(yevhenii):
    // - implement proper contract registration
    // - call IBC precompile methods without bypassing type-checking

    function transferERC20(
        string calldata sourcePort,
        string calldata sourceChannel,
        uint256 amount,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp,
        string calldata memo,
        string calldata kavaERC20Address
    ) external {
        IBC(PRECOMPILED_IBC_CONTRACT_ADDRESS).transferERC20(
            sourcePort,
            sourceChannel,
            amount,
            receiver,
            revisionNumber,
            revisionHeight,
            timeoutTimestamp,
            memo,
            kavaERC20Address
        );
    }

    function transferKavaCall(
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp,
        string calldata memo
    ) public payable {
        bytes memory input = abi.encodeWithSelector(
            IBC.transferKava.selector,
            sourcePort,
            sourceChannel,
            receiver,
            revisionNumber,
            revisionHeight,
            timeoutTimestamp,
            memo
        );
        (bool success, bytes memory output) = PRECOMPILED_IBC_CONTRACT_ADDRESS.call{value: msg.value}(input);
        require(success, string.concat("call to precompile failed, output: ", string(output)));
    }

    function transferCosmosDenomCall(
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata denom,
        uint256 amount,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp,
        string calldata memo
    ) public {
        bytes memory input = abi.encodeWithSelector(
            IBC.transferCosmosDenom.selector,
            sourcePort,
            sourceChannel,
            denom,
            amount,
            receiver,
            revisionNumber,
            revisionHeight,
            timeoutTimestamp,
            memo
        );
        (bool success, bytes memory output) = PRECOMPILED_IBC_CONTRACT_ADDRESS.call(input);
        require(success, string.concat("call to precompile failed, output: ", string(output)));
    }

    function transferERC20Call(
        string calldata sourcePort,
        string calldata sourceChannel,
        uint256 amount,
        string calldata receiver,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp,
        string calldata memo,
        string calldata kavaERC20Address
    ) external {
        bytes memory input = abi.encodeWithSelector(
            IBC.transferERC20.selector,
            sourcePort,
            sourceChannel,
            amount,
            receiver,
            revisionNumber,
            revisionHeight,
            timeoutTimestamp,
            memo,
            kavaERC20Address
        );
        (bool success, bytes memory output) = PRECOMPILED_IBC_CONTRACT_ADDRESS.call(input);
        require(success, string.concat("call to precompile failed, output: ", string(output)));
    }
}
