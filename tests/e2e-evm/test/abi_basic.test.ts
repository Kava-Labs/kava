import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type {
  PublicClient,
  WalletClient,
  ContractName,
  GetContractReturnType,
} from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import { Address, Hex, toFunctionSelector, toFunctionSignature, concat, encodeFunctionData } from "viem";
import { Abi } from "abitype";
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
          const isPayable = funcDesc.stateMutability === "payable";

          it("can be called", async function () {
            const txData = { to: ctx.address, data: funcSelector, gas: defaultGas };

            await expect(publicClient.call(txData)).to.be.fulfilled;

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("success");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
          });

          it("can be called by low level contract call", async function () {
            const data = encodeFunctionData({
              abi: caller.abi,
              functionName: "functionCall",
              args: [ctx.address, funcSelector],
            });
            const txData = { to: caller.address, data: data, gas: contractCallerGas };

            await expect(publicClient.call(txData)).to.be.fulfilled;

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("success");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
          });

          if (funcDesc.stateMutability === "view" || funcDesc.stateMutability === "pure") {
            it("can be called by static call", async function () {
              const data = encodeFunctionData({
                abi: caller.abi,
                functionName: "functionStaticCall",
                args: [ctx.address, funcSelector],
              });
              const txData = { to: caller.address, data: data, gas: contractCallerGas };

              await expect(publicClient.call(txData)).to.be.fulfilled;

              const txHash = await walletClient.sendTransaction(txData);
              const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
              expect(txReceipt.status).to.equal("success");
              expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
            });

            it("can be called by static call with extra data", async function () {
              const data = encodeFunctionData({
                abi: caller.abi,
                functionName: "functionStaticCall",
                args: [ctx.address, concat([funcSelector, "0x01"])],
              });
              const txData = { to: caller.address, data: data, gas: contractCallerGas };

              await expect(publicClient.call(txData)).to.be.fulfilled;

              const txHash = await walletClient.sendTransaction(txData);
              const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
              expect(txReceipt.status).to.equal("success");
              expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
            });
          }

          it("can be called by high level contract call", async function () {
            const txData = { to: ctx.caller, data: funcSelector, gas: contractCallerGas };

            await expect(publicClient.call(txData)).to.be.fulfilled;

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("success");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
          });

          it(`can ${isPayable ? "" : "not "}be called with value`, async function () {
            const txData = { to: ctx.address, data: funcSelector, gas: defaultGas, value: 1n };
            const startingBalance = await publicClient.getBalance({ address: ctx.address });

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal(isPayable ? "success" : "reverted");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

            let expectedBalance = startingBalance;
            if (isPayable) {
              expectedBalance = startingBalance + txData.value;
            }
            const balance = await publicClient.getBalance({ address: ctx.address });
            expect(balance).to.equal(expectedBalance);
          });

          it(`can ${isPayable ? "" : "not "}be called by low level contract call with value`, async function () {
            const txData = { to: ctx.caller, data: funcSelector, gas: 50000n, value: 1n };
            const startingBalance = await publicClient.getBalance({ address: ctx.address });

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal(isPayable ? "success" : "reverted");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

            let expectedBalance = startingBalance;
            if (isPayable) {
              expectedBalance = startingBalance + txData.value;
            }
            const balance = await publicClient.getBalance({ address: ctx.address });
            expect(balance).to.equal(expectedBalance);
          });

          it(`can ${isPayable ? "" : "not "}be called by high level contract call with value`, async function () {
            const data = encodeFunctionData({
              abi: caller.abi,
              functionName: "functionCall",
              args: [ctx.address, funcSelector],
            });
            const txData = { to: caller.address, data: data, gas: contractCallerGas, value: 1n};
            const startingBalance = await publicClient.getBalance({ address: ctx.address });

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal(isPayable ? "success" : "reverted");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

            let expectedBalance = startingBalance;
            if (isPayable) {
              expectedBalance = startingBalance + txData.value;
            }
            const balance = await publicClient.getBalance({ address: ctx.address });
            expect(balance).to.equal(expectedBalance);
          });

          it("can be called with extra data", async function () {
            const data = concat([funcSelector, "0x01"]);
            const txData = { to: ctx.address, data: data, gas: defaultGas };

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("success");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
          });

          it(`can ${isPayable ? "" : "not "}be called with value and extra data`, async function () {
            const data = concat([funcSelector, "0x01"]);
            const txData = { to: ctx.address, data: data, gas: defaultGas, value: 1n };
            const startingBalance = await publicClient.getBalance({ address: ctx.address });

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal(isPayable ? "success" : "reverted");
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

            let expectedBalance = startingBalance;
            if (isPayable) {
              expectedBalance = startingBalance + txData.value;
            }
            const balance = await publicClient.getBalance({ address: ctx.address });
            expect(balance).to.equal(expectedBalance);
          });
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
      if (receiveFunction || fallbackFunction) {
        it("can receive zero value transfers with no data", async function () {
          const txData = { to: ctx.address, gas: defaultGas };

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("success");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
        });

        it("can be called by another contract with no data", async function() {
          const data = encodeFunctionData({
            abi: caller.abi,
            functionName: "functionCall",
            args: [ctx.address, "0x"],
          });
          const txData = { to: caller.address, data: data, gas: contractCallerGas };

          await expect(publicClient.call(txData)).to.be.fulfilled;

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("success");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
        });

        it("can be called by static call with no data", async function () {
          const data = encodeFunctionData({
            abi: caller.abi,
            functionName: "functionStaticCall",
            args: [ctx.address, "0x"],
          });
          const txData = { to: caller.address, data: data, gas: contractCallerGas };

          await expect(publicClient.call(txData)).to.be.fulfilled;

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("success");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
        });
      }

      if (!receiveFunction && !fallbackFunction) {
        it("can not receive zero value transfers with no data", async function () {
          const txData = { to: ctx.address, gas: defaultGas };

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("reverted");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
        });

        it("can not receive zero value transfers with no data", async function () {
          const data = encodeFunctionData({
            abi: caller.abi,
            functionName: "functionCall",
            args: [ctx.address, "0x"],
          });
          const txData = { to: caller.address, data: data, gas: contractCallerGas };

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("reverted");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
        });

        it("can not be called by static call with no data", async function () {
          const data = encodeFunctionData({
            abi: caller.abi,
            functionName: "functionStaticCall",
            args: [ctx.address, "0x"],
          });
          const txData = { to: caller.address, data: data, gas: contractCallerGas };

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("reverted");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
        });
      }

      if (!receiveFunction && (!fallbackFunction || fallbackFunction.stateMutability !== "payable")) {
        it("can not receive plain transfers", async function () {
          const txData = { to: ctx.address, gas: defaultGas, value: 1n };
          const startingBalance = await publicClient.getBalance({ address: ctx.address });

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("reverted");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

          let expectedBalance = startingBalance;
          if (fallbackFunction && fallbackFunction.stateMutability === "payable") {
            expectedBalance = startingBalance + txData.value;
          }
          const balance = await publicClient.getBalance({ address: ctx.address });
          expect(balance).to.equal(expectedBalance);
        });
      }

      if (receiveFunction || (fallbackFunction && fallbackFunction.stateMutability === "payable")) {
        it("can receive plain transfers", async function () {
          const txData = { to: ctx.address, gas: defaultGas, value: 1n };
          const startingBalance = await publicClient.getBalance({ address: ctx.address });

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("success");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

          const expectedBalance = startingBalance + txData.value;
          const balance = await publicClient.getBalance({ address: ctx.address });
          expect(balance).to.equal(expectedBalance);
        });
      }

      it(`can ${fallbackFunction ? "" : "not "}be called with a non-matching function selector`, async function () {
        const data = toFunctionSelector("does_not_exist()");
        const txData = { to: ctx.address, data: data, gas: defaultGas };

        const txHash = await walletClient.sendTransaction(txData);
        const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
        expect(txReceipt.status).to.equal(fallbackFunction ? "success" : "reverted");
        expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
      });

      it(`can ${fallbackFunction ? "" : "not "}be called with an invalid (short) function selector`, async function () {
        const data: Hex = "0x010203";
        const txData = { to: ctx.address, data: data, gas: defaultGas };

        const txHash = await walletClient.sendTransaction(txData);
        const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
        expect(txReceipt.status).to.equal(fallbackFunction ? "success" : "reverted");
        expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
      });

      if (fallbackFunction) {
        it(`can ${fallbackFunction.stateMutability === "payable" ? "" : "not "}receive value with a non-matching function selector`, async function () {
          const data = toFunctionSelector("does_not_exist()");
          const txData = { to: ctx.address, data: data, gas: defaultGas, value: 1n };
          const startingBalance = await publicClient.getBalance({ address: ctx.address });

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal(fallbackFunction.stateMutability === "payable" ? "success" : "reverted");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

          let expectedBalance = startingBalance;
          if (fallbackFunction.stateMutability === "payable") {
            expectedBalance = startingBalance + txData.value;
          }
          const balance = await publicClient.getBalance({ address: ctx.address });
          expect(balance).to.equal(expectedBalance);
        });

        it(`can ${fallbackFunction.stateMutability === "payable" ? "" : "not "}recieve value with an invalid function selector`, async function () {
          const data: Hex = "0x010203";
          const txData = { to: ctx.address, data: data, gas: defaultGas, value: 1n };
          const startingBalance = await publicClient.getBalance({ address: ctx.address });

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal(fallbackFunction.stateMutability === "payable" ? "success" : "reverted");
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;

          let expectedBalance = startingBalance;
          if (fallbackFunction.stateMutability === "payable") {
            expectedBalance = startingBalance + txData.value;
          }
          const balance = await publicClient.getBalance({ address: ctx.address });
          expect(balance).to.equal(expectedBalance);
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
