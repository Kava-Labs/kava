import { expect } from "chai";
import { ethers } from "hardhat";
import {BaseContract, Contract} from "ethers";
import {HardhatEthersSigner} from "@nomicfoundation/hardhat-ethers/src/signers";

const PRECOMPILED_MUL3_CONTRACT_ADDRESS: string = "0x0300000000000000000000000000000000000001";

const mul3ContractABI: string[] = [
    "function calcMul3(uint256 a, uint256 b, uint256 c) external",
    "function getMul3() external view returns (uint256)",
];

describe("Testing precompiled mul3 contract", function() {
    it("Should properly calculate product of 3 numbers (direct call to precompiled contract)", async function () {
        const mul3Contract: Contract = new ethers.Contract(PRECOMPILED_MUL3_CONTRACT_ADDRESS, mul3ContractABI, ethers.provider);
        const signers: HardhatEthersSigner[] = await ethers.getSigners();
        const mul3ContractWithSigner: BaseContract = mul3Contract.connect(signers[0]);

        await mul3ContractWithSigner.calcMul3(2, 3, 4);
        await delay(2000);
        let actual: string = await mul3Contract.getMul3();
        let expected: string = "0x0000000000000000000000000000000000000000000000000000000000000018";
        expect(actual).to.equal(expected);

        await mul3ContractWithSigner.calcMul3(3, 5, 7);
        await delay(2000);
        actual = await mul3Contract.getMul3();
        expected = "0x0000000000000000000000000000000000000000000000000000000000000069"
        expect(actual).to.equal(expected);
    })
})

function delay(ms: number) {
    return new Promise( resolve => setTimeout(resolve, ms) );
}
