#! /bin/bash
set -e




validatorMnemonic="equip town gesture square tomorrow volume nephew minute witness beef rich gadget actress egg sing secret pole winter alarm law today check violin uncover"
# kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c
faucetMnemonic="crash sort dwarf disease change advice attract clump avoid mobile clump right junior axis book fresh mask tube front require until face effort vault"
# kava1adkm6svtzjsxxvg7g6rshg6kj9qwej8gwqadqd

DATA=~/.kava
# remove any old state and config
rm -rf $DATA

BINARY=kava

# Create new data directory, overwriting any that alread existed
chainID="kava-localnet"
$BINARY init validator --chain-id $chainID

# hacky enable of rest api
sed -in-place='' 's/enable = false/enable = true/g' $DATA/config/app.toml

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

# Create a delegation tx for the validator and add to genesis
$BINARY gentx $validatorKeyName 1000000000ukava --keyring-backend test --chain-id $chainID
$BINARY collect-gentxs

# Replace stake with ukava
sed -in-place='' 's/stake/ukava/g' $DATA/config/genesis.json

# Zero out the total supply so it gets recalculated during InitGenesis
jq '.app_state.bank.supply = []' $DATA/config/genesis.json|sponge $DATA/config/genesis.json