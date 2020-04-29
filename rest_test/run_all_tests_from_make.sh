#! /bin/bash

# kill kava if it is already running
pgrep kvd | xargs kill
pgrep kvcli | xargs kill

# TODO import from development environment in envkey
password=""
validatorMnemonic="equip town gesture square tomorrow volume nephew minute witness beef rich gadget actress egg sing secret pole winter alarm law today check violin uncover"
#   address: kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c
#   address: kavavaloper1ffv7nhd3z6sych2qpqkk03ec6hzkmufyz4scd0
faucet="chief access utility giant burger winner jar false naive mobile often perfect advice village love enroll embark bacon under flock harbor render father since"
# address: kava1ls82zzghsx0exkpr52m8vht5jqs3un0ceysshz
# address: kavavaloper1ls82zzghsx0exkpr52m8vht5jqs3un0c5j2c04
# variables for home directories for kvd and kvcli
kvdHome=/tmp/kvdHome
kvcliHome=/tmp/kvcliHome
genesis=$kvdHome/config/genesis.json
swaggerFile=swagger-ui/testnet-4000/swagger-testnet-4000.yaml
# Remove any existing data directory
rm -rf $kvdHome
rm -rf $kvcliHome
# make the directories
mkdir /tmp/kvdHome
mkdir /tmp/kvcliHome
# create validator key
printf "$validatorMnemonic\n" | kvcli keys add vlad --recover --home $kvcliHome --keyring-backend test
# create faucet key
printf "$faucet\n" | kvcli --home $kvcliHome keys add faucet --recover --home $kvcliHome --keyring-backend test

kvcli keys migrate --home "/tmp/kvcliHome"

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
# Create new data directory
{
kvd --home $kvdHome init --chain-id=testing vlad # doesn't need to be the same as the validator
} > /dev/null 2>&1
kvcli --home $kvcliHome config chain-id testing # or set trust-node true
# add validator account to genesis
kvd --home $kvdHome add-genesis-account $(kvcli --home $kvcliHome keys show vlad -a) 10000000000000stake,1000000000000xrp,100000000000btc
# add faucet account to genesis
kvd --home $kvdHome add-genesis-account $(kvcli --home $kvcliHome keys show faucet -a) 10000000000000stake,1000000000000xrp,100000000000btc
# Create a delegation tx for the validator and add to genesis
printf "$password\n" | kvd --home $kvdHome gentx --name vlad --home-client $kvcliHome
{
kvd --home $kvdHome collect-gentxs
} > /dev/null 2>&1
# set blocktime to ~1s
jq '.app_state.mint.params.blocks_per_year = "31540000"' $genesis > $genesis.tmp && mv $genesis.tmp $genesis
# update pricefeed information
jq '.app_state.pricefeed.params.markets += [{"active": true, "base_asset": "xrp", "market_id": "xrp:usd", "oracles": ["kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"], "quote_asset": "usd"}, {"active": true, "base_asset": "btc", "market_id": "btc:usd", "oracles": ["kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c"], "quote_asset": "usd"}]' $genesis > $genesis.tmp && mv $genesis.tmp $genesis
jq '.app_state.pricefeed.posted_prices += [{"expiry": "2050-01-01T00:00:00Z", "market_id": "btc:usd", "oracle_address": "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c", "price": "8700.0"}, {"expiry": "2050-01-01T00:00:00Z", "market_id": "xrp:usd", "oracle_address": "kava1ffv7nhd3z6sych2qpqkk03ec6hzkmufy0r2s4c", "price": "0.25"}]' $genesis > $genesis.tmp && mv $genesis.tmp $genesis
# now update cdp params
jq '.app_state.cdp.params = { "circuit_breaker": false, "collateral_params": [ { "auction_size": "10000000000", "conversion_factor": "8", "debt_limit": [ { "amount": "1000000000", "denom": "usdx" } ], "denom": "btc", "liquidation_penalty": "0.05", "liquidation_ratio": "1.5", "market_id": "btc:usd", "prefix": 0, "stability_fee": "1.0000000007829977" }, { "auction_size": "100000000", "conversion_factor": "6", "debt_limit": [ { "amount": "10000000", "denom": "usdx" } ], "denom": "xrp", "liquidation_penalty": "0.1", "liquidation_ratio": "2.0", "market_id": "xrp:usd", "prefix": 1, "stability_fee": "1.0000000007829977"} ], "debt_auction_threshold": "9000000", "debt_params": [ { "conversion_factor": "6", "debt_floor": "10000000", "debt_limit": [ { "amount": "2000000000000", "denom": "usdx" } ], "denom": "usdx", "reference_asset": "usd",  "savings_rate": "0.95" } ], "global_debt_limit": [ { "amount": "2000000000000", "denom": "usdx" } ], "surplus_auction_threshold": "9000000", "savings_distribution_frequency": "120000000000" }' $genesis > $genesis.tmp && mv $genesis.tmp $genesis
# start the blockchain in the background, wait until it starts making blocks
{
kvd start --home $kvdHome & kvdPid="$!"
} > /dev/null 2>&1
printf "\n"
sleep 10 & showLoading "Starting rest server, please wait"
# start the rest server. Use ./stopchain.sh to stop both rest server and the blockchain
{
kvcli rest-server --laddr tcp://127.0.0.1:1317 --chain-id=testing --home $kvcliHome & kvcliPid="$!"
} > /dev/null 2>&1
printf "\n"
sleep 10 & showLoading "Preparing blockchain setup transactions, please wait"
printf "\n"
# build the go setup test file
rm -f rest_test/setuptest
go build rest_test/setup/setuptest.go & showLoading "Building go test file, please wait"
# run the go code to send transactions to the chain and set it up correctly
./setuptest $kvcliHome
printf "\n"
printf "Blockchain setup completed"
printf "\n\n"
############################
# Now run the dredd tests
############################
dredd $swaggerFile localhost:1317 2>&1 | tee output & showLoading "Running dredd tests"
########################################################
# Now run the check the return code from the dredd command.
# If 0 then all test passed OK, otherwise some failed and propagate the error
########################################################
# check that the error code was zero
if [ $? -eq 0 ]
then
  # check that all the tests passed (ie zero failing)
  if [[ $(cat output | grep "0 failing") ]]
  then
    # check for no errors
    if [[ $(cat output | grep "0 errors") ]]
    then
      echo "Success"
      rm setuptest & showLoading "Cleaning up go binary"
      # kill the kvd and kvcli processes (blockchain and rest api)
      pgrep kvd | xargs kill
      pgrep kvcli | xargs kill & showLoading "Stopping blockchain"
      rm -f output
      exit 0
    fi
  fi
fi
# otherwise return an error code and redirect stderr to stdout so user sees the error output
echo "Failure" >&2
rm setuptest & showLoading "Cleaning up go binary"
# kill the kvd and kvcli processes (blockchain and rest api)
pgrep kvd | xargs kill
pgrep kvcli | xargs kill & showLoading "Stopping blockchain"
rm -f output
exit 1
