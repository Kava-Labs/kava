// SPDX-License-Identifier: MIT
pragma solidity ^0.8.18;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/// @title An ERC20 token contract owned and deployed by the evmutil module of Kava.
///        Tokens are backed one-for-one by cosmos-sdk coins held in the module account.
/// @author Kava Labs, LLC
/// @custom:security-contact security@kava.io
contract ERC20KavaWrappedCosmosCoin is ERC20, Ownable {
    /// @notice The decimals places of the token. For display purposes only.
    uint8 private immutable _decimals;

    /// @notice Registers the ERC20 token with mint and burn permissions for the
    ///         contract owner, by default the account that deploys this contract.
    /// @param name The name of the ERC20 token.
    /// @param symbol The symbol of the ERC20 token.
    /// @param decimals_ The number of decimals of the ERC20 token.
    constructor(
        string memory name,
        string memory symbol,
        uint8 decimals_
    ) ERC20(name, symbol) {
        _decimals = decimals_;
    }

    /// @notice Query the decimal places of the token for display purposes.
    function decimals() public view override returns (uint8) {
        return _decimals;
    }

    /// @notice Mints new tokens to an address. Can only be called by token owner.
    /// @param to Address to which new tokens are minted.
    /// @param amount Number of tokens to mint.
    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
    }

    /// @notice Burns tokens from an address. Can only be called by token owner.
    /// @param from Address from which tokens are burned.
    /// @param amount Number of tokens to burn.
    function burn(address from, uint256 amount) public onlyOwner {
        _burn(from, amount);
    }
}
