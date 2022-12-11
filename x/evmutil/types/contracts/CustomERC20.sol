// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract MyToken is ERC20, Ownable {
    constructor(string memory _name, string memory _symbol)
        ERC20(_name, _symbol)
    {}

    function deposit() public payable {
        require(msg.value > 0, "must deposit non-zero value");
        _mint(msg.sender, msg.value);
    }

    function triggerError() public pure returns (string memory) {
        require(false, "this function will always trigger an error");
        return "we should never get here";
    }

    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
    }
}
