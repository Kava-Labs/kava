import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type {
  PublicClient,
  WalletClient,
  ContractName,
  GetContractReturnType,
} from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import { Address, BaseError, Hex, toFunctionSelector, toFunctionSignature, concat, encodeFunctionData } from "viem";
import { Abi } from "abitype";
import { getAbiFallbackFunction, getAbiReceiveFunction } from "./helpers/abi";
import { whaleAddress } from "./addresses";

const defaultGas = 25000n;
const messageCallGas = defaultGas + 10000n;

interface ContractTestCase {
  interface: keyof ArtifactsMap;
  mock: ContractName<keyof ArtifactsMap>;
  precompile: Address;
}

const disabledPrecompiles: Address[] = [
  "0x9000000000000000000000000000000000000007", // noop recieve and payable fallback
];

// ABI_DisabledTests assert ethereum + solidity transaction ABI interactions perform as expected.
describe("ABI_DisabledTests", function () {
  const testCases = [
    {
      interface: "NoopReceivePayableFallback", // interface to test valid function selectors against
      mock: "NoopDisabledMock", // mimics how a disabled precompile would behave
      precompile: disabledPrecompiles[0], // disabled noop recieve and payable fallback
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

    interface testCase {
      name: string
      value: bigint;
      data: Hex;
    }
    const testCases: testCase[] = [
      { name: "zero value transfer", value: 0n, data: "0x" },
      { name: "value transfer", value: 1n, data: "0x" },
      { name: "invalid function selector", value: 0n, data: "0x010203" },
      { name: "invalid function selector with value", value: 1n, data: "0x010203" },
      { name: "non-matching function selector", value: 0n, data: toFunctionSelector("does_not_exist()")},
      { name: "non-matching function selector with value", value: 1n, data: toFunctionSelector("does_not_exist()")},
      { name: "non-matching function selector with extra data", value: 0n, data: concat([toFunctionSelector("does_not_exist()"), "0x01"])},
      { name: "non-matching function selector with value and extra data", value: 1n, data: concat([toFunctionSelector("does_not_exist()"), "0x01"])},
    ];

    for (const funcDesc of abi) {
      if (funcDesc.type !== "function") continue;
      const funcSelector = toFunctionSelector(toFunctionSignature(funcDesc));
      testCases.concat([
        { name: funcDesc.name, value: 0n, data: funcSelector },
        { name: `${funcDesc.name} with value`, value: 1n, data: funcSelector },
        { name: `${funcDesc.name} with extra data`, value: 0n, data: concat([funcSelector, "0x01"]) },
        { name: `${funcDesc.name} with value and extra data`, value: 1n, data: concat([funcSelector, "0x01"]) },
      ]);
    }

    const callCases: {
      name: string;
      mutateData: (tc: Hex) => Promise<Hex>;
      to: () => Promise<Address>;
      gas: bigint;
    }[] = [
      {
        name: "external call",
        to: async () => ctx.address,
        mutateData: async (data) => data,
        gas: defaultGas,
      },
      {
        name: "message call",
        to: async () => caller.address,
        mutateData: async (data) => encodeFunctionData({ abi: caller.abi, functionName: "functionCall", args: [ctx.address, data] }),
        gas: messageCallGas,
      },
      {
        name: "message delegatecall",
        to: async () => caller.address,
        mutateData: async (data) => encodeFunctionData({ abi: caller.abi, functionName: "functionDelegateCall", args: [ctx.address, data] }),
        gas: messageCallGas,
      },
      {
        name: "message staticcall",
        to: async () => caller.address,
        mutateData: async (data) => encodeFunctionData({ abi: caller.abi, functionName: "functionStaticCall", args: [ctx.address, data] }),
        gas: messageCallGas,
      },
    ];

    for (const tc of testCases) {
      describe(tc.name, function() {
        for (const cc of callCases) {
          it(`reverts on ${cc.name}`, async function() {
            const to = await cc.to();
            const data = await cc.mutateData(tc.data);

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
            } catch(e) {
              if (e instanceof BaseError) {
                revertDetail = e.details;
              }
            }

            const expectedMatch = /call not allowed to disabled contract/;
            expect(expectedMatch.test(revertDetail), `expected ${revertDetail} to match ${expectedMatch}`).to.be.true;
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
          mockAddress = (await hre.viem.deployContract(tc.mock)).address;
          // TODO: fix typing and do not use explicit any, unsafe assignment, or unsafe access
          // eslint-disable-next-line  @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-member-access
          deployedBytecode = ((await hre.artifacts.readArtifact(tc.mock)) as any).deployedBytecode;
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
