import { expect } from "chai";
import { Signer } from "ethers";
import { ethers } from "hardhat";
import {
  ERC20KavaWrappedCosmosCoin,
  ERC20KavaWrappedCosmosCoin__factory as ERC20KavaWrappedCosmosCoinFactory,
} from "../typechain-types";

const decimals = 6n;

describe("ERC20KavaWrappedCosmosCoin", function () {
  let erc20: ERC20KavaWrappedCosmosCoin;
  let erc20Factory: ERC20KavaWrappedCosmosCoinFactory;
  let owner: Signer;
  let sender: Signer;

  beforeEach(async function () {
    erc20Factory = await ethers.getContractFactory(
      "ERC20KavaWrappedCosmosCoin"
    );
    erc20 = await erc20Factory.deploy("Wrapped ATOM", "ATOM", decimals);
    [owner, sender] = await ethers.getSigners();
  });

  describe("decimals", function () {
    it("should be the same as deployed", async function () {
      expect(await erc20.decimals()).to.be.equal(decimals);
    });
  });

  describe("mint", function () {
    it("should reject non-owner", async function () {
      const tx = erc20.connect(sender).mint(await sender.getAddress(), 10n);
      await expect(tx).to.be.revertedWith("Ownable: caller is not the owner");
    });

    it("should be callable by owner", async function () {
      const amount = 10n;

      const tx = erc20.connect(owner).mint(await sender.getAddress(), amount);
      await expect(tx).to.not.be.reverted;

      const bal = await erc20.balanceOf(await sender.getAddress());
      expect(bal).to.equal(amount);
    });
  });

  describe("burn", function () {
    it("should reject non-owner", async function () {
      const tx = erc20.connect(sender).burn(await sender.getAddress(), 10n);
      await expect(tx).to.be.revertedWith("Ownable: caller is not the owner");
    });

    it("should let owner burn some of the tokens", async function () {
      const amount = 10n;

      // give tokens to an account
      const fundTx = erc20
        .connect(owner)
        .mint(await sender.getAddress(), amount);
      await expect(fundTx).to.not.be.reverted;
      const balBefore = await erc20.balanceOf(await sender.getAddress());
      expect(balBefore).to.equal(amount);

      // burn the tokens!
      const burnTx = erc20
        .connect(owner)
        .burn(await sender.getAddress(), amount);
      await expect(burnTx).to.not.be.reverted;
      const balAfter = await erc20.balanceOf(await sender.getAddress());
      expect(balAfter).to.equal(0);
    });

    it("should let owner burn all the tokens", async function () {
      const amount = 10n;

      // give tokens to an account
      const fundTx = erc20
        .connect(owner)
        .mint(await sender.getAddress(), amount);
      await expect(fundTx).to.not.be.reverted;
      const balBefore = await erc20.balanceOf(await sender.getAddress());
      expect(balBefore).to.equal(amount);

      // burn the tokens!
      const burnAmount = 7n;
      const burnTx = erc20
        .connect(owner)
        .burn(await sender.getAddress(), burnAmount);
      await expect(burnTx).to.not.be.reverted;
      const balAfter = await erc20.balanceOf(await sender.getAddress());
      expect(balAfter).to.equal(amount - burnAmount);
    });

    it("should error when trying to burn more than balance", async function () {
      const amount = 10n;

      // give tokens to an account
      const fundTx = erc20
        .connect(owner)
        .mint(await sender.getAddress(), amount);
      await expect(fundTx).to.not.be.reverted;
      const balBefore = await erc20.balanceOf(await sender.getAddress());
      expect(balBefore).to.equal(amount);

      // burn the tokens!
      const burnAmount = amount + 1n;
      const burnTx = erc20
        .connect(owner)
        .burn(await sender.getAddress(), burnAmount);
      await expect(burnTx).to.be.revertedWith(
        "ERC20: burn amount exceeds balance"
      );
    });
  });
});
