#! /bin/bash
set -e

validatorMnemonic="equip town gesture square tomorrow volume nephew minute witness beef rich gadget actress egg sing secret pole winter alarm law today check violin uncover"
# kava1ak4pa9z2aty94ze2cs06wsdnkg9hsvfkp40r02
# 0xedaa1e944aeac85a8b2ac41fa741b3b20b783136
faucetMnemonic="crash sort dwarf disease change advice attract clump avoid mobile clump right junior axis book fresh mask tube front require until face effort vault"
# kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz
# 0x5452643ffa7ab2be7ef7c376f62fdfa7804ed1d2


DATA=~/.kava
# remove any old state and config
rm -rf $DATA

BINARY=kava

# Create new data directory, overwriting any that alread existed
chainID="kavalocalnet_8888-1"
$BINARY init validator --chain-id $chainID

# hacky enable of rest api
sed -in-place='' 's/enable = false/enable = true/g' $DATA/config/app.toml

# Set evm tracer to json
sed -in-place='' 's/tracer = ""/tracer = "json"/g' $DATA/config/app.toml

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

# Replace the default evm denom of aphoton with ukava
sed -in-place='' 's/aphoton/akava/g' $DATA/config/genesis.json

# Zero out the total supply so it gets recalculated during InitGenesis
jq '.app_state.bank.supply = []' $DATA/config/genesis.json|sponge $DATA/config/genesis.json
