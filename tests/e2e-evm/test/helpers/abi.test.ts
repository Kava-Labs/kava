import { expect } from "chai";
import { Abi, AbiFallback, AbiFunction, AbiReceive } from "abitype";
import { getAbiFallbackFunction, getAbiReceiveFunction } from "./abi";

const fallback: AbiFallback = { type: "fallback", stateMutability: "payable" };
const receive: AbiReceive = { type: "receive", stateMutability: "payable" };
const function1: AbiFunction = {
  type: "function",
  name: "function1",
  inputs: [],
  outputs: [],
  stateMutability: "nonpayable",
};
const function2: AbiFunction = {
  type: "function",
  name: "function2",
  inputs: [],
  outputs: [],
  stateMutability: "payable",
};

describe("ABI Helpers", function () {
  describe("getAbiFallbackFunction", function () {
    const testCases: {
      name: string;
      abi: Abi;
      expectedResult: AbiFallback | undefined;
    }[] = [
      { name: "returns undefined with an empty abi", abi: [], expectedResult: undefined },
      { name: "returns undefined with no fallback", abi: [receive, function1, function2], expectedResult: undefined },
      {
        name: "returns the fallback function when in abi",
        abi: [receive, function1, fallback, function2],
        expectedResult: fallback,
      },
      { name: "returns the fallback when it is the only function", abi: [fallback], expectedResult: fallback },
    ];

    for (const tc of testCases) {
      it(tc.name, function () {
        expect(getAbiFallbackFunction(tc.abi)).to.equal(tc.expectedResult);
      });
    }
  });

  describe("getAbiFallbackFunction", function () {
    const testCases: {
      name: string;
      abi: Abi;
      expectedResult: AbiReceive | undefined;
    }[] = [
      { name: "returns undefined with an empty abi", abi: [], expectedResult: undefined },
      { name: "returns undefined with no fallback", abi: [fallback, function1, function2], expectedResult: undefined },
      { name: "returns receive when in abi", abi: [function1, receive, fallback, function2], expectedResult: receive },
      { name: "returns receive when it is the only function", abi: [receive], expectedResult: receive },
    ];

    for (const tc of testCases) {
      it(tc.name, function () {
        expect(getAbiReceiveFunction(tc.abi)).to.equal(tc.expectedResult);
      });
    }
  });
});
