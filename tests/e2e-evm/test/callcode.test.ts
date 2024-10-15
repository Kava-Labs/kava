import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type { PublicClient, WalletClient, GetContractReturnType } from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import { Address, Hex, decodeFunctionResult, encodeFunctionData, getAddress, isAddressEqual } from "viem";
import { whaleAddress } from "./addresses";

const defaultGas = 25000n;
const contractCallerGas = defaultGas + 10000n;

describe("CallCode", () => {
  //
  // Client + Wallet Setup
  //
  let publicClient: PublicClient;
  let walletClient: WalletClient;
  let lowLevelCaller: GetContractReturnType<ArtifactsMap["Caller"]["abi"]>;
  before("setup clients", async function () {
    publicClient = await hre.viem.getPublicClient();
    walletClient = await hre.viem.getWalletClient(whaleAddress);
    lowLevelCaller = await hre.viem.deployContract("Caller");
  });

  interface callCodeTestCase {
    name: string;
    fromAddress: Address;
    value: bigint;
    gas: bigint;
    mutateTxData?: (calledContractAddr: Address, data: Hex) => { to: Address; data: Hex };

    wantSender: () => Address;
    wantValue: bigint;
  }

  describe("msg context", () => {
    let calledContract: GetContractReturnType<ArtifactsMap["TargetContract"]["abi"]>;

    before("deploy called contract", async function () {
      const contract = await hre.viem.deployContract("TargetContract");

      calledContract = contract;
    });

    const testCases: callCodeTestCase[] = [
      {
        name: "direct call",
        fromAddress: whaleAddress,
        value: 0n,
        gas: defaultGas,
        wantSender: () => whaleAddress,
        wantValue: 0n,
      },
      {
        name: "direct call with value",
        fromAddress: whaleAddress,
        value: 100n,
        gas: defaultGas,
        wantSender: () => whaleAddress,
        wantValue: 100n,
      },
      {
        name: "callcode",
        fromAddress: whaleAddress,
        value: 0n,
        gas: contractCallerGas,
        mutateTxData: (contractAddr, data) => ({
          to: lowLevelCaller.address,
          data: encodeFunctionData({
            abi: lowLevelCaller.abi,
            functionName: "functionCallCode",
            args: [contractAddr, data],
          }),
        }),
        // msg.sender is the caller contract, not the whale
        wantSender: () => lowLevelCaller.address,
        wantValue: 0n,
      },
      {
        name: "callcode with value",
        fromAddress: whaleAddress,
        value: 100n,
        gas: contractCallerGas,
        mutateTxData: (contractAddr, data) => ({
          to: lowLevelCaller.address,
          data: encodeFunctionData({
            abi: lowLevelCaller.abi,
            functionName: "functionCallCode",
            args: [contractAddr, data],
          }),
        }),
        // msg.sender is the caller contract, and value is NOT passed
        wantSender: () => lowLevelCaller.address,
        wantValue: 0n,
      },
    ];

    for (const tc of testCases) {
      it(tc.name, async function () {
        let toAddr = calledContract.address;

        let callData = encodeFunctionData({
          abi: calledContract.abi,
          functionName: "getMsgInfo",
          args: [],
        });

        // Modify the call data if needed, wraps the call in a callcode
        if (tc.mutateTxData) {
          const { to, data } = tc.mutateTxData(calledContract.address, callData);
          callData = data;
          toAddr = to;
        }

        const txData = {
          // Transaction sender
          account: tc.fromAddress,
          // Target contract
          to: toAddr,
          data: callData,
          gas: tc.gas,
          value: tc.value,
        };

        const res = await publicClient.call(txData);
        if (!res.data) {
          // Fail this way as a type guard to ensure res.data is not undefined
          expect.fail("no data returned");
        }

        const [returnAddress, returnValue] = decodeFunctionResult({
          abi: calledContract.abi,
          functionName: "getMsgInfo",
          data: res.data,
        });

        // getAddress to ensure both are checksum encoded
        expect(getAddress(returnAddress)).to.equal(getAddress(tc.wantSender()));
        expect(returnValue).to.equal(tc.wantValue);

        const txHash = await walletClient.sendTransaction(txData);
        const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
        expect(txReceipt.status).to.equal("success");
        expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
      });
    }

    it("should be callable directly", async function () {
      const callData = encodeFunctionData({
        abi: calledContract.abi,
        functionName: "getMsgInfo",
        args: [],
      });

      const txData = {
        account: whaleAddress,
        to: calledContract.address,
        data: callData,
        gas: defaultGas,
      };

      const res = await publicClient.call(txData);
      if (!res.data) {
        // Fail this way as a type guard to ensure res.data is not undefined
        expect.fail("no data returned");
      }

      const [returnAddress, returnValue] = decodeFunctionResult({
        abi: calledContract.abi,
        functionName: "getMsgInfo",
        data: res.data,
      });

      expect(isAddressEqual(returnAddress, whaleAddress)).to.be.true;
      expect(returnValue).to.equal(0n);

      const txHash = await walletClient.sendTransaction(txData);
      const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
      expect(txReceipt.status).to.equal("success");
      expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
    });

    it("should be callable via callcode", async function () {
      const callData = encodeFunctionData({
        abi: calledContract.abi,
        functionName: "getMsgInfo",
        args: [],
      });

      const callerData: Hex = encodeFunctionData({
        abi: lowLevelCaller.abi,
        functionName: "functionCallCode",
        args: [calledContract.address, callData],
      });

      const txData = {
        to: lowLevelCaller.address,
        data: callerData,
        gas: contractCallerGas + 50_000n,
      };

      const res = await publicClient.call(txData);
      if (!res.data) {
        expect.fail("no data returned");
      }

      const [returnAddress, returnValue] = decodeFunctionResult({
        abi: calledContract.abi,
        functionName: "getMsgInfo",
        data: res.data,
      });

      expect(isAddressEqual(returnAddress, lowLevelCaller.address)).to.be.true;
      expect(returnValue).to.equal(0n);

      const txHash = await walletClient.sendTransaction(txData);
      const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
      expect(txReceipt.status).to.equal("success");
      expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
    });
  });
});
