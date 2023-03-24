<!--
order: 1
-->

# Concepts

This module is responsible for the minting and burning of liquid staking receipt tokens, collectively referred to as `bkava`. Delegated kava can be converted to delegator-specific `bkava`. Ie, 100 KAVA delegated to validator `kavavaloper123` can be converted to 100 `bkava-kavavaloper123`. Similarly, 100 `bkava-kavavaloper123` can be converted back to a delegation of 100 KAVA to  `kavavaloper123`. In this design, all validators can permissionlessly participate in liquid staking while users retain the delegator specific slashing risk and voting rights of their original validator. Note that because each `bkava` denom is validator specific, this module does not specify a fungibility mechanism for `bkava` denoms. 