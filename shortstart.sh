###########################################################################
######## CODE REVIEW NOTE: THIS FILE IS JUST FOR TESTING AND MAYBE WE DO NOT NEED 
######## TO INCLUDE IT IN THE PULL REQUEST
###########################################################################

#! /bin/bash

kvdHome=/tmp/kvdHome
kvcliHome=/tmp/kvcliHome
swaggerFile=swagger-ui/testnet-4000/swagger-testnet-4000.yaml


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

rest_test/./stopchain.sh
kvd unsafe-reset-all --home $kvdHome

# start the blockchain in the background, wait until it starts making blocks
{
kvd start --home $kvdHome & kvdPid="$!"
} > /dev/null 2>&1

printf "\n"
sleep 3 & showLoading "Starting rest server, please wait"
# start the rest server. Use ./stopchain.sh to stop both rest server and the blockchain
{
kvcli rest-server --laddr tcp://127.0.0.1:1317 --chain-id=testing --home $kvcliHome & kvcliPid="$!"
} > /dev/null 2>&1
printf "\n"
sleep 3 & showLoading "Preparing blockchain setup transactions, please wait"
printf "\n"

# build the go setup test file
rm -f rest_test/setuptest
go build rest_test/setup/setuptest.go & showLoading "Building go test file, please wait"

# run the go code to send transactions to the chain and set it up correctly
./setuptest $kvcliHome & showLoading "Sending messages to blockchain"
printf "\n"
printf "Blockchain setup completed"
printf "\n\n"

############################
# Now run the dredd tests
############################

# dredd $swaggerFile localhost:1317 2>&1 | tee output & showLoading "Running dredd tests"

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
      # pgrep kvd | xargs kill
      # pgrep kvcli | xargs kill & showLoading "Stopping blockchain"
      rm -f output
      exit 0
    fi
  fi
fi

# otherwise return an error code and redirect stderr to stdout so user sees the error output
echo "Failure" >&2
# rm setuptest & showLoading "Cleaning up go binary"
# kill the kvd and kvcli processes (blockchain and rest api)
# pgrep kvd | xargs kill
# pgrep kvcli | xargs kill & showLoading "Stopping blockchain"
rm -f output
exit 1
