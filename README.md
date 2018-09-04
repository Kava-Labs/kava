<h1>
  <img alt="Kava Blockchain" src="./kava-logo.svg" width="200">
</h1>

A decentralized fast-finality blockchain for interoperable payment channel networks.
Proudly building on the work of [Cosmos](https://github.com/cosmos/cosmos-sdk) and [Interledger](https://github.com/interledger).

Project status: We're currently in a very early public testnet. With future features being implemented.

Try it out - run a full node to sync to the testnet, [send some off chain payments](internal/x/paychan/README.md), or set up as a validator.


## Install

<!--### Source-->

Requirements: go installed and set up (version 1.10+).

 0. If installing from a new Ubuntu server (16.04 or 18.04), here's how to setup go:
		
		sudo apt update
		sudo apt upgrade -y
		sudo apt install git gcc make wget -y
		wget https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz
		sudo tar -xvf go1.10.3.linux-amd64.tar.gz
		sudo mv go /usr/local

		cat >> ~/.profile <<EOF
		export GOROOT=/usr/local/go
		export GOPATH=\$HOME/go
		export PATH=\$GOPATH/bin:\$GOROOT/bin:\$PATH
		EOF

		source ~/.profile

 1. Get the code.
 
		mkdir -p $GOPATH/src/github.com/kava-labs
		cd $GOPATH/src/github.com/kava-labs
		git clone https://github.com/kava-labs/kava
		cd kava
		git checkout 8c9406c
	
 2. Install the dependencies.
 
		mkdir $GOPATH/bin
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
		dep ensure

3. Install the code

		go install ./cmd/kvd
		go install ./cmd/kvcli

<!--### Docker

Requirements: docker installed.

No installation necessary, just prepend commands with `docker run kava/kava`.  TODO name necessary to avoid new contianer being created each time?

This will use our docker container `kava/kava` and store all blockchain data and keys within the container. -->

<!-- To store this data outisde the conatiner, attach volumes to the container:

	docker run --rm -v $HOME/.kvd:/root/.kvd -v $HOME/.kvcli:/root/.kvcli kava/kava <further commands>

Now blockchain data will be stored in `$HOME/.kvd` and keys in `$HOME/.kvcli`. Also the `--rm` flag removes the contianer after each run.

 -->
<!-- ## Send Transactions

You can send transactions on the testnet using our node without yncing a local node.
Requirements

TODO users need to set up keys first?

	kvcli <args> --node validator.connector.kava.io:26657 --chain-id kava-test-<current version>
 -->

## Run a Full Node

	kvd init --name <your-name> --chain-id kava-test-2

This will generate config and keys in `$HOME/.kvd` and `$HOME/.kvcli`. The default password is 'password'.

> Note: Make sure `GOBIN` is set and added to your path if you want to be able to run installed go programs from any folder.

Copy the testnet genesis file (from https://raw.githubusercontent.com/Kava-Labs/kava/master/testnets/kava-test-2/genesis.json) into `$HOME/.kvd/config/`, replacing the existing one.

Add the kava node address, `5c2bc5a95b014e4b2897791565398ee6bfd0a04a@validator.connector.kava.io:26656`, to `seeds` in `$HOME/.kvd/config/config.toml`

Start your full node

	kvd start
	
> Note: It might take a while to fully sync. Check the latest block height [here](http://validator.connector.kava.io:26657/abci_info).


## Run a Validator
Join the [validator chat](https://riot.im/app/#/room/#kava-validators:matrix.org). Follow setup for a full node above.

Get you address with `kvcli keys list`. Should look something like `cosmosaccaddr10jpp289accvkhsvrpz4tlj9zhqdaey2tl9m4rg`.

Ask @rhuairahrighairidh in the chat to give you some coins.

Get your validator pubkey with `kvd tendermint show_validator`

Then, your full running in the background or separate window, run:

	kvcli stake create-validator \
            --amount 900KVA \
            --pubkey <you validator pubkey from above> \
            --address-validator <your address from above> \
            --moniker "<your name>" \
            --chain-id kava-test-2 \
            --from <your name>
> Note You'll need to type in the default password "password"

Now your full node should be participating in consensus and validating blocks!

Running a validator requires that you keep validating blocks. If you stop, your stake will be slashed.  
In order to stop validating, first remove yourself as validator, then you can stop your node.

	kvcli stake unbond begin \
		--address-delegator <your address> \
		--address-validator <your address> \
		--chain-id kava-test-2 \
		--shares-percent 1 \
		--from <your name>
