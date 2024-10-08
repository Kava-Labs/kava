import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type {
  PublicClient,
  WalletClient,
  ContractName,
  GetContractReturnType,
} from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import {
  Address,
  Hex,
  toFunctionSelector,
  toFunctionSignature,
  concat,
  encodeFunctionData,
  CallParameters,
  Chain,
} from "viem";
import { Abi, AbiFallback, AbiFunction, AbiReceive } from "abitype";
import { getAbiFallbackFunction, getAbiReceiveFunction } from "./helpers/abi";
import { whaleAddress } from "./addresses";

const defaultGas = 25000n;
const contractCallerGas = defaultGas + 10000n;

interface ContractTestCase {
  interface: keyof ArtifactsMap;
  mock: ContractName<keyof ArtifactsMap>;
  precompile: Address;
  caller: ContractName<keyof ArtifactsMap>;
}

const precompiles: Address[] = [
  "0x9000000000000000000000000000000000000001", // noop no recieve no fallback
  "0x9000000000000000000000000000000000000002", // noop receive no fallback
  "0x9000000000000000000000000000000000000003", // noop receive payable fallback
  "0x9000000000000000000000000000000000000004", // noop receive non payable fallback
  "0x9000000000000000000000000000000000000005", // noop no receive payable fallback
  "0x9000000000000000000000000000000000000006", // noop no recieve non payable fallback
];

