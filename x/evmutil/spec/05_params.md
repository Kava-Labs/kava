<!--
order: 5
-->

# Parameters

The evmutil module contains the following parameters:

| Key                    | Type                   | Example       |
| ---------------------- | ---------------------- | ------------- |
| EnabledConversionPairs | array (ConversionPair) | [{see below}] |

Example parameters for `ConversionPair`:

| Key                | Type   | Example                                      | Description                        |
| ------------------ | ------ | -------------------------------------------- | ---------------------------------- |
| kava_erc20_Address | string | "0x43d8814fdfb9b8854422df13f1c66e34e4fa91fd" | ERC20 contract address             |
| denom              | string | "erc20/chain/usdc"                           | sdk.Coin denom for the ERC20 token |

## EnabledConversionPairs

The enabled conversion pairs parameter is an array of ConversionPair entries mapping an erc20 address to a sdk.Coin denom. Only erc20 contract addresses that are in this list can be converted to sdk.Coin and vice versa.
