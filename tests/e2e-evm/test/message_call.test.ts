import hre from "hardhat";
import type { ArtifactsMap } from "hardhat/types/artifacts";
import type { PublicClient, WalletClient, GetContractReturnType } from "@nomicfoundation/hardhat-viem/types";
import { expect } from "chai";
import {
  Abi,
  AbiParameter,
  Address,
  Chain,
  ContractFunctionName,
  decodeAbiParameters,
  DecodeAbiParametersReturnType,
  decodeFunctionResult,
  encodeFunctionData,
  getAddress,
  Hex,
  isAddress,
  pad,
  parseAbiParameters,
  parseEventLogs,
  SendTransactionParameters,
  toHex,
  TransactionReceipt,
} from "viem";
import { whaleAddress } from "./addresses";
import { ExtractAbiEventNames } from "abitype";

const defaultGas = 25000n;
const contractCallerGas = defaultGas + 10000n;

describe("Message calls", () => {
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

  interface TestContext<Impl extends keyof ArtifactsMap> {
    lowLevelCaller: GetContractReturnType<ArtifactsMap["Caller"]["abi"]>;
    // Using a generic to enforce type safety on functions that use this Abi,
    // eg. decodeFunctionResult would have the functionName parameter type
    // checked.
    implementationAbi: ArtifactsMap[Impl]["abi"];
    implementationAddress: Address;
  }

  type CallType = "call" | "callcode" | "delegatecall" | "staticcall";

  /**
   * buildCallData creates the transaction data for the given message call type,
   * wrapping the function call data in the appropriate low level caller
   * function.
   */
  function buildCallData(
    type: CallType,
    implAddr: Address,
    data: Hex,
    callCodeValue?: bigint,
  ): SendTransactionParameters<Chain> {
    if (callCodeValue && type !== "callcode") {
      expect.fail("callCodeValue parameter is only valid for callcode");
    }

    switch (type) {
      case "call":
        // Direct call, just pass through the tx data
        return {
          account: walletClient.account,
          to: implAddr,
          gas: defaultGas,
          data,
        };

      case "callcode":
        if (callCodeValue) {
          // Custom value, use functionCallCodeWithValue that calls the
          // implementation contract with the given value that may be different
          // from the parent value.
          return {
            account: walletClient.account,
            to: lowLevelCaller.address,
            gas: contractCallerGas,
            data: encodeFunctionData({
              abi: lowLevelCaller.abi,
              functionName: "functionCallCodeWithValue",
              args: [implAddr, callCodeValue, data],
            }),
          };
        }

        // No custom value, use functionCallCode which just uses msg.value
        return {
          account: walletClient.account,
          to: lowLevelCaller.address,
          // Needs additional gas since it calls the external
          // functionCallCodeWithValue function.
          gas: contractCallerGas + 1_105n,
          data: encodeFunctionData({
            abi: lowLevelCaller.abi,
            functionName: "functionCallCode",
            args: [implAddr, data],
          }),
        };

      case "delegatecall":
        return {
          account: walletClient.account,
          to: lowLevelCaller.address,
          gas: contractCallerGas,
          data: encodeFunctionData({
            abi: lowLevelCaller.abi,
            functionName: "functionDelegateCall",
            args: [implAddr, data],
          }),
        };

      case "staticcall":
        return {
          account: walletClient.account,
          to: lowLevelCaller.address,
          gas: contractCallerGas * 100n,
          data: encodeFunctionData({
            abi: lowLevelCaller.abi,
            functionName: "functionStaticCall",
            args: [implAddr, data],
          }),
        };
    }
  }

  /**
   * getResponseFromReceiptLogs decodes the event log data from the given
   * transaction receipt and returns the decoded parameters.
   */
  function getResponseFromReceiptLogs<abi extends Abi, const params extends readonly AbiParameter[]>(
    abi: abi,
    eventName: ExtractAbiEventNames<abi>,
    params: params,
    txReceipt: TransactionReceipt,
  ): DecodeAbiParametersReturnType<params> {
    // This can also be scoped to specific eventNames, but it would be more
    // clear if this produces an eventName mismatch instead of silently
    // returning an empty array.
    const logs = parseEventLogs({
      abi: abi,
      logs: txReceipt.logs,
    });

    expect(logs.length).to.equal(1, "unexpected number of logs");
    const [log] = logs;

    if (log.eventName !== eventName) {
      expect.fail(`unexpected event name`);
    }

    return decodeAbiParameters(params, log.data);
  }

  function itHasCorrectMsgSender(getContext: () => TestContext<"ContextInspector">) {
    describe("msg.sender", () => {
      let ctx: TestContext<"ContextInspector">;

      before("Setup context", function () {
        ctx = getContext();
      });

      // senderTestCase is to validate the expected msg.sender for a given message call type.
      interface senderTestCase {
        name: string;
        type: CallType;
        wantChildMsgSender: (ctx: TestContext<"ContextInspector">, signer: Address) => Address;
      }

      const testCases: senderTestCase[] = [
        {
          name: "direct call is parent (signer)",
          type: "call",
          wantChildMsgSender: (_, signer) => signer,
        },
        {
          name: "callcode msg.sender is parent (caller)",
          type: "callcode",
          wantChildMsgSender: (ctx) => ctx.lowLevelCaller.address,
        },
        {
          name: "delegatecall propagates msg.sender (same as caller, parent signer)",
          type: "delegatecall",
          wantChildMsgSender: (_, signer) => signer,
        },
        {
          name: "staticcall msg.sender is parent (caller)",
          type: "staticcall",
          wantChildMsgSender: (ctx) => ctx.lowLevelCaller.address,
        },
      ];

      for (const tc of testCases) {
        it(tc.name, async function () {
          let functionName: ContractFunctionName<typeof ctx.implementationAbi> = "emitMsgSender";

          // Static Call cannot emit events so we use a different function that
          // just returns the msg.sender.
          if (tc.type === "staticcall") {
            functionName = "getMsgSender";
          }

          // ContextInspector function call data
          const baseData = encodeFunctionData({
            abi: ctx.implementationAbi,
            functionName,
            args: [],
          });

          // Modify the call data based on the test case type
          const txData = buildCallData(tc.type, ctx.implementationAddress, baseData);
          // No value for this test
          txData.value = 0n;

          if (tc.type === "staticcall") {
            // Check return value
            const returnedMsgSender = await publicClient.call(txData);
            if (returnedMsgSender.data === undefined) {
              expect.fail("call return data is undefined");
            }

            // Decode low level caller first since it is a byte array that
            // includes the offset, length, and data.
            const dataBytes = decodeFunctionResult({
              abi: ctx.lowLevelCaller.abi,
              functionName: "functionStaticCall",
              data: returnedMsgSender.data,
            });

            // Decode dataBytes as an address
            const address = decodeFunctionResult({
              abi: ctx.implementationAbi,
              functionName: "getMsgSender",
              data: dataBytes,
            });

            const expectedSender = tc.wantChildMsgSender(ctx, walletClient.account.address);
            expect(getAddress(address)).to.equal(getAddress(expectedSender), "unexpected msg.sender");

            // Skip the rest of the test since staticcall does not emit events.
            return;
          }

          await publicClient.call(txData);

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("success");

          const [receivedAddress] = getResponseFromReceiptLogs(
            ctx.implementationAbi,
            "MsgSender",
            parseAbiParameters("address"),
            txReceipt,
          );

          expect(isAddress(receivedAddress), "log.data should be an address").to.be.true;

          const expectedSender = tc.wantChildMsgSender(ctx, walletClient.account.address);
          expect(getAddress(receivedAddress)).to.equal(getAddress(expectedSender), "unexpected msg.sender");
        });
      }
    });
  }

  function itHasCorrectMsgValue(getContext: () => TestContext<"ContextInspector">) {
    describe("msg.value", () => {
      let ctx: TestContext<"ContextInspector">;

      before("Setup context", function () {
        ctx = getContext();
      });

      interface valueTestCase {
        name: string;
        type: CallType;
        // msg.value for the parent
        giveParentValue: bigint;
        // Call value for the child, only applicable for call, callcode
        giveChildCallValue?: bigint;
        wantRevertReason?: string;
        // Expected msg.value for the child
        wantChildMsgValue: (txValue: bigint) => bigint;
      }

      const testCases: valueTestCase[] = [
        {
          name: "direct call",
          type: "call",
          giveParentValue: 10n,
          wantChildMsgValue: (txValue) => txValue,
        },
        {
          name: "delegatecall propagates msg.value",
          type: "delegatecall",
          giveParentValue: 10n,
          wantChildMsgValue: (txValue) => txValue,
        },
        {
          name: "callcode with value == parent msg.value",
          type: "callcode",
          giveParentValue: 10n,
          wantChildMsgValue: (txValue) => txValue,
        },
        {
          name: "callcode with a value != parent msg.value",
          type: "callcode",
          // Transfers 10 signer -> caller contract
          giveParentValue: 10n,
          // Transfers 5 Caller -> implementation contract
          giveChildCallValue: 5n,
          wantChildMsgValue: () => 5n,
        },
        {
          name: "staticcall",
          type: "staticcall",
          giveParentValue: 10n,
          wantRevertReason: "non-payable function was called with value 10",
          wantChildMsgValue: () => 0n,
        },
      ];

      for (const tc of testCases) {
        it(tc.name, async function () {
          // Initial data of a emitMsgSender() call.
          const baseData = encodeFunctionData({
            abi: ctx.implementationAbi,
            functionName: "emitMsgValue",
            args: [],
          });

          // Modify the call data based on the test case type
          const txData = buildCallData(tc.type, ctx.implementationAddress, baseData, tc.giveChildCallValue);
          txData.value = tc.giveParentValue;

          if (!tc.wantRevertReason) {
            // This throws an error with revert reason to make it easier to debug
            // if the transaction fails.
            await publicClient.call(txData);
          } else {
            // rejectedWith is an string includes matcher
            await expect(publicClient.call(txData)).to.be.rejectedWith(tc.wantRevertReason);

            // Cannot include msg.value with static call as it changes state so
            // skip the rest of the test.
            return;
          }

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("success");

          const [emittedAmount] = getResponseFromReceiptLogs(
            ctx.implementationAbi,
            "MsgValue",
            parseAbiParameters("uint256"),
            txReceipt,
          );

          // Assert msg.value is as expected
          const expectedValue = tc.wantChildMsgValue(txData.value);
          expect(emittedAmount).to.equal(expectedValue, "unexpected msg.value");
        });
      }
    });
  }

  function itHasCorrectStorageLocation(getContext: () => TestContext<"StorageBasic">) {
    describe("storage location", () => {
      let ctx: TestContext<"StorageBasic">;

      before("Setup context", function () {
        ctx = getContext();
      });

      interface storageTestCase {
        name: string;
        callType: CallType;
        wantRevert?: boolean;
        wantStorageContract?: (ctx: TestContext<"StorageBasic">) => Address;
      }

      const testCases: storageTestCase[] = [
        {
          name: "call storage in implementation",
          callType: "call",
          wantStorageContract: (ctx) => ctx.implementationAddress,
        },
        {
          name: "callcode storage in caller",
          callType: "callcode",
          // Storage in caller contract
          wantStorageContract: (ctx) => ctx.lowLevelCaller.address,
        },
        {
          name: "delegatecall storage in caller",
          callType: "delegatecall",
          // Storage in caller contract
          wantStorageContract: (ctx) => ctx.lowLevelCaller.address,
        },
        {
          name: "staticcall storage not allowed",
          callType: "staticcall",
          wantRevert: true,
          // No expected storage due to revert
        },
      ];

      let giveStoreValue = 0n;

      for (const tc of testCases) {
        it(tc.name, async function () {
          // Increment storage value for a different value each test
          giveStoreValue++;

          const baseData = encodeFunctionData({
            abi: ctx.implementationAbi,
            functionName: "setStorageValue",
            args: [giveStoreValue],
          });

          const txData = buildCallData(tc.callType, ctx.implementationAddress, baseData);
          // Signer is the whale
          txData.account = whaleAddress;
          // Call gas + storage gas
          txData.gas = contractCallerGas + 20_2000n;

          if (!txData.to) {
            expect.fail("to field not set");
          }

          if (tc.wantRevert) {
            await expect(publicClient.call(txData)).to.be.rejected;

            // No actual transaction or storage to check! Skip the rest of the test.
            return;
          } else {
            // Throw revert errors if the transaction fails. Not using expect()
            // here since this still fails the test if the transaction fails and
            // does not mess up the formatting of the error message.
            await publicClient.call(txData);
          }

          if (!tc.wantStorageContract) {
            expect.fail("expected storage contract set");
          }

          const txHash = await walletClient.sendTransaction(txData);
          const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
          expect(txReceipt.status).to.equal("success");

          if (txData.gas) {
            expect(txReceipt.gasUsed < txData.gas, "gas to not be exhausted").to.be.true;
          }

          // Check which contract the storage was set on
          const storageContract = tc.wantStorageContract(ctx);
          const storageValue = await publicClient.getStorageAt({
            address: storageContract,
            slot: toHex(0),
          });

          const expectedStorage = pad(toHex(giveStoreValue));
          expect(storageValue).to.equal(expectedStorage, "unexpected storage value");
        });
      }
    });
  }

  // Mock contracts & precompiles need to implement these interfaces. These
  // ABIs will be used to test the message call behavior.
  const contextInspectorAbi = hre.artifacts.readArtifactSync("ContextInspector").abi;
  const storageBasicAbi = hre.artifacts.readArtifactSync("StorageBasic").abi;

  // Test context and storage for mock contracts as a baseline check
  describe("Mock", () => {
    let contextInspectorMock: GetContractReturnType<ArtifactsMap["ContextInspectorMock"]["abi"]>;
    let storageBasicMock: GetContractReturnType<ArtifactsMap["StorageBasicMock"]["abi"]>;

    before("deploy mock contracts", async function () {
      contextInspectorMock = await hre.viem.deployContract("ContextInspectorMock");
      storageBasicMock = await hre.viem.deployContract("StorageBasicMock");
    });

    itHasCorrectMsgSender(() => ({
      lowLevelCaller: lowLevelCaller,
      implementationAbi: contextInspectorAbi,
      implementationAddress: contextInspectorMock.address,
    }));

    itHasCorrectMsgValue(() => ({
      lowLevelCaller: lowLevelCaller,
      implementationAbi: contextInspectorAbi,
      implementationAddress: contextInspectorMock.address,
    }));

    itHasCorrectStorageLocation(() => ({
      lowLevelCaller: lowLevelCaller,
      implementationAbi: storageBasicAbi,
      implementationAddress: storageBasicMock.address,
    }));
  });

  // TODO: Test context and storage for precompiles
});
