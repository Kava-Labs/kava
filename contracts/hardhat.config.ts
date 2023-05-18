import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.18",
    settings: {
      evmVersion: "istanbul",
      optimizer: {
        enabled: true,
        runs: 1000,
      },
    },
  },
};

export default config;
