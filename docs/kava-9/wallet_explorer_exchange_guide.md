# Kava-9 Upgrade Guide for Wallets, Explorers, and Exchanges

The Kava-9 Upgrade migrates the Kava blockchain from v0.39.x of the cosmos-sdk to v0.44.x of the cosmos-sdk. A full description of the v0.39 -> v0.40-42 REST changes can be found at:

https://github.com/cosmos/cosmos-sdk/blob/v0.42.11/docs/migrations/rest.md  

**Note**: Kava engineers have implemented a custom POST /txs endpoint to preserve backwards compatibility (v0.42 -> v0.44  REST changes) for legacy encoded transaction:

https://docs.cosmos.network/master/migrations/rest.html 

**Note**: Cosmos, Terra, and Osmosis blockchains have also been upgraded to v0.44 of the cosmos-sdk. Changes needed to support those chains will apply to Kava as well.

For wallets, explorers, and exchanges, there are a few particular changes to be aware of:

### Account data
Account data has been separated from balances data. In v0.15 of kava (kava-8), querying an account would return the following JSON, which contains both the account data and balance data:

```sh
curl /auth/accounts/kava173w2zz287s36ewnnkf4mjansnthnnsz7rtrxqc`
```
```json
{"height":"38","result":{"type":"cosmos-sdk/Account","value":{"address":"kava173w2zz287s36ewnnkf4mjansnthnnsz7rtrxqc","coins":[{"denom":"ukava","amount":"1000000000000"}],"public_key":null,"account_number":"16","sequence":"0"}}}`
```

In v0.16 (kava-9):


```sh 
curl /auth/accounts/kava173w2zz287s36ewnnkf4mjansnthnnsz7rtrxqc`
```
```json
{"height":"6","result":{"type":"cosmos-sdk/BaseAccount","value":{"address":"kava173w2zz287s36ewnnkf4mjansnthnnsz7rtrxqc","account_number":"15"}}}`
```


Note: The `sequence` field is now omitted from the response if an account has not signed a transaction.

To get the balance data, the /bank/balances endpoint must be used in addition to /auth/accounts:

```sh
curl /bank/balances/kava173w2zz287s36ewnnkf4mjansnthnnsz7rtrxqc`
```
```json
{"height":"198","result":[{"denom":"ukava","amount":"10000000000000000"}]}`
```


### Vesting Account Data
The address, account sequence and account number fields for periodic vesting accounts have moved:

In v0.15 (kava-8):


```sh
curl /auth/accounts/kava1z3ytjpr6ancl8gw80z6f47z9smug7986x29vtj
```
```json
{"height":"3","result":{"type":"cosmos-sdk/PeriodicVestingAccount","value":{"address":"kava1z3ytjpr6ancl8gw80z6f47z9smug7986x29vtj","coins":[{"denom":"ukava","amount":"565077579"},{"denom":"usdx","amount":"1363200"}],"public_key":null,"account_number":"6","sequence":"0","original_vesting":[{"denom":"ukava","amount":"560159828"}],"delegated_free":[],"delegated_vesting":[],"end_time":"1628213878","start_time":"1596677878","vesting_periods":[{"length":"31536000","amount":[{"denom":"ukava","amount":"560159828"}]}]}}}
```


**Note**: address, account sequence and account number can be found at 
- `.result.value.address`
- `.result.value.sequence`
- `.result.value.account_number`

In v0.16 (kava-9):

```sh
curl /auth/accounts/kava1z3ytjpr6ancl8gw80z6f47z9smug7986x29vtj`
```

```json
{"height":"4059","result":{"type":"cosmos-sdk/PeriodicVestingAccount","value":{"base_vesting_account":{"base_account":{"address":"kava1fwfwmt6vupf3m9uvpdsuuc4dga8p5dtl4npcqz","public_key":{"type":"tendermint/PubKeySecp256k1","value":"A3CJ0ejMGhGhxC9dRqKooEkiOj++kMh+lFDbdN283QHE"},"account_number":"18","sequence":"2"},"original_vesting":[{"denom":"ukava","amount":"560159828"}],"delegated_free":[],"delegated_vesting":[],"end_time":"1664632800"},"start_time":"1633096800","vesting_periods":[{"length":"31536000","amount":[{"denom":"ukava","amount":"560159828"}]}]}}}`
```

**Note**: address, account sequence and account number can be found at 
- `.result.value.base_vesting_account.base_account.address`
- `.result.value.base_vesting_account.base_account.sequence`
- `.result.value.base_vesting_account.base_account.account_number`

### Legacy encoded transactions (POST /txs)
The Cosmos team found vulnerabilities in legacy transaction support and removed the [POST] /txs endpoints from v0.44.0 of the cosmos-sdk. To avoid this backwards incompatibility, kava-9 introduces a custom [POST] /txs endpoint that converts legacy (amino) transactions to the new proto encoding and broadcasts the converted tx. Implementation details can be found [here](https://github.com/Kava-Labs/kava/pull/1070).

### Javascript-SDK

The Kava Javascript-SDK has been updated to support kava-9, including the new [POST] /txs endpoint. This means that applications utilizing the Javascript-SDK can simply update:

`npm i @kava-labs/javascript-sdk@latest`

