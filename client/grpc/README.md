# Kava gRPC Client

The Kava gRPC client is a tool for making gRPC queries on a Kava chain.

## Features

- Easy-to-use gRPC client for the Kava chain.
- Access all query clients for Cosmos and Kava modules using `client.Query` (e.g., `client.Query.Bank.Balance`).
- Utilize utility functions for common queries (e.g., `client.BaseAccount(str)`).

## Usage

### Creating a new client

```go
package main

import (
  kavaGrpc "github.com/kava-labs/kava/client/grpc"
)
grpcUrl := "https://grpc.kava.io:443"
client, err := kavaGrpc.NewClient(grpcUrl)
if err != nil {
  panic(err)
}
```

### Making grpc queries

Query clients for both Cosmos and Kava modules are available via `client.Query`.

Example: Query Cosmos module `x/bank` for address balance

```go
import (
  banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

rsp, err := client.Query.Bank.Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: "kava19rjk5qmmwywnzfccwzyn02jywgpwjqf60afj92",
		Denom:   "ukava",
	})
```

Example: Query Kava module `x/evmutil` for params

```go
import (
  evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

rsp, err := client.Query.Evmutil.Params(
  context.Background(), &evmutiltypes.QueryParamsRequest{},
)
```

#### Query Utilities

Utility functions for common queries are available directly on the client.

Example: Util query to get a base account

```go
kavaAcc := "kava19rjk5qmmwywnzfccwzyn02jywgpwjqf60afj92"
rsp, err := client.BaseAccount(kavaAcc)
if err != nil {
  panic(err)
}
fmt.Printf("account sequence for %s: %d\n", kavaAcc, rsp.Sequence)
```

## Query Tests

To test queries, a Kava node is required. Therefore, the e2e tests for the gRPC client queries can be found in the `tests/e2e` directory. Tests for new utility queries should be added as e2e tests under the `test/e2e` directory.
