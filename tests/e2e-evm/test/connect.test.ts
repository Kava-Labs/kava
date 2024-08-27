import hre from "hardhat";
import { expect } from "chai";

describe("Viem", function () {
  it("getChainID() matches the configured network", async function() {
    const publicClient = await hre.viem.getPublicClient();

    const expectedChainId = (() => {
      switch(hre.network.name) {
        case "hardhat":
          return 31337;
        case "kvtool":
          return 8888;
      }
    })();


    expect(await publicClient.getChainId()).to.eq(expectedChainId);
  });
});
