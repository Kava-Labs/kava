<!--
title: start
-->
## kvd start

Run the full node

### Synopsis

Run the full node application with Tendermint in or out of process. By
default, the application will run with Tendermint in process.

Pruning options can be provided via the '--pruning' flag or alternatively with '--pruning-keep-recent',
'pruning-keep-every', and 'pruning-interval' together.

For '--pruning' the options are as follows:

default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: all saved states will be deleted, storing only the current state; pruning at 10 block intervals
custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'. During
the ABCI Commit phase, the node will check if the current block height is greater than or equal to
the halt-height or if the current block time is greater than or equal to the halt-time. If so, the
node will attempt to gracefully shutdown and the block will not be committed. In addition, the node
will not be able to commit subsequent blocks.

For profiling and benchmarking purposes, CPU profiling can be enabled via the '--cpu-profile' flag
which accepts a path for the resulting pprof file.


```
kvd start [flags]
```

### Options

```
      --abci string                                     Specify abci transport (socket | grpc) (default "socket")
      --address string                                  Listen address (default "tcp://0.0.0.0:26658")
      --consensus.create_empty_blocks                   Set this to false to only produce blocks when there are txs or when the AppHash changes (default true)
      --consensus.create_empty_blocks_interval string   The possible interval between empty blocks (default "0s")
      --cpu-profile string                              Enable CPU profiling and write to the provided file
      --db_backend string                               Database backend: goleveldb | cleveldb | boltdb | rocksdb (default "goleveldb")
      --db_dir string                                   Database directory (default "data")
      --fast_sync                                       Fast blockchain syncing (default true)
      --genesis_hash bytesHex                           Optional SHA-256 hash of the genesis file
      --halt-height uint                                Block height at which to gracefully halt the chain and shutdown the node
      --halt-time uint                                  Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node
  -h, --help                                            help for start
      --inter-block-cache                               Enable inter-block caching (default true)
      --mempool.authorized-addresses strings            Additional addresses to accept transactions from when the mempool is running in authorized mode (comma separated kava addresses)
      --mempool.enable-authentication                   Configure the mempool to only accept transactions from authorized addresses
      --minimum-gas-prices string                       Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)
      --moniker string                                  Node Name (default "MacBook-Pro.home")
      --p2p.laddr string                                Node listen address. (0.0.0.0:0 means any interface, any port) (default "tcp://0.0.0.0:26656")
      --p2p.persistent_peers string                     Comma-delimited ID@host:port persistent peers
      --p2p.pex                                         Enable/disable Peer-Exchange (default true)
      --p2p.private_peer_ids string                     Comma-delimited private peer IDs
      --p2p.seed_mode                                   Enable/disable seed mode
      --p2p.seeds string                                Comma-delimited ID@host:port seed nodes
      --p2p.unconditional_peer_ids string               Comma-delimited IDs of unconditional peers
      --p2p.upnp                                        Enable/disable UPNP port forwarding
      --priv_validator_laddr string                     Socket address to listen on for connections from external priv_validator process
      --proxy_app string                                Proxy app address, or one of: 'kvstore', 'persistent_kvstore', 'counter', 'counter_serial' or 'noop' for local testing. (default "tcp://127.0.0.1:26658")
      --pruning string                                  Pruning strategy (default|nothing|everything|custom) (default "default")
      --pruning-interval uint                           Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')
      --pruning-keep-every uint                         Offset heights to keep on disk after 'keep-every' (ignored if pruning is not 'custom')
      --pruning-keep-recent uint                        Number of recent heights to keep on disk (ignored if pruning is not 'custom')
      --rpc.grpc_laddr string                           GRPC listen address (BroadcastTx only). Port required
      --rpc.laddr string                                RPC listen address. Port required (default "tcp://127.0.0.1:26657")
      --rpc.unsafe                                      Enabled unsafe rpc methods
      --trace                                           Provide full stack traces for errors in ABCI Log
      --trace-store string                              Enable KVStore tracing to an output file
      --unsafe-skip-upgrades ints                       Skip a set of upgrade heights to continue the old binary
      --with-tendermint                                 Run abci app embedded in-process with tendermint (default true)
```

### Options inherited from parent commands

```
      --inv-check-period uint   Assert registered invariants every N blocks
      --log_level string        Log level (default "main:info,state:info,*:error")
```

