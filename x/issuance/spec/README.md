<!--
order: 0
title: "Issuance Overview"
parent:
  title: "issuance"
-->

# `issuance`

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Params](05_params.md)**
6. **[BeginBlock](06_begin_block.md)**

## Abstract

`x/issuance` is an implementation of a Cosmos SDK Module that allows for an issuer to control the minting and burning of an asset. The issuer is a white-listed address, which may be a multisig address, that can mint and burn coins of a particular denomination. The issuer is the only entity that can mint or burn tokens of that asset type (i.e., there is no outside inflation or deflation). Additionally, the issuer can place transfer-restrictions on addresses, which prevents addresses from transferring and holding tokens with the issued denomination. Coins are minted and burned at the discretion of the issuer, using `MsgIssueTokens` and `MsgRedeemTokens` transaction types, respectively.
