import { HardhatUserConfig, extendEnvironment } from "hardhat/config";
import "@nomicfoundation/hardhat-viem";
import { parseEther } from "viem";
import { extendViem } from "./test/extensions/viem";
import chai from "chai";
import chaiAsPromised from "chai-as-promised";
import "hardhat-ignore-warnings";

//
// Chai setup
//
chai.use(chaiAsPromised);

//
// Load HRE extensions
//
extendEnvironment(extendViem);

// These accounts are defined in the kvtool/config/common/addresses.json.
//
// The balances match the set balances in kvtool and these accounts are used for both the hardhat and kvtool networks.
// This ensures test consistency when running against either network.
const accounts = [
  // 0x03db6b11f47d074a532b9eb8a98ab7ada5845087
  // kava1q0dkky0505r555etn6u2nz4h4kjcg5y8dg863a
  //
  // jq -r '.kava.users.whale2.mnemonic' < tests/e2e/kvtool/config/common/addresses.json | kava keys add whale2 --eth --recover
  // kava keys unsafe-export-eth-key whale2
  {
    privateKey: "AA50F4C6C15190D9E18BF8B14FC09BFBA0E7306331A4F232D10A77C2879E7966",
    balance: parseEther("1000000").toString(),
  },

  // 0x7bbf300890857b8c241b219c6a489431669b3afa
  // kava10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t
  //
  // jq -r '.kava.users.user.mnemonic' < tests/e2e/kvtool/config/common/addresses.json | kava keys add user --eth --recover
  // kava keys unsafe-export-eth-key user
  {
    privateKey: "9549F115B0A21E5071A8AEC1B74AC093190E18DD83D019AC6497B0ADFBEFF26D",
    balance: parseEther("1000").toString(),
  },
];

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.24",
    settings: {
      evmVersion: "berlin", // The current hardfork of kava mainnet
      optimizer: {
        enabled: false, // The contracts here are only for testing and not deployment to mainnet.
      },
    },
  },
  networks: {
    hardhat: {
      chainId: 31337, // The default hardhat network chain id
      hardfork: "berlin", // The current hardfork of kava mainnet
      accounts: accounts,
      //
      // This is required for hardhat-viem to have the same behavior
      // for reverted transactions as on Kava.
      //
      throwOnTransactionFailures: false,
    },
    kvtool: {
      chainId: 8888, // The evm chain id of the kvtool network
      url: "http://127.0.0.1:8545",
      accounts: accounts.map((account) => account.privateKey),
    },
  },
};

export default config;
