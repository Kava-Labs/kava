<!--
order: 1
-->

# Concepts

 The BEP3 module implements the [BEP3 protocol](https://github.com/binance-chain/BEPs/blob/master/BEP3.md) for secure cross-chain asset transfers between Kava and other BEP3 compatible chains, such as Binance Chain. Transactions are witnessed and relayed between the two blockchains by Binance's BEP3 deputy process. The deputy maintains an address on both chains and is responsible for delivering tokens upon the successful completion of an Atomic Swap. Learn more about the BEP3 deputy process [here](https://github.com/binance-chain/bep3-deputy).

## Requirements
Kava
- The deputy’s Kava testnet-5000 address is **kava1aphsdnz5hu2t5ty2au6znprug5kx3zpy6zwq29**.
- We recommend using http://kava-testnet-5000.kava.io:1317 as Kava’s API endpoint.

Binance Chain
- The deputy’s Binance Chain testnet address is **tbnb1et8vmd0dgvswjnyaf73ez8ye0jehc8a7t7fljv**.
- We recommend using https://testnet-dex.binance.org/ as Binance Chain’s API endpoint.

Kava's [JavaScript SDK](https://github.com/Kava-Labs/javascript-sdk) and Binance Chain’s [JavaScript SDK](https://github.com/binance-chain/javascript-sdk) can be used to create, claim, and refund swaps.

## Binance Chain to Kava

When a user wants to transfer tokens from Binance Chain to Kava, the following steps are taken:
1. User’s tokens are locked on Binance Chain along with the hash of a secret only known to the user. If the secret is not revealed before the deadline, the tokens are refundable.
2. The deputy sends a message to Kava saying “a user has locked X tokens, if their secret is revealed before the deadline issue them an equivalent amount of pegged tokens”.
3. The user reveals the secret on Kava and receives the pegged tokens.
4. The deputy relays the secret to Binance Chain and the original tokens are locked permanently.   


![Binance Chain to Kava Diagram](./diagrams/BEP3_binance_chain_to_kava.jpg)

## Kava to Binance Chain
1. When a user wants to transfer tokens from Kava to Binance Chain by redeeming pegged tokens, the following steps are taken:
User’s pegged tokens are locked on Kava along with the hash of a secret only known to the user. If the secret is not revealed before the deadline, the tokens are refundable.
2. The deputy sends a message to Binance Chain saying “a user has locked X pegged tokens, if their secret is revealed before the deadline issue them an equivalent amount of tokens”.
3. The user reveals the secret on Binance Chain and receives the tokens.
4. The deputy relays the secret to Kava and the pegged tokens are locked permanently.   


![Kava to Binance Chain Diagram](./diagrams/BEP3_kava_to_binance_chain.jpg)

