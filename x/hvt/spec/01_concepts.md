<!--
order: 1
-->

# Concepts

The harvest module introduces the hard token to the kava blockchain. This module distributes hard tokens to two types of ecosystem participants:

1. Kava stakers - any address that stakes (delegates) kava tokens will be eligible to claim hard tokens. For each delegator, hard tokens are accumulated ratably based on the total number of kava tokens staked. For example, if a user stakes 1 million KAVA tokens and there are 100 million staked KAVA, that user will accumulate 1% of hard tokens earmarked for stakers during the distribution period. Distribution periods are defined by a start date, an end date, and a number of hard tokens that are distributed per second.
2. Depositors - any address that deposits eligible tokens to the harvest module will be eligible to claim hard tokens. For each depositor, hard tokens are accumulated ratable based on the total number of tokens staked of that denomination. For example, if a user deposits 1 million "xyz" tokens and there are 100 million xyz deposited, that user will accumulate 1% of hard tokens earmarked for depositors of that denomination during the distribution period. Distribution periods are defined by a start date, an end date, and a number of hard tokens that are distributed per second.

Users are not air-dropped tokens, rather they accumulate `Claim` objects that they my submit a transaction in order to claim. In order to better align long term incentives, when users claim hard tokens, they have three options, called 'multipliers', for how tokens are distributed.

* Liquid - users can immediately receive hard tokens, but they will receive a smaller fraction of tokens than if they choose medium-term or long-term locked tokens.
* Medium-term locked - users can receive tokens that are medium-term transfer restricted. They will receive more tokens than users who choose liquid tokens, but fewer than those who choose long term locked tokens.
* Long-term locked - users can receive tokens that are long-term transfer restricted. Users choosing this option will receive more tokens than users who choose liquid or medium-term locked tokens.

The exact multipliers will be voted by governance and can be changed via a governance vote. An example multiplier schedule would be:

* Liquid - 10% multiplier and no lock up. Users receive 10% as many tokens as users who choose long-term locked tokens.
* Medium-term locked - 33% multiplier and 6 month transfer restriction. Users receive 33% as many tokens as users who choose long-term locked tokens.
* Long-term locked - 100% multiplier and 2 year transfer restriction. Users receive 10x as many tokens as users who choose liquid tokens and 3x as many tokens as users who choose medium-term locked tokens.