import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.18",
    settings: {
      // istanbul upgrade occurred before the london hardfork, so is compatible with kava's evm
      evmVersion: "istanbul",
      // optimize build for deployment to mainnet!
      optimizer: {
        enabled: true,
        runs: 1000,
      },
      viaIR: true,
    },
  },
  networks: {
    // kvtool's local network
    kava: {
      url: "http://127.0.0.1:8545",
      accounts: [
        // kava keys unsafe-export-eth-key whale2
        "AA50F4C6C15190D9E18BF8B14FC09BFBA0E7306331A4F232D10A77C2879E7966",
      ],
    },
    protonet: {
      url: "https://evm.app.protonet.us-east.production.kava.io:443",
      accounts: [
        "247069F0BC3A5914CB2FD41E4133BBDAA6DBED9F47A01B9F110B5602C6E4CDD9",
      ],
    },
    internal_testnet: {
      url: "https://evm.data.internal.testnet.us-east.production.kava.io:443",
      accounts: [
        "247069F0BC3A5914CB2FD41E4133BBDAA6DBED9F47A01B9F110B5602C6E4CDD9",
      ],
    },
  },
};

export default config;
