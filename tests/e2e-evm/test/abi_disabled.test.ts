import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type { PublicClient, WalletClient, GetContractReturnType } from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import {
  Address,
  BaseError,
  Hex,
  toFunctionSelector,
  toFunctionSignature,
  concat,
  encodeFunctionData,
  isHex,
} from "viem";
import { Abi } from "abitype";
import { getAbiFallbackFunction, getAbiReceiveFunction } from "./helpers/abi";
import { whaleAddress } from "./addresses";

const defaultGas = 25000n;
const messageCallGas = defaultGas + 10000n;

interface ContractTestCase {
  interface: keyof ArtifactsMap;
  mock: keyof ArtifactsMap;
  precompile: Address;
}

const disabledPrecompiles: Address[] = [
  "0x9000000000000000000000000000000000000007", // noop receive and payable fallback
];

// ABI_DisabledTests assert ethereum + solidity transaction ABI interactions perform as expected.
describe("ABI_DisabledTests", function () {
  const testCases: ContractTestCase[] = [
    {
      interface: "NoopReceivePayableFallback", // interface to test valid function selectors against
      mock: "NoopDisabledMock", // mimics how a disabled precompile would behave
      precompile: disabledPrecompiles[0], // disabled noop receive and payable fallback
    },
  ];

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
    });
  }

  interface AbiContext {
    address: Address;
  }
  function itImplementsTheAbi(abi: Abi, getContext: () => Promise<AbiContext>) {
    let ctx: AbiContext;
    const receiveFunction = getAbiReceiveFunction(abi);
    const fallbackFunction = getAbiFallbackFunction(abi);

    before("Setup context", async function () {
      ctx = await getContext();
    });

    // testCase is a single test case with the value and data to send. Each of
    // these will be tested against each of the callCases.
    interface testCase {
      name: string;
      value: bigint;
      data: Hex;
    }
    const testCases: testCase[] = [];

    if (receiveFunction) {
      testCases.push(
        { name: "zero value transfer", value: 0n, data: "0x" },
        { name: "value transfer", value: 1n, data: "0x" },
      );
    }

    if (fallbackFunction) {
      testCases.push(
        { name: "invalid function selector", value: 0n, data: "0x010203" },
        { name: "invalid function selector with value", value: 1n, data: "0x010203" },
        { name: "non-matching function selector", value: 0n, data: toFunctionSelector("does_not_exist()") },
        { name: "non-matching function selector with value", value: 1n, data: toFunctionSelector("does_not_exist()") },
        {
          name: "non-matching function selector with extra data",
          value: 0n,
          data: concat([toFunctionSelector("does_not_exist()"), "0x01"]),
        },
        {
          name: "non-matching function selector with value and extra data",
          value: 1n,
          data: concat([toFunctionSelector("does_not_exist()"), "0x01"]),
        },
      );
    }

    for (const funcDesc of abi) {
      if (funcDesc.type !== "function") {
        continue;
      }

      const funcSelector = toFunctionSelector(toFunctionSignature(funcDesc));

      testCases.push(
        { name: funcDesc.name, value: 0n, data: funcSelector },
        { name: `${funcDesc.name} with value`, value: 1n, data: funcSelector },
        { name: `${funcDesc.name} with extra data`, value: 0n, data: concat([funcSelector, "0x01"]) },
        { name: `${funcDesc.name} with value and extra data`, value: 1n, data: concat([funcSelector, "0x01"]) },
      );
    }

    const callCases: {
      name: string;
      mutateData: (tc: Hex) => Hex;
      to: () => Address;
      gas: bigint;
    }[] = [
      {
        name: "external call",
        to: () => ctx.address,
        mutateData: (data) => data,
        gas: defaultGas,
      },
      {
        name: "message call",
        to: () => caller.address,
        mutateData: (data) =>
          encodeFunctionData({ abi: caller.abi, functionName: "functionCall", args: [ctx.address, data] }),
        gas: messageCallGas,
      },
      {
        name: "message delegatecall",
        to: () => caller.address,
        mutateData: (data) =>
          encodeFunctionData({ abi: caller.abi, functionName: "functionDelegateCall", args: [ctx.address, data] }),
        gas: messageCallGas,
      },
      {
        name: "message staticcall",
        to: () => caller.address,
        mutateData: (data) =>
          encodeFunctionData({ abi: caller.abi, functionName: "functionStaticCall", args: [ctx.address, data] }),
        gas: messageCallGas,
      },
    ];

    for (const tc of testCases) {
      describe(tc.name, function () {
        for (const cc of callCases) {
          it(`reverts on ${cc.name}`, async function () {
            const to = cc.to();
            const data = cc.mutateData(tc.data);

            const txData = { to: to, data: data, gas: cc.gas };
            const startingBalance = await publicClient.getBalance({ address: ctx.address });

            const call = publicClient.call(txData);
            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("reverted");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

            const balance = await publicClient.getBalance({ address: ctx.address });
            expect(balance).to.equal(startingBalance);

            let revertDetail = "";
            try {
              await call;
            } catch (e) {
              expect(e).to.be.instanceOf(BaseError, "expected error to be a BaseError");

              if (e instanceof BaseError) {
                revertDetail = e.details;
              }
            }

            const expectedMatch = "call not allowed to disabled contract";
            expect(revertDetail).to.equal(revertDetail, `expected ${revertDetail} to match ${expectedMatch}`);
          });
        }
      });
    }
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
        let deployedBytecode: Hex;

        before("deploy mock", async function () {
          const mockContractName: string = tc.mock;

          mockAddress = (await hre.viem.deployContract(mockContractName)).address;

          const mockArtifact = await hre.artifacts.readArtifact(mockContractName);
          if (isHex(mockArtifact.deployedBytecode)) {
            deployedBytecode = mockArtifact.deployedBytecode;
          } else {
            expect.fail("deployedBytecode is not hex");
          }
        });

        itHasCorrectState(() =>
          Promise.resolve({
            address: mockAddress,
            expectedCode: deployedBytecode,
            expectedNonce: 1,
          }),
        );

        itImplementsTheAbi(abi, () =>
          Promise.resolve({
            address: mockAddress,
          }),
        );
      });

      // The precompile tests only run on the kvtool network.  These tests ensure that the precompile
      // implemented the interface as expected and uses the same tests as the mocks above.
      (hre.network.name === "kvtool" ? describe : describe.skip)("Precompile", function () {
        const precompileAddress = tc.precompile;

        itHasCorrectState(() =>
          Promise.resolve({
            address: precompileAddress,
            expectedCode: "0x01",
            expectedNonce: 1,
          }),
        );

        itImplementsTheAbi(abi, () =>
          Promise.resolve({
            address: precompileAddress,
          }),
        );
      });
    });
  }
});
