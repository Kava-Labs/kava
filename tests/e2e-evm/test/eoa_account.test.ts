import hre from "hardhat";
import { expect } from "chai";
import { whaleAddress, userAddress } from "./addresses";

// EOA Account describes how transactions behave against an account with
// with no code while holding a balance.
//
// Accounts without code can be called with any data and value.
describe("EOA Account", function () {
  it("has a balance and no code", async function () {
    const publicClient = await hre.viem.getPublicClient();

    const balance = await publicClient.getBalance({ address: userAddress });
    const code = await publicClient.getCode({ address: userAddress });

    expect(balance).to.not.equal(0n);
    expect(code).to.be.undefined;
  });

  it("can be called with no data or value", async function () {
    const publicClient = await hre.viem.getPublicClient();
    const walletClient = await hre.viem.getWalletClient(whaleAddress);

    const txHash = await walletClient.sendTransaction({
      to: userAddress,
    });
    const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
    const tx = await publicClient.getTransaction({ hash: txHash });

    expect(txReceipt.status).to.equal("success");
    expect(txReceipt.gasUsed).to.equal(21000n);
    expect(tx.value).to.equal(0n);
    expect(tx.to).to.equal(userAddress);
    expect(tx.input).to.equal("0x");
  });

  it("can be called with data", async function () {
    const publicClient = await hre.viem.getPublicClient();
    const walletClient = await hre.viem.getWalletClient(whaleAddress);

    // 16 bytes with 1 of them zero
    // 16 * 15 + 4 * 1 = 244 gas
    const calldata = "0x1eb478108900a0b492ef5dd03921d02d";

    const txHash = await walletClient.sendTransaction({
      to: userAddress,
      data: calldata,
    });
    const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
    const tx = await publicClient.getTransaction({ hash: txHash });

    expect(txReceipt.status).to.equal("success");
    // exact gas -- no memory expansion or op code charges
    expect(txReceipt.gasUsed).to.equal(21000n + 244n);
    expect(tx.value).to.equal(0n);
    expect(tx.to).to.equal(userAddress);
    expect(tx.input).to.equal(calldata);
  });

  it("can be called with data and value", async function () {
    const publicClient = await hre.viem.getPublicClient();
    const walletClient = await hre.viem.getWalletClient(whaleAddress);

    // 4 non-zero bytes
    // 16 * 4 = 64 gas
    const calldata = "0x1eb47810";

    const txHash = await walletClient.sendTransaction({
      to: userAddress,
      data: calldata,
    });
    const txReceipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
    const tx = await publicClient.getTransaction({ hash: txHash });

    expect(txReceipt.status).to.equal("success");
    // exact gas -- no memory expansion or op code charges
    expect(txReceipt.gasUsed).to.equal(21000n + 64n);
    expect(tx.value).to.equal(0n);
    expect(tx.to).to.equal(userAddress);
    expect(tx.input).to.equal(calldata);
  });
});
