
# Contrib

## Requests

### Create CDP example request
 
 First, query account information for the signing account. Note the 'accountnumber' and 'sequence' fields, we'll need them later in order to send our request:

```bash
    kvcli q auth account $(kvcli keys show accB -a)
```

If testing locally, start the Kava rest server:

```bash
    kvcli rest-server
```

 Format the base request in create-cdp.json. You'll need to update the 'from', 'chain-id', 'account_number', 'sequence', and 'gas' as appropriate. Then, populate the CDP creation request's params 'owner', 'collateral', and 'principal'. An example formatted base request can be found in `example-create-cdp.json`.
   
 Now we'll create an unsigned request, sign it, and broadcast it to the Kava blockchain via the rest server.
 Note that if you're using the mainnet or testnet, the host IP address will need to be updated to point at an active rest server.
   
```bash
    # Create an unsigned request
    curl -H "Content-Type: application/json" -X PUT -d @./contrib/requests/create-cdp.json http://127.0.0.1:1317/cdp | jq > ./contrib/requests/create-cdp-unsigned.json

    # Sign the request
    kvcli tx sign ./contrib/requests/create-cdp-unsigned.json --from accB --offline --chain-id testing --sequence 1 --account-number 2 > ./contrib/requests/broadcast-create-cdp.json

    # Broadcast the request
    kvcli tx broadcast ./contrib/requests/broadcast-create-cdp.json
```   

Congratulations, you've just created a CDP on Kava using the rest server!

## Governance proposals

Example governance proposals are located in `/proposal_examples`.