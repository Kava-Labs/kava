<!--
order: 5
-->

# Parameters

The evmutil module contains the following parameters:

| Key                    | Type                                 | Example       |
| ---------------------- | ------------------------------------ | ------------- |
| EnabledConversionPairs | array (ConversionPair)               | [{see below}] |
| AllowedCosmosDenoms    | array (AllowedCosmosCoinERC20Tokens) | [{see below}] |

Example parameters for `ConversionPair`:

| Key                | Type   | Example                                      | Description                        |
| ------------------ | ------ | -------------------------------------------- | ---------------------------------- |
| kava_erc20_Address | string | "0x43d8814fdfb9b8854422df13f1c66e34e4fa91fd" | ERC20 contract address             |
| denom              | string | "erc20/chain/usdc"                           | sdk.Coin denom for the ERC20 token |

Example parameters for `AllowedCosmosCoinERC20Token`:

| Key          | Type   | Example                                                                | Description                                        |
| ------------ | ------ | ---------------------------------------------------------------------- | -------------------------------------------------- |
| cosmos_denom | string | "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2" | denom of the sdk.Coin                              |
| name         | string | "Kava-wrapped Atom"                                                    | name field of the erc20 token                      |
| symbol       | string | "kATOM"                                                                | symbol field of the erc20 token                    |
| decimal      | uint32 | 6                                                                      | decimal field of the erc20 token, for display only |

## EnabledConversionPairs

The enabled conversion pairs parameter is an array of ConversionPair entries mapping an erc20 address to a sdk.Coin denom. Only erc20 contract addresses that are in this list can be converted to sdk.Coin and vice versa.

## AllowedCosmosDenoms

The allowed cosmos denoms parameter is an array of AllowedCosmosCoinERC20Token entries. They include the cosmos-sdk.Coin denom and metadata for the ERC20 representation of the asset in Kava's EVM. Coins may only be transferred to the EVM if they are included in this list. A token in this list will have an ERC20 token contract deployed on first conversion. The token will be deployed with the metadata included in the AllowedCosmosCoinERC20Token. Once deployed, changes to the metadata will not affect or change the deployed contract.
