#! /bin/bash
# TODO import from development environment in envkey
password="password"
validatorMnemonic="equip town gesture square tomorrow volume nephew minute witness beef rich gadget actress egg sing secret pole winter alarm law today check violin uncover"
# kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw

#   address: kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c
#   address: kavavaloper1ffv7nhd3z6sych2qpqkk03ec6hzkmufyz4scd0

faucet="chief access utility giant burger winner jar false naive mobile often perfect advice village love enroll embark bacon under flock harbor render father since"
# kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj

# address: kava1ls82zzghsx0exkpr52m8vht5jqs3un0ceysshz
# address: kavavaloper1ls82zzghsx0exkpr52m8vht5jqs3un0c5j2c04

# Remove any existing data directory
rm -rf ~/.kvd
rm -rf ~/.kvcli

# create validator key
printf "$password\n$validatorMnemonic\n" | kvcli keys add vlad --recover
# create faucet key
printf "$password\n$faucet\n" | kvcli keys add faucet --recover


# Create new data directory
kvd init --chain-id=testing vlad # doesn't need to be the same as the validator
kvcli config chain-id testing # or set trust-node true

# add validator account to genesis
kvd add-genesis-account $(kvcli keys show vlad -a) 10000000000000stake
# add faucet account to genesis
kvd add-genesis-account $(kvcli keys show faucet -a) 10000000000000stake,1000000000000xrp,100000000000btc

# Create a delegation tx for the validator and add to genesis
printf "$password\n" | kvd gentx --name vlad
kvd collect-gentxs

# start the blockchain in the background, record the process id so that it can be stopped, wait until it starts making blocks
kvd start & kvdPid="$!"
sleep 10
# start the rest server. Ctrl-C  will stop both rest server and the blockchain (on crtl-c the kill thing runs and stops the blockchain process)
kvcli rest-server ; kill $kvdPid