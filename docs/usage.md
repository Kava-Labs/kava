# Basic Usage
>The following commands require communication with a full node. By default they expect one to be running locally (accessible on localhost), but a remote can be provided with the `--node` flag.
## View Your Address
List locally stored account addresses and their names. The name is used in other commands to specify which address to use to sign the transaction.

    kvcli keys list

## View Account Balances

    kvcli account <address>

## Send Some KVA

    kvcli send --from <your key name> \
               --to <address> \
               --amount 100KVA


