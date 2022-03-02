#! /bin/bash
set -e

validatorMnemonic="equip town gesture square tomorrow volume nephew minute witness beef rich gadget actress egg sing secret pole winter alarm law today check violin uncover"
# kava1ak4pa9z2aty94ze2cs06wsdnkg9hsvfkp40r02

faucetMnemonic="crash sort dwarf disease change advice attract clump avoid mobile clump right junior axis book fresh mask tube front require until face effort vault"
# kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz

evmFaucetMnemonic="hundred flash cattle inquiry gorilla quick enact lazy galaxy apple bitter liberty print sun hurdle oak town cash because round chalk marriage response success"
# 0x3C854F92F726A7897C8B23F55B2D6E2C482EF3E0
# kava18jz5lyhhy6ncjlyty064kttw93yzaulq7rlptu

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

# Create a delegation tx for the validator and add to genesis
$BINARY gentx $validatorKeyName 1000000000ukava --keyring-backend test --chain-id $chainID
$BINARY collect-gentxs

# Replace stake with ukava
sed -in-place='' 's/stake/ukava/g' $DATA/config/genesis.json

# Replace the default evm denom of aphoton with ukava
sed -in-place='' 's/aphoton/ukava/g' $DATA/config/genesis.json

# Zero out the total supply so it gets recalculated during InitGenesis
jq '.app_state.bank.supply = []' $DATA/config/genesis.json|sponge $DATA/config/genesis.json
