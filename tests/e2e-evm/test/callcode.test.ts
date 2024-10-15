import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type { PublicClient, WalletClient, GetContractReturnType } from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import { Address, CallParameters, Chain, decodeFunctionResult, encodeFunctionData, getAddress, pad, toHex } from "viem";
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

  interface CallCodeContext {
    lowLevelCaller: GetContractReturnType<ArtifactsMap["Caller"]["abi"]>;
    implementationContract: GetContractReturnType<ArtifactsMap["TargetContract"]["abi"]>;
  }

  interface callCodeTestCaseBase {
    name: string;
    txParams: (ctx: CallCodeContext) => CallParameters<Chain>;
  }

  // callCodeTestCaseSendAndValue is for getMsgInfo() call tests to validate
  // expected msg.sender and msg.value
  type callCodeTestCaseSendAndValue = callCodeTestCaseBase & {
    wantSender: (ctx: CallCodeContext, signer: Address) => Address;
    wantValue: bigint;
    wantStorageContract?: never;
    wantStorageValue?: never;
  };

  // callCodeTestCaseStorage is for setStorageValue() call tests to validate
  // which contract the storage is set on and the expected storage value
  type callCodeTestCaseStorage = callCodeTestCaseBase & {
    wantSender?: never;
    wantValue?: never;
    wantStorageContract: (ctx: CallCodeContext) => Address;
    wantStorageValue: bigint;
  };

  type callCodeTestCase = callCodeTestCaseSendAndValue | callCodeTestCaseStorage;

  describe("msg context", () => {
    let ctx: CallCodeContext;

    before("deploy called contract", async function () {
      const contract = await hre.viem.deployContract("TargetContract");
      ctx = {
        lowLevelCaller: lowLevelCaller,
        implementationContract: contract,
      };
    });

    const testCases: callCodeTestCase[] = [
      {
        name: "direct call",
        txParams: (ctx) => ({
          to: ctx.implementationContract.address,
          value: 0n,
          data: encodeFunctionData({
            abi: ctx.implementationContract.abi,
            functionName: "getMsgInfo",
            args: [],
          }),
          gas: defaultGas,
        }),
        wantSender: (_, signer) => signer,
        wantValue: 0n,
      },
      {
        name: "direct call with value",
        txParams: (ctx) => ({
          to: ctx.implementationContract.address,
          value: 100n,
          gas: defaultGas,
          data: encodeFunctionData({
            abi: ctx.implementationContract.abi,
            functionName: "getMsgInfo",
            args: [],
          }),
        }),
        wantSender: (_, signer) => signer,
        wantValue: 100n,
      },
      {
        name: "direct call with storage",
        txParams: (ctx) => ({
          to: ctx.implementationContract.address,
          value: 0n,
          gas: defaultGas + 20_000n,
          data: encodeFunctionData({
            abi: ctx.implementationContract.abi,
            functionName: "setStorageValue",
            args: [1n],
          }),
        }),
        wantStorageContract: (ctx) => ctx.implementationContract.address,
        wantStorageValue: 1n,
      },
      {
        name: "callcode",
        txParams: (ctx) => ({
          to: ctx.lowLevelCaller.address,
          value: 0n,
          gas: contractCallerGas,
          data: encodeFunctionData({
            abi: lowLevelCaller.abi,
            functionName: "functionCallCode",
            args: [
              ctx.implementationContract.address,
              encodeFunctionData({
                abi: ctx.implementationContract.abi,
                functionName: "getMsgInfo",
                args: [],
              }),
            ],
          }),
        }),
        // msg.sender is the caller contract, not the whale
        wantSender: (ctx) => ctx.lowLevelCaller.address,
        wantValue: 0n,
      },
      {
        name: "callcode with value",
        txParams: (ctx) => ({
          to: ctx.lowLevelCaller.address,
          value: 100n,
          gas: contractCallerGas,
          data: encodeFunctionData({
            abi: lowLevelCaller.abi,
            functionName: "functionCallCode",
            args: [
              ctx.implementationContract.address,
              encodeFunctionData({
                abi: ctx.implementationContract.abi,
                functionName: "getMsgInfo",
                args: [],
              }),
            ],
          }),
        }),
        // msg.sender is the caller contract
        wantSender: (ctx) => ctx.lowLevelCaller.address,
        // msg.value is not preserved
        wantValue: 0n,
      },
      {
        name: "callcode with storage",
        txParams: (ctx) => ({
          to: ctx.lowLevelCaller.address,
          value: 0n,
          gas: contractCallerGas + 20_000n,
          data: encodeFunctionData({
            abi: lowLevelCaller.abi,
            functionName: "functionCallCode",
            args: [
              ctx.implementationContract.address,
              encodeFunctionData({
                abi: ctx.implementationContract.abi,
                functionName: "setStorageValue",
                args: [2n],
              }),
            ],
          }),
        }),
        wantStorageContract: (ctx) => ctx.lowLevelCaller.address,
        wantStorageValue: 2n,
      },
    ];

    for (const tc of testCases) {
      it(tc.name, async function () {
        const txData = tc.txParams(ctx);
        // Signer is the whale
        txData.account = whaleAddress;

        const res = await publicClient.call(txData);

        // Check the return value for the msg.sender and msg.value if applicable
        if (tc.wantSender && tc.wantValue) {
          if (!res.data) {
            // Fail this way as a type guard to ensure res.data is not undefined
            expect.fail("no data returned");
          }

          // Expect all test cases call the getMsgInfo function
          const [returnAddress, returnValue] = decodeFunctionResult({
            abi: ctx.implementationContract.abi,
            functionName: "getMsgInfo",
            data: res.data,
          });

          const expectedSender = tc.wantSender(ctx, whaleAddress);

          // getAddress to ensure both are checksum encoded
          expect(getAddress(returnAddress)).to.equal(getAddress(expectedSender), "unexpected msg.sender");
          expect(returnValue).to.equal(tc.wantValue, "unexpected msg.value");
        }

        const txHash = await walletClient.sendTransaction(txData);
        const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
        expect(txReceipt.status).to.equal("success");

        if (txData.gas) {
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
        }

        // Storage tests if applicable
        if (tc.wantStorageContract) {
          const storageContract = tc.wantStorageContract(ctx);
          const storageValue = await publicClient.getStorageAt({
            address: storageContract,
            slot: toHex(0),
          });

          const expectedStorage = pad(toHex(tc.wantStorageValue));
          expect(storageValue).to.equal(expectedStorage, "unexpected storage value");
        }
      });
    }
  });
});
