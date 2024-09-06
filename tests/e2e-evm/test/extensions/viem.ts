import { Address, defineChain, Chain, PublicClientConfig, WalletClientConfig } from "viem";
import { DeployContractConfig, ContractName } from "@nomicfoundation/hardhat-viem/types";
import { HardhatRuntimeEnvironment } from "hardhat/types/runtime";

// defaultPublicClientConfig sets default values for viem public client configuration
const defaultPublicClientConfig: Partial<PublicClientConfig> = {
  pollingInterval: 100,
};

// kavalocalnet defines the viem Chain configuration for a kvtool chain
const kavalocalnet: Chain = defineChain({
  id: 8888,
  name: "Kava EVM Localnet",
  network: "kava-localnet",
  nativeCurrency: {
    name: "Kava",
    decimals: 18,
    symbol: "KAVA",
  },
  rpcUrls: {
    default: {
      http: ["http://127.0.0.1:8545"],
      webSocket: ["ws://127.0.0.1:8546"],
    },
  },
  blockExplorers: undefined,
  contracts: {},
});

// getChainConfig returns the kvtoollocalnet Chain if the hardhat environment kvtool network is set, else undefined
function getChainConfig(hre: HardhatRuntimeEnvironment): { chain?: Chain } {
  if (hre.network.name === "kvtool") {
    return { chain: kavalocalnet };
  }

  return {};
}

// extendViem wraps the viem hardhat runtime environment in order to support kvtool chain configuration
export function extendViem(hre: HardhatRuntimeEnvironment) {
  /* eslint-disable @typescript-eslint/unbound-method */
  const { getPublicClient, getWalletClients, getWalletClient, deployContract } = hre.viem;
  /* eslint-enable @typescript-eslint/unbound-method */

  hre.viem.getPublicClient = (publicClientConfig?: Partial<PublicClientConfig>) =>
    getPublicClient({ ...defaultPublicClientConfig, ...publicClientConfig, ...getChainConfig(hre) });

  hre.viem.getWalletClients = (walletClientConfig?: Partial<WalletClientConfig>) =>
    getWalletClients({ ...walletClientConfig, ...getChainConfig(hre) });

  hre.viem.getWalletClient = (address: Address, walletClientConfig?: Partial<WalletClientConfig>) =>
    getWalletClient(address, { ...walletClientConfig, ...getChainConfig(hre) });

  hre.viem.deployContract = (async <CN extends string>(
    contractName: ContractName<CN>,
    // eslint-disable-next-line  @typescript-eslint/no-explicit-any
    constructorArgs: any[] = [],
    config: DeployContractConfig = {},
  ) => {
    const publicClient = config.client?.public ?? (await hre.viem.getPublicClient());
    const walletClient = config.client?.wallet ?? (await hre.viem.getWalletClients())[0];

    config.client = { public: publicClient, wallet: walletClient };

    return deployContract(contractName, constructorArgs, config);
  }) as HardhatRuntimeEnvironment["viem"]["deployContract"];
}
