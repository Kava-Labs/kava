# Testnet-5000

Testnet-5000 introduces transfers between Kava and Bnbchain via BEP3.

This guide will walk you through interacting with the blockchains and transferring coins via the rest server. To send transactions, we'll create an unsigned request, sign it, and broadcast it to the Kava blockchain.

## Rest server requests

### Setup

TODO: Add testnet REST endpoints for [Kava] and [Bnbchain] and replace "localhost" with live endpoints.

TODO: Steps for downloading kvcli

Before making a request, query account information for the signing account. Note the 'accountnumber' and 'sequence' fields, we'll need them later in order to send our request:

```bash
    kvcli q auth account $(kvcli keys show testuser -a)
```

### Create swap

Use the example file in `rest_examples/create-swap.json` to format the request. First, update the header parameters 'from', 'chain-id', 'account_number', 'sequence'.

Next, we'll update the swap's creation parameters. For that, we need a unique random number that will be used to claim the funds.

// TODO: Explain how to generate these client-side

```bash
    # Generate a sample random number, timestamp, and random number hash
    kvcli q bep3 calc-rnh now

    # Expected output:
    # Random number: 110802331073994018312675691928205725441742309715720953510374321628333109608728
    # Timestamp: 1585203985
    # Random number hash: 4644fc2d9a2389c60e621785b873ae187e320eaded1687edaa120961428eba9e
```

// TODO: add limits
In the same json file, populate each parameter within to the following limits

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
    kvcli tx sign ./contrib/testnet-5000/rest_examples/create-swap-unsigned.json --from testuser --offline --chain-id testing --sequence 0 --account-number 5 | jq > ./contrib/testnet-5000/rest_examples/broadcast-create-swap.json

    # Broadcast the request
    kvcli tx broadcast ./contrib/testnet-5000/broadcast-create-swap.json
```

// TODO: They need the expected bnbchain swap ID

Congratulations, you've just created a swap on Kava! The swap will be automatically relayed over to Bnbchain where it it can be claimed using the secret random number from above.

// TODO: Add link to a doc with steps to create, claim, and refund swaps on bnbchain

# Claim swap

// TODO: add link to Bnbchain document with interaction steps
Generally, claimable swaps must be created on Bnbchain.

Use the example file in `rest_examples/claim-swap.json` to format the request. Again, update the header parameters 'from', 'account_number', 'sequence'. Check your account using the command from above to ensure that the parameters match the blockchain's state.

// TODO: add limits
In the same json file, populate each parameter within to the following limits:

- swap_id: only unexpired swaps can be claimed.
- random_number:

Once the `swap_id` parameter is populated, it's time to claim our swap:

```bash
    # Create an unsigned request
    curl -H "Content-Type: application/json" -X POST -d @./contrib/testnet-5000/rest_examples/claim-swap.json http://127.0.0.1:1317/bep3/swap/claim | jq > ./contrib/testnet-5000/rest_examples/claim-swap-unsigned.json

    # Sign the request
    kvcli tx sign ./contrib/testnet-5000/rest_examples/claim-swap-unsigned.json --from testuser --offline --chain-id testing --sequence 1 --account-number 1 | jq  > ./contrib/testnet-5000/rest_examples/broadcast-claim-swap.json

    # Broadcast the request
    kvcli tx broadcast ./contrib/testnet-5000/rest_examples/broadcast-claim-swap.json
```

# Refund swap

Use the example file in `rest_examples/refund-swap.json` to format the request. Again, update the header parameters 'from', 'account_number', 'sequence'. Check your account using the command from above to ensure that the parameters match the blockchain's state.

// TODO: add limits
In the same json file, populate each parameter within to the following limits:

- swap_id: only expired swaps can be refunded.

Once the `swap_id` parameter is populated, it's time to refund our swap:

```bash
    # Create an unsigned request
    curl -H "Content-Type: application/json" -X POST -d @./contrib/testnet-5000/rest_examples/refund-swap.json http://127.0.0.1:1317/bep3/swap/refund | jq > ./contrib/testnet-5000/rest_examples/refund-swap-unsigned.json

    # Sign the request
    kvcli tx sign ./contrib/testnet-5000/rest_examples/refund-swap-unsigned.json --from user --offline --chain-id testing --sequence 0 --account-number 1 | jq  > ./contrib/testnet-5000/rest_examples/broadcast-refund-swap.json

    # Broadcast the request
    kvcli tx broadcast ./contrib/testnet-5000/rest_examples/broadcast-refund-swap.json
```
