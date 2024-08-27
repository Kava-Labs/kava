import type { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox-viem";

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  networks: {
    kvtool: {
      url: "http://127.0.0.1:8545",
    },
  },
};

export default config;