// ABI_BasicTests assert ethereum + solidity transaction ABI interactions perform as expected.
describe("ABI_BasicTests", function () {
  const testCases = [
    // Test function modifiers without receive & fallback
    {
      interface: "NoopNoReceiveNoFallback",
      mock: "NoopNoReceiveNoFallbackMock",
      precompile: precompiles[0],
      caller: "NoopCaller",
    },

    // Test receive + fallback scenarios
    {
      interface: "NoopReceiveNoFallback",
      mock: "NoopReceiveNoFallbackMock",
      precompile: precompiles[1],
      caller: "NoopCaller",
    },
    {
      interface: "NoopReceivePayableFallback",
      mock: "NoopReceivePayableFallbackMock",
      precompile: precompiles[2],
      caller: "NoopCaller",
    },
    {
      interface: "NoopReceiveNonpayableFallback",
      mock: "NoopReceiveNonpayableFallbackMock",
      precompile: precompiles[3],
      caller: "NoopCaller",
    },

    // Test no receive + fallback scenarios
    {
      interface: "NoopNoReceivePayableFallback",
      mock: "NoopNoReceivePayableFallbackMock",
      precompile: precompiles[4],
      caller: "NoopCaller",
    },
    {
      interface: "NoopNoReceiveNonpayableFallback",
      mock: "NoopNoReceiveNonpayableFallbackMock",
      precompile: precompiles[5],
      caller: "NoopCaller",
    },
  ] as ContractTestCase[];

  //
  // Client + Wallet Setup
  //
  let publicClient: PublicClient;
  let walletClient: WalletClient;
  let caller: GetContractReturnType<ArtifactsMap["Caller"]["abi"]>;
  before("setup clients", async function () {
    publicClient = await hre.viem.getPublicClient();
    walletClient = await hre.viem.getWalletClient(whaleAddress);
    caller = await hre.viem.deployContract("Caller");
  });

  interface StateContext {
    address: Address;
    expectedCode: Hex;
    expectedNonce: number;
    expectedBalance: bigint;
  }
  function itHasCorrectState(getContext: () => Promise<StateContext>) {
    let ctx: StateContext;

    before("Setup context", async function () {
      ctx = await getContext();
    });

    describe("State", function () {
      it(`has correct code set in state`, async function () {
        const code = await publicClient.getCode({ address: ctx.address });
        expect(code).to.equal(ctx.expectedCode);
      });

      it("has default nonce set", async function () {
        const nonce = await publicClient.getTransactionCount({ address: ctx.address });
        expect(nonce).to.equal(ctx.expectedNonce);
      });

      it("has correct starting balance", async function () {
        const balance = await publicClient.getBalance({ address: ctx.address });
        expect(balance).to.equal(ctx.expectedBalance);
      });
    });
  }

  interface AbiContext {
    address: Address;
    caller: Address;
  }

  // ShouldRunFn is a function that determines if a test case should be run on a
  // specific function. This is useful if you only want to run a test case on
  // specific functions.
  type ShouldRunFn = (
    fn: AbiFunction,
    receiveFunction: AbiReceive | undefined,
    fallbackFunction: AbiFallback | undefined,
  ) => boolean;

  // AbiFunctionTestCaseBase defines a test case for an ABI function, this is run
  // on EVERY function defined in an expected ABI. Use shouldRun to only run the
  // test case on matching functions.
  interface AbiFunctionTestCaseBase {
    name: string;
    // If defined, only run this test case on functions that return true
    shouldRun?: ShouldRunFn;
    expectedStatus: "success" | "reverted";
    // If defined, check the balance of the contract after the transaction
    expectedBalance?: (startingBalance: bigint) => bigint;
  }

  // AbiFunctionTestCase is a test case for a specific function in an ABI that
  // includes the function selector.
  type AbiFunctionTestCase = AbiFunctionTestCaseBase & {
    txParams: (ctx: AbiContext, funcSelector: `0x${string}`) => CallParameters<Chain>;
  };

  // AbiFallbackTestCase is a test case for the fallback function, which does
  // not use a function selector.
  type AbiFallbackTestCase = AbiFunctionTestCaseBase & {
    // No function selector for fallback
    txParams: (ctx: AbiContext) => CallParameters<Chain>;
  };

  const abiFunctionTestCases: AbiFunctionTestCase[] = [
    {
      name: "can be called",
      txParams: (ctx, funcSelector) => ({ to: ctx.address, data: funcSelector, gas: defaultGas }),
      expectedStatus: "success",
    },
    {
      name: "can be called by low level contract call",
      txParams: (ctx, funcSelector) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, funcSelector],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can be called by static call",
      shouldRun(fn) {
        return fn.stateMutability === "view" || fn.stateMutability === "pure";
      },
      txParams: (ctx, funcSelector) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionStaticCall",
          args: [ctx.address, funcSelector],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can be called by static call with extra data",
      shouldRun(fn) {
        return fn.stateMutability === "view" || fn.stateMutability === "pure";
      },
      txParams: (ctx, funcSelector) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionStaticCall",
          args: [ctx.address, concat([funcSelector, "0x01"])],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can be called by high level contract call",
      txParams: (ctx, funcSelector) => ({
        to: ctx.caller,
        data: funcSelector,
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can be called with value",
      shouldRun(fn) {
        return fn.stateMutability === "payable";
      },
      txParams: (ctx, funcSelector) => ({
        to: ctx.address,
        data: funcSelector,
        gas: defaultGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance + 1n,
      expectedStatus: "success",
    },
    {
      name: "can not be called with value",
      shouldRun(fn) {
        return fn.stateMutability !== "payable";
      },
      txParams: (ctx, funcSelector) => ({
        to: ctx.address,
        data: funcSelector,
        gas: defaultGas,
        value: 1n,
      }),
      // Balance stays the same
      expectedBalance: (startingBalance) => startingBalance,
      expectedStatus: "reverted",
    },
    {
      name: "can be called by low level contract call with value",
      shouldRun(fn) {
        return fn.stateMutability === "payable";
      },
      txParams: (ctx, funcSelector) => ({
        to: ctx.caller,
        data: funcSelector,
        gas: 50000n,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance + 1n,
      expectedStatus: "success",
    },
    {
      name: "can not be called by low level contract call with value",
      shouldRun(fn) {
        return fn.stateMutability !== "payable";
      },
      txParams: (ctx, funcSelector) => ({
        to: ctx.caller,
        data: funcSelector,
        gas: 50000n,
        value: 1n,
      }),
      // Same
      expectedBalance: (startingBalance) => startingBalance,
      expectedStatus: "reverted",
    },
    {
      name: "can be called by high level contract call with value",
      shouldRun(fn) {
        return fn.stateMutability === "payable";
      },
      txParams: (ctx, funcSelector) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, funcSelector],
        }),
        gas: contractCallerGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance + 1n,
      expectedStatus: "success",
    },
    {
      name: "can not be called by high level contract call with value",
      shouldRun(fn) {
        return fn.stateMutability !== "payable";
      },
      txParams: (ctx, funcSelector) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, funcSelector],
        }),
        gas: contractCallerGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance,
      expectedStatus: "reverted",
    },
    {
      name: "can be called with extra data",
      txParams: (ctx, funcSelector) => ({
        to: ctx.address,
        data: concat([funcSelector, "0x01"]),
        gas: defaultGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can be called with value and extra data",
      shouldRun(fn) {
        return fn.stateMutability === "payable";
      },
      txParams: (ctx, funcSelector) => ({
        to: ctx.address,
        data: concat([funcSelector, "0x01"]),
        gas: defaultGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance + 1n,
      expectedStatus: "success",
    },
    {
      name: "can not be called with value and extra data",
      shouldRun(fn) {
        return fn.stateMutability !== "payable";
      },
      txParams: (ctx, funcSelector) => ({
        to: ctx.address,
        data: concat([funcSelector, "0x01"]),
        gas: defaultGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance,
      expectedStatus: "reverted",
    },
  ];

  // Test cases for special functions, receive and fallback
  const specialFunctionTests: AbiFallbackTestCase[] = [
    // Has receive function OR payable fallback
    {
      name: "can receive zero value transfers with no data",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!receiveFunction || !!fallbackFunction;
      },
      txParams: (ctx) => ({ to: ctx.address, gas: defaultGas }),
      expectedStatus: "success",
    },
    {
      name: "can be called by another contract with no data",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!receiveFunction || !!fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, "0x"],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can be called by static call with no data",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!receiveFunction || !!fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionStaticCall",
          args: [ctx.address, "0x"],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },

    // No receive function AND no payable fallback
    {
      name: "can not receive zero value transfers with no data",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !receiveFunction && !fallbackFunction;
      },
      txParams: (ctx) => ({ to: ctx.address, gas: defaultGas }),
      expectedStatus: "reverted",
    },
    {
      name: "can not receive zero value transfers by high level contract call with no data",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !receiveFunction && !fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, "0x"],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "reverted",
    },
    {
      name: "can not receive zero value transfers by static call with no data",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !receiveFunction && !fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionStaticCall",
          args: [ctx.address, "0x"],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "reverted",
    },

    // No receive function AND no payable fallback
    {
      name: "can not receive plain transfers",
      shouldRun(_, receiveFunction, fallbackFunction) {
        // No receive function and no payable fallback
        return !receiveFunction && (!fallbackFunction || fallbackFunction.stateMutability !== "payable");
      },
      txParams: (ctx) => ({ to: ctx.address, gas: defaultGas, value: 1n }),
      expectedStatus: "reverted",
    },
    {
      name: "can not receive plain transfers via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        // No receive function and no payable fallback
        return !receiveFunction && (!fallbackFunction || fallbackFunction.stateMutability !== "payable");
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, "0x"],
        }),
        gas: contractCallerGas,
        value: 1n,
      }),
      expectedStatus: "reverted",
    },
    // Has receive function OR payable fallback
    {
      name: "can receive plain transfers",
      shouldRun(_, receiveFunction, fallbackFunction) {
        // Has receive function OR payable fallback
        return !!receiveFunction || (!!fallbackFunction && fallbackFunction.stateMutability === "payable");
      },
      txParams: (ctx) => ({ to: ctx.address, gas: defaultGas, value: 1n }),
      expectedStatus: "success",
    },
    {
      name: "can receive plain transfers via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        // Has receive function OR payable fallback
        return !!receiveFunction || (!!fallbackFunction && fallbackFunction.stateMutability === "payable");
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, "0x"],
        }),
        gas: contractCallerGas,
        value: 1n,
      }),
      expectedStatus: "success",
    },

    {
      name: "can be called with a non-matching function selector",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction;
      },
      txParams: (ctx) => ({
        to: ctx.address,
        data: toFunctionSelector("does_not_exist()"),
        gas: defaultGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can not be called with a non-matching function selector",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !fallbackFunction;
      },
      txParams: (ctx) => ({
        to: ctx.address,
        data: toFunctionSelector("does_not_exist()"),
        gas: defaultGas,
      }),
      expectedStatus: "reverted",
    },

    {
      name: "can be called with a non-matching function selector via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, toFunctionSelector("does_not_exist()")],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can not be called with a non-matching function selector via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, toFunctionSelector("does_not_exist()")],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "reverted",
    },

    {
      name: "can be called with a non-matching function selector via static call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionStaticCall",
          args: [ctx.address, toFunctionSelector("does_not_exist()")],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can not be called with a non-matching function selector via static call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionStaticCall",
          args: [ctx.address, toFunctionSelector("does_not_exist()")],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "reverted",
    },

    {
      name: "can be called with an invalid (short) function selector",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction;
      },
      txParams: (ctx) => ({
        to: ctx.address,
        data: "0x010203",
        gas: defaultGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can not be called with an invalid (short) function selector",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !fallbackFunction;
      },
      txParams: (ctx) => ({
        to: ctx.address,
        data: "0x010203",
        gas: defaultGas,
      }),
      expectedStatus: "reverted",
    },

    {
      name: "can be called with an invalid (short) function selector via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, "0x010203"],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can not be called with an invalid (short) function selector via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, "0x010203"],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "reverted",
    },

    {
      name: "can be called with an invalid (short) function selector via static call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionStaticCall",
          args: [ctx.address, "0x010203"],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "success",
    },
    {
      name: "can not be called with an invalid (short) function selector via static call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !fallbackFunction;
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionStaticCall",
          args: [ctx.address, "0x010203"],
        }),
        gas: contractCallerGas,
      }),
      expectedStatus: "reverted",
    },

    // Fallback payable tests
    {
      name: "can receive value with a non-matching function selector",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction && fallbackFunction.stateMutability === "payable";
      },
      txParams: (ctx) => ({
        to: ctx.address,
        data: toFunctionSelector("does_not_exist()"),
        gas: defaultGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance + 1n,
      expectedStatus: "success",
    },
    {
      name: "can not receive value with a non-matching function selector",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction && fallbackFunction.stateMutability !== "payable";
      },
      txParams: (ctx) => ({
        to: ctx.address,
        data: toFunctionSelector("does_not_exist()"),
        gas: defaultGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance,
      expectedStatus: "reverted",
    },

    {
      name: "can receive value with a non-matching function selector via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction && fallbackFunction.stateMutability === "payable";
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, toFunctionSelector("does_not_exist()")],
        }),
        gas: contractCallerGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance + 1n,
      expectedStatus: "success",
    },
    {
      name: "can not receive value with a non-matching function selector via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction && fallbackFunction.stateMutability !== "payable";
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, toFunctionSelector("does_not_exist()")],
        }),
        gas: contractCallerGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance,
      expectedStatus: "reverted",
    },

    {
      name: "can receive value with an invalid (short) function selector",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction && fallbackFunction.stateMutability === "payable";
      },
      txParams: (ctx) => ({
        to: ctx.address,
        data: "0x010203",
        gas: defaultGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance + 1n,
      expectedStatus: "success",
    },
    {
      name: "can not receive value with an invalid (short) function selector",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction && fallbackFunction.stateMutability !== "payable";
      },
      txParams: (ctx) => ({
        to: ctx.address,
        data: "0x010203",
        gas: defaultGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance,
      expectedStatus: "reverted",
    },

    {
      name: "can receive value with an invalid (short) function selector via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction && fallbackFunction.stateMutability === "payable";
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, "0x010203"],
        }),
        gas: contractCallerGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance + 1n,
      expectedStatus: "success",
    },
    {
      name: "can not receive value with an invalid (short) function selector via message call",
      shouldRun(_, receiveFunction, fallbackFunction) {
        return !!fallbackFunction && fallbackFunction.stateMutability !== "payable";
      },
      txParams: (ctx) => ({
        to: caller.address,
        data: encodeFunctionData({
          abi: caller.abi,
          functionName: "functionCall",
          args: [ctx.address, "0x010203"],
        }),
        gas: contractCallerGas,
        value: 1n,
      }),
      expectedBalance: (startingBalance) => startingBalance,
      expectedStatus: "reverted",
    },
  ];

  function itImplementsTheAbi(abi: Abi, getContext: () => Promise<AbiContext>) {
    let ctx: AbiContext;
    const receiveFunction = getAbiReceiveFunction(abi);
    const fallbackFunction = getAbiFallbackFunction(abi);

    before("Setup context", async function () {
      ctx = await getContext();
    });

    describe("ABI functions", function () {
      for (const funcDesc of abi) {
        if (funcDesc.type !== "function") {
          continue;
        }

        const funcSelector = toFunctionSelector(toFunctionSignature(funcDesc));

        describe(`${funcDesc.name} ${funcDesc.stateMutability}`, function () {
          // Run test cases for each function
          for (const testCase of abiFunctionTestCases) {
            // Check if we should run this test case
            if (testCase.shouldRun && !testCase.shouldRun(funcDesc, receiveFunction, fallbackFunction)) {
              continue;
            }

            it(testCase.name, async function () {
              const startingBalance = await publicClient.getBalance({ address: ctx.address });

              const txData = testCase.txParams(ctx, funcSelector);

              if (testCase.expectedStatus === "success") {
                await expect(publicClient.call(txData)).to.be.fulfilled;
              } else {
                await expect(publicClient.call(txData)).to.be.rejected;
              }

              const txHash = await walletClient.sendTransaction(txData);
              const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
              expect(txReceipt.status).to.equal(testCase.expectedStatus);

              if (txData.gas) {
                expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
              }

              if (testCase.expectedBalance) {
                const expectedBalance = testCase.expectedBalance(startingBalance);

                const balance = await publicClient.getBalance({ address: ctx.address });
                expect(balance).to.equal(expectedBalance);
              } else {
                const balance = await publicClient.getBalance({ address: ctx.address });
                expect(balance).to.equal(startingBalance, "balance to not change if expectedBalance is not defined");
              }
            });
          }
        });
      }
    });

    //
    // Test ABI special functions -- receive & fallback
    //
    // Receive functions are always payable and can not receive data
    // Fallback functions can be payable or non-payable and can receive data in both cases
    const testName = `ABI special functions: ${receiveFunction ? "" : "no "}receive and ${fallbackFunction ? fallbackFunction.stateMutability : "no"} fallback`;
    describe(testName, function () {
      for (const testCase of specialFunctionTests) {
        it(testCase.name, async function () {
          const startingBalance = await publicClient.getBalance({ address: ctx.address });

          const txData = testCase.txParams(ctx);

          if (testCase.expectedStatus === "success") {
            await expect(publicClient.call(txData)).to.be.fulfilled;
          } else {
            await expect(publicClient.call(txData)).to.be.rejected;
          }

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal(testCase.expectedStatus);

          if (txData.gas) {
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
          }

          if (testCase.expectedBalance) {
            const expectedBalance = testCase.expectedBalance(startingBalance);

            const balance = await publicClient.getBalance({ address: ctx.address });
            expect(balance).to.equal(expectedBalance);
          } else {
            const balance = await publicClient.getBalance({ address: ctx.address });
            expect(balance).to.equal(startingBalance, "balance to not change if expectedBalance is not defined");
          }
        });
      }
    });
  }

  //
  // Test each function defined in the interface ABI
  //
  // Only payable functions may be sent value
  // Any function (payable, non-payable, view, pure) can be called via a transaction
  // All functions can be provided calldata regardless of their arguments
  for (const tc of testCases) {
    describe(tc.interface, function () {
      const abi = hre.artifacts.readArtifactSync(tc.interface).abi;

      // The interface is tested against a mock on all networks.
      // This serves as a reference to ensure all precompiles have mocks that behavior similarly
      // for testing on non-kava networks, in addition to ensure that we match normal contract
      // behavior as closely as possible.
      describe("Mock", function () {
        let mockAddress: Address;
        let callerAddress: Address;
        let deployedBytecode: Hex;

        before("deploy mock", async function () {
          mockAddress = (await hre.viem.deployContract(tc.mock)).address;
          callerAddress = (await hre.viem.deployContract(tc.caller, [mockAddress])).address;
          // TODO: fix typing and do not use explicit any, unsafe assignment, or unsafe access
          // eslint-disable-next-line  @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-member-access
          deployedBytecode = ((await hre.artifacts.readArtifact(tc.mock)) as any).deployedBytecode;
        });

        itHasCorrectState(() =>
          Promise.resolve({
            address: mockAddress,
            expectedCode: deployedBytecode,
            expectedNonce: 1,
            expectedBalance: 0n,
          }),
        );

        itImplementsTheAbi(abi, () =>
          Promise.resolve({
            address: mockAddress,
            caller: callerAddress,
          }),
        );
      });

      // The precompile tests only run on the kvtool network.  These tests ensure that the precompile
      // implemented the interface as expected and uses the same tests as the mocks above.
      (hre.network.name === "kvtool" ? describe : describe.skip)("Precompile", function () {
        const precompileAddress = tc.precompile;
        let callerAddress: Address;

        before("deploy pecompile caller", async function () {
          callerAddress = (await hre.viem.deployContract(tc.caller, [precompileAddress])).address;
        });

        itHasCorrectState(() =>
          Promise.resolve({
            address: precompileAddress,
            expectedCode: "0x01",
            expectedNonce: 1,
            expectedBalance: 0n,
          }),
        );

        itImplementsTheAbi(abi, () =>
          Promise.resolve({
            address: precompileAddress,
            caller: callerAddress,
          }),
        );
      });
    });
  }
});
