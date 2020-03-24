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
# rm -rf ~/.kvd
# rm -rf ~/.kvcli

rm -rf ~/kavatmp

# create validator key
printf "$password\n$validatorMnemonic\n" | kvcli keys add vlad --recover --home ~/kavatmp
# create faucet key
printf "$password\n$faucet\n" | kvcli --home ~/kavatmp keys add faucet --recover --home ~/kavatmp

# function used to show that it is still loading
showLoading() {
  mypid=$!
  loadingText=$1

  echo -ne "$loadingText\r"

  while kill -0 $mypid 2>/dev/null; do
    echo -ne "$loadingText.\r"
    sleep 0.5
    echo -ne "$loadingText..\r"
    sleep 0.5
    echo -ne "$loadingText...\r"
    sleep 0.5
    echo -ne "\r\033[K"
    echo -ne "$loadingText\r"
    sleep 0.5
  done

  echo "$loadingText...FINISHED"
}



# Create new data directory
{
kvd --home ~/kavatmp init --chain-id=testing vlad # doesn't need to be the same as the validator
} > /dev/null 2>&1
kvcli --home ~/kavatmp config chain-id testing # or set trust-node true

# add validator account to genesis
kvd --home ~/kavatmp add-genesis-account $(kvcli --home ~/kavatmp keys show vlad -a) 10000000000000stake
# add faucet account to genesis
kvd --home ~/kavatmp add-genesis-account $(kvcli --home ~/kavatmp keys show faucet -a) 10000000000000stake,1000000000000xrp,100000000000btc

# Create a delegation tx for the validator and add to genesis
printf "$password\n" | kvd --home ~/kavatmp gentx --name vlad --home-client ~/kavatmp
{
kvd --home ~/kavatmp collect-gentxs
} > /dev/null 2>&1

# start the blockchain in the background, wait until it starts making blocks
{
kvd start --home ~/kavatmp & kvdPid="$!"
} > /dev/null 2>&1

printf "\n"
sleep 10 & showLoading "STARTING REST SERVER, PLEASE WAIT"
# start the rest server. Use ./stopchain.sh to stop both rest server and the blockchain
{
kvcli rest-server --laddr tcp://127.0.0.1:1317 --chain-id=testing --home ~/kavatmp & kvcliPid="$!"
} > /dev/null 2>&1
printf "\n"
sleep 10 & showLoading "PREPARING BLOCKCHAIN SETUP TRANSACTIONS, PLEASE WAIT"
# run the go code to send transactions to the chain and set it up correctly
../debugging_tools/./test
printf "\n"
printf "Blockchain setup completed"
printf "\n"
