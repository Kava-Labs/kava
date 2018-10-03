
# Installation and Setup

#### Who this guide is for
The code currently consists of a full node daemon (`kvd`) and it's command line interface (`kvcli`).

Full nodes are fairly resource intensive and are designed to be run continuously on servers, primarily by people validating the network. While it is possible to run locally it is not recommended except for development purposes. In the future light clients will enable secure transactions for clients.

A **full node** syncs with the blockchain and processes transactions. A **validator** is a full node that has declared itself to be a "validator" on chain. This obligates it to participate in consensus by proposing and signing blocks and maintaining uptime. By not following the protocol, the validator's stake will be slashed.

Use these instructions to set up a full node, and to optionally declare yourself as a validator.

## Install

Requirements: go installed and set up (version 1.10+).

>If installing from a new Ubuntu server (16.04 or 18.04), here's how to setup go.
>
>       sudo apt update
>       sudo apt upgrade -y
>       sudo apt install git gcc make wget -y
>       wget https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz
>       sudo tar -xvf go1.10.3.linux-amd64.tar.gz
>       sudo mv go /usr/local
>
>       cat >> ~/.profile <<EOF
>       export GOROOT=/usr/local/go
>       export GOPATH=\$HOME/go
>       export PATH=\$GOPATH/bin:\$GOROOT/bin:\$PATH
>       EOF
>
>       source ~/.profile

 1. Get the code.
 
        mkdir -p $GOPATH/src/github.com/kava-labs
        cd $GOPATH/src/github.com/kava-labs
        git clone https://github.com/kava-labs/kava
        cd kava
    
 2. Install the dependencies.
 
        mkdir $GOPATH/bin
        curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
        dep ensure -vendor-only

 3. Install the code.

        go install ./cmd/kvd
        go install ./cmd/kvcli

## Run a Full Node

    kvd init --name <your-name>

Enter a new password for your validator key. This will generate generic config and private keys in `$HOME/.kvd` and `$HOME/.kvcli`.

> Note: Make sure `GOBIN` is set and added to your path if you want to be able to run installed go programs from any folder.

Setup the correct config for the current testnet

    kvd init-testnet --name <your-name> --chain-id kava-test-3

Start your full node

    kvd start

> Note: It might take a while to fully sync. Check the latest block height [here](http://validator.connector.kava.io:17127/abci_info).

### Running in the Background
It's often more convenient to start `kvd` in the background and capture output to a log file.

    kvd start &> kvd.log &

To see the output of the log.

    tail -f kvd.log



## Become a Validator
Join the [validator chat](https://riot.im/app/#/room/#kava-validators:matrix.org).

Follow the setup for a full node above.

Get your address with `kvcli keys list` and your validator pubkey with `kvd tendermint show_validator`.

Get some testnet coins from [the faucet](http://kava-test-3.faucet.connector.kava.io).

Then, with your full node running in the background or separate window, run:

    kvcli stake create-validator \
            --amount 900KVA \
            --pubkey <you validator pubkey from above> \
            --address-validator <your address from above> \
            --moniker "<your name>" \
            --from <your name> \
            --gas 1000000

> Note You'll need to type in your password you set earlier.

Now your full node should be participating in consensus and validating blocks!

Running a validator requires that you keep validating blocks. If you stop, your stake will be slashed.  
In order to stop validating, first remove yourself as validator, then you can stop your node.

    kvcli stake unbond begin \
        --address-delegator <your address> \
        --address-validator <your address> \
        --shares-percent 1 \
        --from <your name> \
        --gas 1000000
