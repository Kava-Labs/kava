#! /bin/bash
set -e

validatorMnemonic="equip town gesture square tomorrow volume nephew minute witness beef rich gadget actress egg sing secret pole winter alarm law today check violin uncover"
# kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c

faucetMnemonic="crash sort dwarf disease change advice attract clump avoid mobile clump right junior axis book fresh mask tube front require until face effort vault"
# kava1adkm6svtzjsxxvg7g6rshg6kj9qwej8gwqadqd

evmFaucetMnemonic="hundred flash cattle inquiry gorilla quick enact lazy galaxy apple bitter liberty print sun hurdle oak town cash because round chalk marriage response success"
# 0x3C854F92F726A7897C8B23F55B2D6E2C482EF3E0
# kava18jz5lyhhy6ncjlyty064kttw93yzaulq7rlptu

userMnemonic="news tornado sponsor drastic dolphin awful plastic select true lizard width idle ability pigeon runway lift oppose isolate maple aspect safe jungle author hole"
# 0x7Bbf300890857b8c241b219C6a489431669b3aFA
# kava10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t

relayerMnemonic="never reject sniff east arctic funny twin feed upper series stay shoot vivid adapt defense economy pledge fetch invite approve ceiling admit gloom exit"
# 0xa2F728F997f62F47D4262a70947F6c36885dF9fa
# kava15tmj37vh7ch504px9fcfglmvx6y9m70646ev8t

DATA=~/.kava
# remove any old state and config
rm -rf $DATA

BINARY=kava

# Create new data directory, overwriting any that alread existed
chainID="kavalocalnet_8888-1"
$BINARY init validator --chain-id $chainID

# hacky enable of rest api
sed -in-place='' 's/enable = false/enable = true/g' $DATA/config/app.toml

sed -in-place='' 's/enable = false/enable = true/g' $DATA/config/app.toml

sed -in-place='' 's/pruning = "default"/pruning = "nothing"/g' $DATA/config/app.toml

# Set evm tracer to json
sed -in-place='' 's/tracer = ""/tracer = "json"/g' $DATA/config/app.toml

# Set client chain id
sed -in-place='' 's/chain-id = ""/chain-id = "kavalocalnet_8888-1"/g' $DATA/config/client.toml

# avoid having to use password for keys
$BINARY config keyring-backend test

# Create validator keys and add account to genesis
validatorKeyName="validator"
printf "$validatorMnemonic\n" | $BINARY keys add $validatorKeyName --recover
$BINARY add-genesis-account $validatorKeyName 2000000000ukava,100000000000bnb

# Create faucet keys and add account to genesis
faucetKeyName="faucet"
printf "$faucetMnemonic\n" | $BINARY keys add $faucetKeyName --recover
$BINARY add-genesis-account $faucetKeyName 1000000000ukava,100000000000bnb

evmFaucetKeyName="evm-faucet"
printf "$evmFaucetMnemonic\n" | $BINARY keys add $evmFaucetKeyName --eth --recover
$BINARY add-genesis-account $evmFaucetKeyName 1000000000ukava

userKeyName="user"
printf "$userMnemonic\n" | $BINARY keys add $userKeyName --eth --recover
$BINARY add-genesis-account $userKeyName 1000000000ukava

relayerKeyName="relayer"
printf "$relayerMnemonic\n" | $BINARY keys add $relayerKeyName --eth --recover
$BINARY add-genesis-account $relayerKeyName 1000000000ukava

# Create a delegation tx for the validator and add to genesis
$BINARY gentx $validatorKeyName 1000000000ukava --keyring-backend test --chain-id $chainID
$BINARY collect-gentxs

# Replace stake with ukava
sed -in-place='' 's/stake/ukava/g' $DATA/config/genesis.json

# Replace the default evm denom of aphoton with ukava
sed -in-place='' 's/aphoton/akava/g' $DATA/config/genesis.json

# Zero out the total supply so it gets recalculated during InitGenesis
jq '.app_state.bank.supply = []' $DATA/config/genesis.json|sponge $DATA/config/genesis.json

# Disable fee market
jq '.app_state.feemarket.params.no_base_fee = true' $DATA/config/genesis.json|sponge $DATA/config/genesis.json

