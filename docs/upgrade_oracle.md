# kava-8 Oracle Update Guide
The Kava blockchain is upgrading and minor updates are needed for oracles

## Update Guide For Standalone Oracle Operators
1. Install the latest version of kava-tools
2. In your `.env` configuration file, add an entry for `FEE`. This will set the default fee (in ukava) for each oracle transaction (can be 0).
   - `FEE=”10000”`
3. In your `.env` configuration file, update the MARKET_IDS entry to include the latest markets for kava-5:
   - `MARKET_IDS="bnb:usd,bnb:usd:30,btc:usd,btc:usd:30,xrp:usd,xrp:usd:30,busd:usd,busd:usd:30,kava:usd,kava:usd:30,hard:usd,hard:usd:30,usdx:usd"`
4. Restart your oracle process

## Update Guide For Chainlink Oracle Operators
1. Pull the latest version of Kava’s external-adapters-js repo
2. Install `yarn`
3. Build the latest version of the kava adapter
   - from top level external-adapter-js directory
   - make docker adapter=kava
4. Edit your configuration file and add an entry for `FEE`. This will set the default fee (in ukava) for each oracle transaction (can be 0).
   - `FEE=”10000”`
5. Restart the kava adapter with the latest version
6. If necessary, create jobs for the following market_ids, if they do not already exist
   - bnb:usd
   - bnb:usd:30
   - btc:usd
   - btc:usd:30
   - xrp:usd
   - xrp:usd:30
   - busd:usd
   - busd:usd:30
   - kava:usd
   - kava:usd:30
   - hard:usd
   - hard:usd:30
   - usdx:usd

