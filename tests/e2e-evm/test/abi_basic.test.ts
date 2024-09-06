import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type { PublicClient, WalletClient, ContractName } from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import { Address, toFunctionSelector, toFunctionSignature, concat } from "viem";
import { getAbiFallbackFunction, getAbiReceiveFunction } from "./helpers/abi";
import { whaleAddress } from "./addresses";

interface ContractTestCase {
  interface: keyof ArtifactsMap;
  mock: ContractName<keyof ArtifactsMap>;
}

// ABI_BasicTests assert ethereum + solidity transaction ABI interactions perform as expected.
describe("ABI_BasicTests", function () {
  const testCases = [
    // Test function modifiers without receive & fallback
    { interface: "NoopNoReceiveNoFallback", mock: "NoopNoReceiveNoFallbackMock" },

    // Test receive + fallback scenarios
    { interface: "NoopReceiveNoFallback", mock: "NoopReceiveNoFallbackMock" },
    { interface: "NoopReceivePayableFallback", mock: "NoopReceivePayableFallbackMock" },
    { interface: "NoopReceiveNonpayableFallback", mock: "NoopReceiveNonpayableFallbackMock" },

    // Test no receive + fallback scenarios
    { interface: "NoopNoReceivePayableFallback", mock: "NoopNoReceivePayableFallbackMock" },
    { interface: "NoopNoReceiveNonpayableFallback", mock: "NoopNoReceiveNonpayableFallbackMock" },
  ] as ContractTestCase[];

  //
  // Client + Wallet Setup
  //
  let publicClient: PublicClient;
  let walletClient: WalletClient;
  before("setup clients", async function () {
    publicClient = await hre.viem.getPublicClient();
    walletClient = await hre.viem.getWalletClient(whaleAddress);
  });

  //
  // Test each function defined in the interface ABI
  //
  // Only payable functions may be sent value
  // Any function (payable, non-payable, view, pure) can be called via a transaction
  // All functions can be provided calldata regardless of their arguments
  for (const tc of testCases) {
    describe(tc.interface, function () {
      const abi = hre.artifacts.readArtifactSync(tc.interface).abi;
      const receiveFunction = getAbiReceiveFunction(abi);
      const fallbackFunction = getAbiFallbackFunction(abi);

      let mockAddress: Address;
      before("deploy mock", async function () {
        mockAddress = (await hre.viem.deployContract(tc.mock)).address;
      });

      describe("State", function () {
        it("has code set", async function () {
          const code = await publicClient.getCode({ address: mockAddress });
          expect(code).to.not.equal(0);
        });

        it("has nonce default of 1", async function () {
          const nonce = await publicClient.getTransactionCount({ address: mockAddress });
          expect(nonce).to.equal(1);
        });

        it("has a starting balance of 0", async function () {
          const balance = await publicClient.getBalance({ address: mockAddress });
          expect(balance).to.equal(0n);
        });
      });

      for (const funcDesc of abi) {
        if (funcDesc.type !== "function") {
          continue;
        }

        const funcSelector = toFunctionSelector(toFunctionSignature(funcDesc));

        describe(`ABI function ${funcDesc.name} ${funcDesc.stateMutability}`, function () {
          const isPayable = funcDesc.stateMutability === "payable";

          it("can be called", async function () {
            const txData = { to: mockAddress, data: funcSelector, gas: 25000n };

            await expect(publicClient.call(txData)).to.be.fulfilled;

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("success");
          });

          it(`can ${isPayable ? "" : "not "}be called with value`, async function () {
            const txData = { to: mockAddress, data: funcSelector, gas: 25000n, value: 1n };

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal(isPayable ? "success" : "reverted");
          });

          it("can be called with extra data", async function () {
            const data = concat([funcSelector, "0x01"]);
            const txData = { to: mockAddress, data: data, gas: 25000n };

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("success");
          });
        });
      }

      //
      // Test ABI special functions -- receive & fallback
      //
      // Receive functions are always payable and can not receive data
      // Fallback functions can be payable or non-payable and can receive data in both cases
      const testName = `ABI special functions: ${receiveFunction ? "" : "no "}receive and ${fallbackFunction ? fallbackFunction.stateMutability : "no"} fallback`;
      describe(testName, function () {
        if (!receiveFunction && (!fallbackFunction || fallbackFunction.stateMutability !== "payable")) {
          it("can not receive plain transfers", async function () {
            const txData = { to: mockAddress, gas: 25000n, value: 1n };

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("reverted");
          });
        }

        if (receiveFunction || (fallbackFunction && fallbackFunction.stateMutability === "payable")) {
          it("can receive plain transfers", async function () {
            const txData = { to: mockAddress, gas: 25000n, value: 1n };

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal("success");
          });
        }

        it(`can ${fallbackFunction ? "" : "not "}be called with an invalid function selector`, async function () {
          const data = toFunctionSelector("does_not_exist()");
          const txData = { to: mockAddress, data: data, gas: 25000n };

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal(fallbackFunction ? "success" : "reverted");
        });

        if (fallbackFunction) {
          it(`can ${fallbackFunction.stateMutability === "payable" ? "" : "not "}receive transfer with data or invalid function selector`, async function () {
            const data = toFunctionSelector("does_not_exist()");
            const txData = { to: mockAddress, data: data, gas: 25000n, value: 1n };

            const txHash = await walletClient.sendTransaction(txData);
            const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
            expect(txReceipt.status).to.equal(fallbackFunction.stateMutability === "payable" ? "success" : "reverted");
          });
        }
      });
    });
  }
});