# Disable london fork
jq '.app_state.evm.params.chain_config.london_block = null' $DATA/config/genesis.json|sponge $DATA/config/genesis.json
jq '.app_state.evm.params.chain_config.arrow_glacier_block = null' $DATA/config/genesis.json|sponge $DATA/config/genesis.json
jq '.app_state.evm.params.chain_config.merge_fork_block = null' $DATA/config/genesis.json|sponge $DATA/config/genesis.json

# Enable bridge
jq '.app_state.bridge.params.bridge_enabled = true' $DATA/config/genesis.json | sponge $DATA/config/genesis.json

# Set relayer to devnet relayer address
jq '.app_state.bridge.params.relayer = "kava15tmj37vh7ch504px9fcfglmvx6y9m70646ev8t"' $DATA/config/genesis.json | sponge $DATA/config/genesis.json

# Set enabled erc20 tokens to match local geth testnet
jq '.app_state.bridge.params.enabled_erc20_tokens = [
    {
        address: "0x6098c27D41ec6dc280c2200A737D443b0AaA2E8F",
        name: "Wrapped ETH",
        symbol: "WETH",
        decimals: 18,
        minimum_withdraw_amount: "10000000000000000"
    },
    {
        address: "0x4Fb48E68842bb59f07569c623ACa5826b600F8F7",
        name: "USDC",
        symbol: "USDC",
        decimals: 6,
        minimum_withdraw_amount: "10000000"
    }]' $DATA/config/genesis.json | sponge $DATA/config/genesis.json

# Set enabled conversion pairs - weth address is the first contract bridge module
# deploys
jq '.app_state.bridge.params.enabled_conversion_pairs = [
    {
        kava_erc20_address: "0x404F9466d758eA33eA84CeBE9E444b06533b369e",
        denom: "erc20/weth",
    }]' $DATA/config/genesis.json | sponge $DATA/config/genesis.json


# create stability committee for migration test
jq '.app_state.committee.committees = [
    {
      "@type": "/kava.committee.v1beta1.MemberCommittee",
      "base_committee": {
        "id": "1",
        "description": "Kava Stability Committee",
        "members": ["kava1n96qpdfcz2m7y364ewk8srv9zuq6ucwduyjaag"],
        "permissions": [
          {
            "@type": "/kava.committee.v1beta1.TextPermission"
          },
          {
            "@type": "/kava.committee.v1beta1.ParamsChangePermission",
            "allowed_params_changes": [
              {
                "subspace": "auction",
                "key": "BidDuration"
              },
              {
                "subspace": "auction",
                "key": "IncrementSurplus"
              },
              {
                "subspace": "auction",
                "key": "IncrementDebt"
              },
              {
                "subspace": "auction",
                "key": "IncrementCollateral"
              },
              {
                "subspace": "bep3",
                "key": "AssetParams"
              },
              {
                "subspace": "cdp",
                "key": "GlobalDebtLimit"
              },
              {
                "subspace": "cdp",
                "key": "SurplusThreshold"
              },
              {
                "subspace": "cdp",
                "key": "SurplusLot"
              },
              {
                "subspace": "cdp",
                "key": "DebtThreshold"
              },
              {
                "subspace": "cdp",
                "key": "DebtLot"
              },
              {
                "subspace": "cdp",
                "key": "DistributionFrequency"
              },
              {
                "subspace": "cdp",
                "key": "CollateralParams"
              },
              {
                "subspace": "cdp",
                "key": "DebtParam"
              },
              {
                "subspace": "incentive",
                "key": "Active"
              },
              {
                "subspace": "kavadist",
                "key": "Active"
              },
              {
                "subspace": "pricefeed",
                "key": "Markets"
              },
              {
                "subspace": "hard",
                "key": "MoneyMarkets"
              },
              {
                "subspace": "hard",
                "key": "MinimumBorrowUSDValue"
              }
            ]
          }
        ],
        "vote_threshold": "0.667000000000000000",
        "proposal_duration": "604800s",
        "tally_option": "TALLY_OPTION_FIRST_PAST_THE_POST"
      }
    }
]' $DATA/config/genesis.json | sponge $DATA/config/genesis.json
