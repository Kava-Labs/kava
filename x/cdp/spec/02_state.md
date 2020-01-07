# State

Listed here is the state stored by the module.
For details see the types package. In particular `keys.go` describes how state is stored in the key-value store.
<!--
Structs from types were not copied in here to avoid to creating future work when the structs are updated.
Similarly key structures are not listed here - they seem like an implementation detail best documented by code comments.
-->

## CDP

A CDP is a struct representing a debt position owned by one or more depositors. It has one collateral type and records the debt that has been drawn and how much fees should be repaid.

These are stored with a couple of database indexes for faster lookup:

- by collateral ratio - to look up cdps that are close to the liquidation ratio
- by owner index - to look up cdps that an address has deposited to

## Deposit

A Deposit is a struct recording collateral added to a CDP by one address. That address only has authorization to change their deposited amount.

## Params

Module parameters controlled by governance. See [Parameters](07_params.md) for details.

## NextCDPID

A global counter used to create unique CDP ids.

## DebtDenom

The denom for the internal debt coin. This is in the store as its value can be configured at genesis.

## Total Principle

Total pegged assets drawn for each pegged asset denom.
<!-- TODO is this different from the total amount of the assets created? ie does it include amount in liquidator? -->

## FeeRate

<!-- TODO -->

## Previous Block Time

A record of the last block time used in calculating fees.
<!-- TODO add details -->
