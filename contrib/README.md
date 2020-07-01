# Contrib

Resources and examples for running and interacting with the kava blockchain. `contrib` contains sample genesis files, example governance proposals, and rest-server request examples for different versions of kava.

## Running local testnet

Because kava interacts with other blockchains, it is often insufficient to run a testnet with only kava nodes. This section will describe how to setup a local docker network that consists of

* 4 kava nodes running a local testnet
* 1 kava oracle posting prices
* 1 kava rest server
* 1 binance-chain node running a local testnet
* 1 bep3-deputy that processes swap transactions between kava and binance-chain testnets

### Setting up kava testnet

To create a 4 node kava testnet, you will use the `kvd testnet` command. This command initializes the initial state and binaries needed for docker containers to run a local testnet. In this example, we are running the command from the root of the kava directory. The command looks like

```sh
kvd testnet --v 4 --output-dir ./build --starting-ip-address 192.168.10.2 --genesis-template contrib/kava-3-bep3-asset-supply/genesis-template.json --keyring-backend test --chain-id testing
```

* `--v 4` specifies that 4 node directories will be created
* `--output-dir ./build` specifies that the node directories will be located in a directory called build
* `--starting-ip-address 192.168.10.2` specifies that the nodes will use ip addresses 192.168.10.2 - 192.168.10.5. This is used for peering.
* `--genesis-template contrib/kava-3-bep3-asset-supply/genesis-template.json` specifies the genesis file that will be used as the template for the local testnet. The genesis template can contain pre-populated state, such as CDP parameters, addresses that receive genesis tokens, governance committees, etc. The genesis template should not specify any validators - those will be added by the `testnet` command.
* `--keyring-backend test` specifies that each node will use local, un-encrypted keys with no password. This is useful for testnets where transactions can be sent with minimal authorization.
* `--chain-id testing` sets the chain-id that will be used in the genesis file. Note that `testing` is required for chain-id if using the bep3-deputy.

After the command is run, you should be able to see directories `node0` - `node3` in the `build` directory.

The final step in setting up the network is to create a docker-compose file that specifies which docker container to use for each node and networks the docker containers so they can peer. The portion of the docker compose file for the kava testnet looks like

```yml
services:
  kvdnode0:
    container_name: kvdnode0
    image: "kava/kavanode"
    ports:
      - "26656-26657:26656-26657"
    environment:
      - ID=0
      - LOG=${LOG:-kvd.log}
    volumes:
      - ./build:/kvd:Z
    networks:
      localnet:
        ipv4_address: 192.168.10.2

  kvdnode1:
    container_name: kvdnode1
    image: "kava/kavanode"
    ports:
      - "26659-26660:26656-26657"
    environment:
      - ID=1
      - LOG=${LOG:-kvd.log}
    volumes:
      - ./build:/kvd:Z
    networks:
      localnet:
        ipv4_address: 192.168.10.3

  kvdnode2:
    container_name: kvdnode2
    image: "kava/kavanode"
    environment:
      - ID=2
      - LOG=${LOG:-kvd.log}
    ports:
      - "26661-26662:26656-26657"
    volumes:
      - ./build:/kvd:Z
    networks:
      localnet:
        ipv4_address: 192.168.10.4

  kvdnode3:
    container_name: kvdnode3
    image: "kava/kavanode"
    environment:
      - ID=3
      - LOG=${LOG:-kvd.log}
    ports:
      - "26663-26664:26656-26657"
    volumes:
      - ./build:/kvd:Z
    networks:
      localnet:
        ipv4_address: 192.168.10.5

networks:
  localnet:
    driver: bridge
    ipam:
      driver: default
      config:
      -
        subnet: 192.168.10.0/16
```

Note that the container is named `kava/kavanode`. This docker file is located in `networks/local/kavanode` and can be built by running `make build-docker-local-kava`. Note that the container will use the binary that is built from the current repository. If you want to use a different binary from a different branch or release, simply checkout that branch or release using git.

### Adding Kava Rest Server

To build the rest server container, simply run `make build-docker-local-kava`, which will create a container called `kava/kavarest`. To add it to the docker compose file, add the following under `services`:

