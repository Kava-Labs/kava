import { expect } from "chai";
import { ethers } from "hardhat";

describe("Testing precompiled stateless sum3 contract", function() {
    it("Should properly calculate sum of 3 numbers (call to example contract)", async function () {
        const ExampleSum3 = await ethers.getContractFactory("ExampleStatelessSum3");
        const exampleSum3 = await ExampleSum3.deploy();

        const actual: string = await exampleSum3.sum3(3, 4, 5);
        const expected: string = "0x000000000000000000000000000000000000000000000000000000000000000c";
        expect(actual).to.equal(expected);
    })
})
