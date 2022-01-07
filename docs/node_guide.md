# Node Guide

This is an updated guide on setting up a mainnet full or light node. If you have questions, please join the active conversation in the #validators thread of our [__Discord Channel__](https://discord.com/invite/kQzh3Uv).

## Installing Kava

### Prerequisites
You should select an all-purpose server with at least 8GB of RAM, good connectivity, and a solid state drive with sufficient disk space. Storage requirements are discussed further in the section below. In addition, you’ll need to open **port 26656** to connect to the Kava peer-to-peer network. As the usage of the blockchain grows, the server requirements may increase as well, so you should have a plan for updating your server as well.

### Storage
The monthly storage requirements for a node are as follows. These are estimated values based on experience, but should serve as a good guide.

- A light node (`pruning = "everything"`) grows at a rate of ~5 GB per month.
- A full node (`pruning = "nothing"`) grows at a rate of ~100 GB per month

## Install Go
Kava is built using Go and requires Go version 1.13+. In this example, you will be installing Go on a fresh install of ubuntu 18.04.

```bash
# Update ubuntu
sudo apt update
sudo apt upgrade -y

# Install packages necessary to run go and jq for pretty formatting command line outputs
sudo apt install build-essential jq -y

# Install go
wget https://dl.google.com/go/go1.17.1.linux-amd64.tar.gz (or latest version at https://golang.org/dl/)
sudo tar -xvf go1.17.1.linux-amd64.tar.gz
sudo mv go /usr/local

# Updates environmental variables to include go
cat <<EOF>> ~/.profile
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
EOF
source ~/.profile
```
To verify that Go is installed:
```bash
go version
# Should return go version go1.17.1 linux/amd64
```

## Install Kava
Install Kava using `git clone`. Note that version 0.15.1 is the correct version for mainnet.

```bash
git clone https://github.com/kava-labs/kava
cd kava
git checkout v0.15.1
make install
```
To verify that kava is installed:
```bash
kvd version --long
# name: kava
# server_name: kvd
# client_name: kvcli
# version: 0.15.1
# commit: f0c90f0cbf96d230a83cd2309b8fd032e52d7fb933881541472df1bf2703a939

# build_tags: netgo,ledger
# go: go version go1.15.14 linux/amd64
```

## Configuring Your Node
Next, download the correct genesis file and sync your node with the Kava mainnet. To download the genesis file:
```bash
# First, initialize kvd. Replace <name> with the public name of your node
kvd init --chain-id kava-8 <name>
# Download the genesis file
wget https://kava-genesis-files.s3.amazonaws.com/kava-8-genesis-migrated-from-block-1878508.json -O ~/.kvd/config/genesis.json
# Verify genesis hash
jq -S -c -M '' $HOME/.kvd/config/genesis.json | shasum -a 256
# f0c90f0cbf96d230a83cd2309b8fd032e52d7fb933881541472df1bf2703a939
```
Next,  adjust some configurations. To open the config file:
```bash
vim $HOME/.kvd/config/config.toml
```
At line 160, add [seeds](https://docs.google.com/spreadsheets/d/1TWsD2lMi1idkPI6W9xFCn5W64x75yn9PDjvQaJVkIRk/edit?usp=sharing). These are used to connect to the peer-to-peer network:

At line 163, add some [persistent peers](https://docs.google.com/spreadsheets/d/1TWsD2lMi1idkPI6W9xFCn5W64x75yn9PDjvQaJVkIRk/edit?usp=sharing), which help maintain a connection to the peer-to-peer network


Next, chose how much historical state you want to store. To open the application config file:
```bash
vim $HOME/.kvd/config/app.toml
```
In this file, choose between `default`, `nothing`, and `everything`. To reduce hard drive storage, choose `everything` or `default`. To run an archival node, chose `nothing`.
```bash
pruning = "default"
```
In the same file, you will want to set minimum gas prices — setting a minimum prevents spam transactions:
```bash
minimum-gas-prices = "0.001ukava"
```
### Syncing Your Node
To sync your node, you will use systemd, which manages the Kava daemon and automatically restarts it in case of failure. To use systemd, you will create a service file. Be sure to replace `<your_user>` with the user on your server:
```bash
sudo tee /etc/systemd/system/kvd.service > /dev/null <<'EOF'
[Unit]
Description=Kava daemon
After=network-online.target

[Service]
User=<your_user>
ExecStart=/home/<your_user>/go/bin/kvd start
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF
```
To start syncing:
```bash
# Start the node
sudo systemctl enable kvd
sudo systemctl start kvd
```
To check on the status of syncing:
```bash
kvcli status --output json | jq '.sync_info'
```
This will give output like:
```bash
{
"latest_block_hash": "21D7E37A0A5992E1992DD43E42C05E4475A6E212694F746ABEE132267067847D",
"latest_app_hash": "FBE0E799BCCA57F12F781252787BD6340782E5D45E591294D01269F481B128AC",
"latest_block_height": "183566",
"latest_block_time": "2021-03-22T17:21:41.848445277Z",
"earliest_block_hash": "09E688467E5016159D74CEDE2EE870D671CAA772F76E6697AEEB685A398ACB08",
"earliest_app_hash": "",
"earliest_block_height": "1",
"earliest_block_time": "2021-03-05T06:00:00Z",
"catching_up": false
}
```
The main thing to watch is that the block height is increasing. Once you are caught up with the chain, `catching_up` will become false. At that point, you can start using your node to create a validator. If you need to sync using a snapshot, please use https://kava.quicksync.io/

To check the logs of the node:
```bash
sudo journalctl -u kvd -f
```

If you have questions, please join the active conversation in the #validators thread of the [__Kava Discord Channel__](https://discord.com/invite/kQzh3Uv).
