import hre from "hardhat";
import { expect } from "chai";
import { Address, toHex } from "viem";
import { randomBytes } from "crypto";
import { whaleAddress } from "./addresses";

// Empty Account describes how transactions behave against an account with
// with no balance, no code, and no nonce.
//
// Accounts without code can be called with any data and value.
describe("Empty Account", function () {
  const emptyAddress: Address = toHex(randomBytes(20));

  // The definition of an empty account
  it("has no balance, no code, and zero nonce", async function () {
    const publicClient = await hre.viem.getPublicClient();

    const balance = await publicClient.getBalance({ address: emptyAddress });
    const code = await publicClient.getCode({ address: emptyAddress });
    const nonce = await publicClient.getTransactionCount({ address: emptyAddress });

    expect(balance).to.equal(0n);
    expect(code).to.be.undefined;
    expect(nonce).to.equal(0);
  });

  // An empty account can receive a 0 value transaction
  it("can be called with no data or value", async function () {
    const publicClient = await hre.viem.getPublicClient();
    const walletClient = await hre.viem.getWalletClient(whaleAddress);

    const txHash = await walletClient.sendTransaction({
      to: emptyAddress,
    });
    const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
    const tx = await publicClient.getTransaction({ hash: txHash });

    expect(txReceipt.status).to.equal("success");
    expect(txReceipt.gasUsed).to.equal(21000n);
    expect(tx.value).to.equal(0n);
    expect(tx.to).to.equal(emptyAddress);
    expect(tx.input).to.equal("0x");
  });

  // An empty account can receive a call with any data payload
  it("can be called with data", async function () {
    const publicClient = await hre.viem.getPublicClient();
    const walletClient = await hre.viem.getWalletClient(whaleAddress);

    // 16 bytes with 1 of them zero
    // 16 * 15 + 4 * 1 = 244 gas
    const calldata = "0x1eb478108900a0b492ef5dd03921d02d";

    const txHash = await walletClient.sendTransaction({
      to: emptyAddress,
      data: calldata,
    });
    const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
    const tx = await publicClient.getTransaction({ hash: txHash });

    expect(txReceipt.status).to.equal("success");
    // exact gas -- no memory expansion or op code charges
    expect(txReceipt.gasUsed).to.equal(21000n + 244n);
    expect(tx.value).to.equal(0n);
    expect(tx.to).to.equal(emptyAddress);
    expect(tx.input).to.equal(calldata);
  });

  // Transaction may create an account at no extra cost
  it("can be called with data", async function () {
    const publicClient = await hre.viem.getPublicClient();
    const walletClient = await hre.viem.getWalletClient(whaleAddress);

    // Use a random account to not violate the assumption that parent emptyAccount is empty
    const emptyAddress = toHex(randomBytes(20));

    const txHash = await walletClient.sendTransaction({
      to: emptyAddress,
      value: 1n,
    });
    const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
    const tx = await publicClient.getTransaction({ hash: txHash });

    expect(txReceipt.status).to.equal("success");
    // A transaction may create an account with no extra charge beyond 21000 instrinic
    expect(txReceipt.gasUsed).to.equal(21000n);
    expect(tx.value).to.equal(1n);
    expect(tx.to).to.equal(emptyAddress);

    // Verify account was committed & value transferred
    const balance = await publicClient.getBalance({ address: emptyAddress });
    expect(balance).to.equal(1n);
  });
});
