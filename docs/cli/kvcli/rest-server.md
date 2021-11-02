<!--
title: rest-server
-->
## kvcli rest-server

Start LCD (light-client daemon), a local REST server

```
kvcli rest-server [flags]
```

### Options

```
      --height int           Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                 help for rest-server
      --indent               Add indent to JSON response
      --laddr string         The address for the server to listen on (default "tcp://localhost:1317")
      --ledger               Use a connected Ledger device
      --max-open uint        The number of maximum open connections (default 1000)
      --node string          <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --read-timeout uint    The RPC read timeout (in seconds) (default 10)
      --trust-node           Trust connected full node (don't verify proofs for responses)
      --unsafe-cors          Allows CORS requests from all domains. For development purposes only, use it at your own risk.
      --write-timeout uint   The RPC write timeout (in seconds) (default 10)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

