<!--![Kava](./kava-logo.svg)-->
# Kava Blockchain

A decentralized fast-finality blockchain for interoperable payment channel networks.
Building on the work of Tendermint and Interledger.

Project status: We're currently in a very early public testnet. With future features being implemented.

Try it out - send txs using our public node, or run a full node to sync to the testnet, or even run a validator.



## Install

### Source

Requirements: go installed and set up.

	mkdir -p $GOPATH/src/github.com/kava-labs
	cd $GOPATH/src/github.com/kava-labs
	git clone https://github.com/kava-labs/kava
	cd kava
	dep ensure
	go install ./cmd/kvd
	go install ./cmd/kvcli

<!-- Make sure GOBIN environment variable is set if you want to access programs anywhere -->

### Docker

TODO
<!-- Requirements: docker installed.

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

	kvd init --name <your-name> --chain-id kava-test-0

This will generate config and keys in `$HOME/.kvd` and `$HOME/.kvcli`.
The default password is 'password'.

Copy the testnet genesis file (from https://raw.githubusercontent.com/Kava-Labs/kava/master/testnets/kava-test-1/genesis.json) into `$HOME/.kvd/config/`, replacing the existing one.

Add the kava node address (`0dfd43e440e34fc193ddee4ae99547184f3cb5d1@validator.connector.kava.io:26656`) to `seeds` in `$HOME/.kvd/config/config.toml`

Start your full node

	kvd start


## Run a Validator
Join the [validator chat](https://riot.im/app/#/room/#kava-validators:matrix.org). Follow setup for a full node above.
Get you address with `kvcli keys list`. Should look like `cosmosaccaddr10jpp289accvkhsvrpz4tlj9zhqdaey2tl9m4rg`.
Ask @rhuairahrighairidh in the chat to give you some coins.

Get your validator pubkey with `kvd tendermint show_validator`

	kvcli stake create-validator \
            --amount 1000KVA \
            --pubkey <you validator pubkey from above> \
            --address-validator <your address from above> \
            --moniker "<your name>" \
            --chain-id kava-test-0 \
            --from <your name>

Now you should be participating in consensus and validating blocks!


Running a validator requires that you keep validating blocks. If you stop then your stake will be slashed.
In order to stop validating, first remove yourself as validator, then you can stop your node.

	kvcli stake unbond begin \
		--address-delegator <your address> \
		--address-validator <your address> \
		--chain-id kava-test-0 \
		--shares-percent 1 \
		--from <your name>
