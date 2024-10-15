import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type { PublicClient, WalletClient, GetContractReturnType } from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import { Address, CallParameters, Chain, decodeFunctionResult, encodeFunctionData, getAddress } from "viem";
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

  interface callCodeTestCase {
    name: string;
    txParams: (ctx: CallCodeContext) => CallParameters<Chain>;
    wantSender: (ctx: CallCodeContext, signer: Address) => Address;
    wantValue: bigint;
  }

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
        // msg.sender is the caller contract, and value is NOT passed
        wantSender: (ctx) => ctx.lowLevelCaller.address,
        wantValue: 0n,
      },
    ];

    for (const tc of testCases) {
      it(tc.name, async function () {
        const txData = tc.txParams(ctx);
        txData.account = whaleAddress;

        const res = await publicClient.call(txData);
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
        expect(getAddress(returnAddress)).to.equal(getAddress(expectedSender));
        expect(returnValue).to.equal(tc.wantValue);

        const txHash = await walletClient.sendTransaction(txData);
        const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
        expect(txReceipt.status).to.equal("success");

        if (txData.gas) {
          expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
        }
      });
    }
  });
});