```yml
  kvdrest0:
    container_name: kvdrest0
    image: "kava/kavarest"
    environment:
      - ID=0
      - LOG=${LOG:-kvcli.log}
    ports:
      - "26665:1317"
    volumes:
      - ./build:/kvd:Z
    networks:
      localnet:
        ipv4_address: 192.168.10.6
```

### Adding an Oracle Node

If you need a node that will fetch realtime price data and post it to the local kava testnet, you will need to build the docker container and add it to the network. The docker file is located in the root of the `kava-labs/kava-tools` repository. From `kava-tools` run,

```sh
docker build -t kava/oracle .
```

To add it to the docker compose file, add the following under `services`:

```yml
  oracle:
    container_name: oracle
    image: "kava/oracle"
    dns:
      - 8.8.8.8
    networks:
      localnet:
        ipv4_address: 192.168.10.7
```

Note the `dns` argument is required for the node to be able to resolve outside domain names, which it needs to do in order to fetch prices.

### Setting up a local binance-chain

Docker files for creating a local binance-chain testnet can be found in `contrib/bnbchain-testnet`. From that directory, run `make` to create docker files for a binance-chain node and rest server. To add them to the docker compose file, add the following under `services`:

```yml
  bnbnode:
    container_name: bnbnode
    image: "binance/bnbnode"
    ports:
      - "26666:26657"
    networks:
      localnet:
        ipv4_address: 192.168.10.8

  bnbrest:
    container_name: bnbrest
    image: binance/bnbrest
    ports:
      - "26667:8080"
    networks:
      localnet:
        ipv4_address: 192.168.10.9
```

### Setting up BEP3 deputy

A docker file for the bep3 deputy can be found the main [bep3-deputy repo](https://github.com/binance-chain/bep3-deputy). Will have to edit the following ENVs for a local testnet:

```dockerfile
ENV BNB_NETWORK 0
ENV KAVA_NETWORK 0
ENV CONFIG_FILE_PATH config/config.json
```

and delete the line

```dockerfile
USER 1000
```

After that, from the main repository run:

```sh
docker build -t binance/deputy .
```

To add the bep3-deputy to the docker compose file, add the following under `services`:

```yml
  deputy:
    container_name: deputy
    image: binance/deputy
    networks:
      localnet:
        ipv4_address: 192.168.10.10
```

### Starting the local testnet

Save your docker compose file in the root of the kava repository and run:

```sh
make localnet-start
```

When the network is up you will see

```sh
Creating network "kava_localnet" with driver "bridge"
Creating kvdnode1 ... done
Creating kvdrest0 ... done
Creating kvdnode2 ... done
Creating bnbrest  ... done
Creating bnbnode  ... done
Creating kvdnode3 ... done
Creating oracle   ... done
Creating kvdnode0 ... done
Creating deputy   ... done
```

To get the logs of a particular container (in this example kvdnode0):

```sh
docker logs kvdnode0
```

### Interacting with the local testnet

It's helpful to write aliases around docker commands that make interacting with the containers more similar to interacting with the regular command line tools.

For example, to interact with  `kvlci` on `kvdnode0` you can use the following alias

```sh
dkvcli='docker exec -it kvdnode0 /kvd/linux/kvcli'
```

Running `dkvcli config keyring-backend test` will set the keyring backend for testnet, for example. To get the sync status of the chain, run `dkvcli status`.

Similarly, for `tbnbcli`, you can use the following alias

```sh
dtbnbcli='docker exec -it bnbnode ./tbnbcli'
```

### Local testnet shutdown

To shutdown the local testnet run:

```sh
make localnet-stop && make localnet-reset
```

## testnet-4000

kava-testnet-4000 introduces the cdp and auction modules, allowing users to create CDPs and mint usdx as well as participate in auctions.

Configuration and rest-server interaction files are available in the [testnet-4000](./testnet-4000/README.md) directory.

## testnet-5000

kava-testnet-5000 introduces the bep3 modules, allowing users to transfer BNB between the bnbchain testnet and kava.

Configuration and rest-server interaction files are available in the [testnet-5000](./testnet-5000/README.md) directory.

## dev

During development new features often require modifications to the genesis file. The dev directory houses these genesis files.
