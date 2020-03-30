# testnet-5000

testnet-5000 introduces transfers between Kava and Bnbchain via BEP3.

This guide will walk you through interacting with the blockchains and transferring coins via the rest server. To send transactions, we'll create an unsigned request, sign it, and broadcast it to the Kava blockchain.

## Rest server requests

### Setup

We'll be using Kava's CLI to build, sign, and broadcast the transactions:

```bash
    # Download kvcli
    make install
```

Before making a request, query account information for the signing account. Note the 'accountnumber' and 'sequence' fields, we'll need them later in order to send our request:

```bash
    kvcli q auth account $(kvcli keys show testuser -a)
```

### Create swap

Use the example file in `rest_examples/create-swap.json` to format the request. First, update the header parameters 'from', 'chain-id', 'account_number', 'sequence'.

Next, we'll update the swap's creation parameters. For that, we need a unique random number that will be used to claim the funds.

WARNING: Don't use `calc-rnh` for the generation of secrets in production. These values should be generated client-side for the safety of user funds.

```bash
    # Generate a sample random number, timestamp, and random number hash
    kvcli q bep3 calc-rnh now

    # Expected output:
    # Random number: 110802331073994018312675691928205725441742309715720953510374321628333109608728
    # Timestamp: 1585203985
    # Random number hash: 4644fc2d9a2389c60e621785b873ae187e320eaded1687edaa120961428eba9e
```

In the same json file, populate each of the following parameters

- from
- to
- recipient_other_chain
- sender_other_chain
- random_number_hash
- timestamp
- amount
- expected_income
- height_span
- cross_chain

Once each parameter is populated, it's time to create our swap:

```bash
    # Create an unsigned request
    curl -H "Content-Type: application/json" -X POST -d @./contrib/testnet-5000/rest_examples/create-swap.json http://127.0.0.1:1317/bep3/swap/create | jq > ./contrib/testnet-5000/rest_examples/create-swap-unsigned.json

    # Sign the request
    kvcli tx sign ./contrib/testnet-5000/rest_examples/create-swap-unsigned.json --from testnetdeputy --offline --chain-id testing --sequence 0 --account-number 5 | jq > ./contrib/testnet-5000/rest_examples/broadcast-create-swap.json

    # Broadcast the request
    kvcli tx broadcast ./contrib/testnet-5000/rest_examples/broadcast-create-swap.json
```

The tx broadcast will log information in the terminal, including the txhash. This tx hash can be used to get information about the transaction, including the swap creation event that includes the swap's ID:

```bash
    # Get information about the transaction
    curl -H "Content-Type: application/json" -X GET http://localhost:1317/txs/81A1955216F6D985ECB4770E29B9BCED8F73A42D0C0FD566372CF673CCB81587
```

Congratulations, you've just created a swap on Kava! The swap will be automatically relayed over to Bnbchain where it it can be claimed using the secret random number from above.

# Claim swap

Only unexpired swaps can be claimed. To claim a swap, we'll use the secret random number that matches this swap's timestamp and random number hash.

Generally, claimable swaps must be created on Bnbchain.
// TODO: add link to Bnbchain document with interaction steps

Use the example file in `rest_examples/claim-swap.json` to format the request. Again, update the header parameters 'from', 'account_number', 'sequence'. Check your account using the command from above to ensure that the parameters match the blockchain's state.

In the same json file, populate each listed parameter:

- swap_id
- random_number

Once the `swap_id` parameter is populated, it's time to claim our swap:

```bash
    # Create an unsigned request
    curl -H "Content-Type: application/json" -X POST -d @./contrib/testnet-5000/rest_examples/claim-swap.json http://127.0.0.1:1317/bep3/swap/claim | jq > ./contrib/testnet-5000/rest_examples/claim-swap-unsigned.json

    # Sign the request
    kvcli tx sign ./contrib/testnet-5000/rest_examples/claim-swap-unsigned.json --from user --offline --chain-id testing --sequence 0 --account-number 1 | jq  > ./contrib/testnet-5000/rest_examples/broadcast-claim-swap.json

    # Broadcast the request
    kvcli tx broadcast ./contrib/testnet-5000/rest_examples/broadcast-claim-swap.json
```

# Refund swap

Only expired swaps may be refunded.

Use the example file in `rest_examples/refund-swap.json` to format the request. Again, update the header parameters 'from', 'account_number', 'sequence'. Check your account using the command from above to ensure that the parameters match the blockchain's state.

In the same json file, populate each parameter:

- swap_id

Once the `swap_id` parameter is populated, it's time to refund our swap:

```bash
    # Create an unsigned request
    curl -H "Content-Type: application/json" -X POST -d @./contrib/testnet-5000/rest_examples/refund-swap.json http://127.0.0.1:1317/bep3/swap/refund | jq > ./contrib/testnet-5000/rest_examples/refund-swap-unsigned.json

    # Sign the request
    kvcli tx sign ./contrib/testnet-5000/rest_examples/refund-swap-unsigned.json --from user --offline --chain-id testing --sequence 0 --account-number 1 | jq  > ./contrib/testnet-5000/rest_examples/broadcast-refund-swap.json

    # Broadcast the request
    kvcli tx broadcast ./contrib/testnet-5000/rest_examples/broadcast-refund-swap.json
```
