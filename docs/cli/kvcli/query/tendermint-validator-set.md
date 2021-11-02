<!--
title: tendermint-validator-set
-->
## kvcli query tendermint-validator-set

Get the full tendermint validator set at given height

```
kvcli query tendermint-validator-set [height] [flags]
```

### Options

```
  -h, --help          help for tendermint-validator-set
      --indent        indent JSON response
      --limit int     Query number of results returned per page (default 100)
  -n, --node string   Node to connect to (default "tcp://localhost:26657")
      --page int      Query a specific page of paginated results
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

