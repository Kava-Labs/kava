<!--
title: testnet
-->
## kvd testnet

Initialize files for a local kava testnet

### Synopsis

testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	$ kvd testnet --v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	

```
kvd testnet [flags]
```

### Options

```
      --chain-id string              genesis file chain-id, if left blank will be randomly created
  -h, --help                         help for testnet
      --keyring-backend string       Select keyring's backend (os|file|test) (default "os")
      --minimum-gas-prices string    Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake) (default "0.000006stake")
      --node-cli-home string         Home directory of the node's cli configuration (default "kvcli")
      --node-daemon-home string      Home directory of the node's daemon configuration (default "kvd")
      --node-dir-prefix string       Prefix the directory name for each node with (node results in node0, node1, ...) (default "node")
  -o, --output-dir string            Directory to store initialization data for the testnet (default "./kavatestnet")
      --starting-ip-address string   Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...) (default "192.168.0.1")
      --v int                        Number of validators to initialize the testnet with (default 4)
```

### Options inherited from parent commands

```
      --inv-check-period uint   Assert registered invariants every N blocks
      --log_level string        Log level (default "main:info,state:info,*:error")
```

