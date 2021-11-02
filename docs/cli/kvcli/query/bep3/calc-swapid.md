<!--
title: calc-swapid
-->
## kvcli query bep3 calc-swapid

calculate swap ID for the given random number hash, sender, and sender other chain

```
kvcli query bep3 calc-swapid [random-number-hash] [sender] [sender-other-chain] [flags]
```

### Examples

```
bep3 calc-swapid 0677bd8a303dd981810f34d8e5cc6507f13b391899b84d3c1be6c6045a17d747 kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny bnb1ud3q90r98l3mhd87kswv3h8cgrymzeljct8qn7
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for calc-swapid
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

