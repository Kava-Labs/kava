#! /bin/bash
# TODO import from development environment in envkey
password="password"
validatorMnemonic="equip town gesture square tomorrow volume nephew minute witness beef rich gadget actress egg sing secret pole winter alarm law today check violin uncover"

#   address: kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c
#   address: kavavaloper1ffv7nhd3z6sych2qpqkk03ec6hzkmufyz4scd0

faucet="chief access utility giant burger winner jar false naive mobile often perfect advice village love enroll embark bacon under flock harbor render father since"

# address: kava1ls82zzghsx0exkpr52m8vht5jqs3un0ceysshz
# address: kavavaloper1ls82zzghsx0exkpr52m8vht5jqs3un0c5j2c04

# Remove any existing data directory
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

  echo "$loadingText...finished"
}

# build the go setup test file
rm -f test
go build debugging_tools/test.go & showLoading "Building go test file, please wait"

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
sleep 10 & showLoading "Starting rest server, please wait"
# start the rest server. Use ./stopchain.sh to stop both rest server and the blockchain
{
kvcli rest-server --laddr tcp://127.0.0.1:1317 --chain-id=testing --home ~/kavatmp & kvcliPid="$!"
} > /dev/null 2>&1
printf "\n"
sleep 10 & showLoading "Preparing blockchain setup transactions, please wait"
printf "\n"
# run the go code to send transactions to the chain and set it up correctly
./test & showLoading "Sending messages to blockchain"
printf "\n"
printf "Blockchain setup completed"
printf "\n\n"

############################
# Now run the dredd tests
############################

dredd swagger.yaml localhost:1317 & showLoading "Running dredd tests"

########################################################
# Now run the check the return code from the dredd command. 
# If 0 then all test passed OK, otherwise some failed and propagate the error
########################################################

if [ $? -eq 0 ]
then
  echo "Success"
  rm test & showLoading "Cleaning up go binary"
  pgrep kvd | xargs kill
  pgrep kvcli | xargs kill & showLoading "Stopping blockchain"
  exit 0
else
  echo "Failure" >&2
  rm test & showLoading "Cleaning up go binary"
  pgrep kvd | xargs kill
  pgrep kvcli | xargs kill & showLoading "Stopping blockchain"
  exit 1
fi
