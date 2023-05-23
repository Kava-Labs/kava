import { ethers } from "hardhat";

async function main() {
  const tokenName = "Kava-wrapped ATOM";
  const tokenSymbol = "kATOM";
  const tokenDecimals = 6;

  const ERC20KavaWrappedCosmosCoin = await ethers.getContractFactory(
    "ERC20KavaWrappedCosmosCoin"
  );
  const token = await ERC20KavaWrappedCosmosCoin.deploy(
    tokenName,
    tokenSymbol,
    tokenDecimals
  );

  await token.deployed();

  console.log(
    `Token "${tokenName}" (${tokenSymbol}) with ${tokenDecimals} decimals is deployed to ${token.address}!`
  );
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
