# kava-8 Validator Update Guide
The kava-8 update includes new features. This document contains important information about the new functionality and breaking changes.

kava-8 will use the same major version of the cosmos-sdk (v0.39.x).  kava-8 will have the same golang compatibility as kava-7, requiring v1.13+. Golang v1.15 has been tested and is suitable for use on kava-8 mainnet.

## Migration Procedure

The specific steps to migrate your node can be found [here](https://github.com/Kava-Labs/kava/blob/master/migrate/v0_15/migrate.md).

## Breaking Changes

#### Claiming Rewards
Reward claims in the `x/incentive` module have been updated to enable selective reward claiming by type and token denom. There are four primary reward claim message types:

- **MsgClaimUSDXMintingReward** has arguments Sender (sdk.AccAddress) and MultiplierName (string). This message will claim all USDX minting rewards for the user, applying the specified reward multiplier.
- **MsgClaimHardReward** takes three arguments: Sender (sdk.AccAddress), Multiplier Name (string), and Denoms To Claim ([]string). This message will claim any available HARD supply/borrow rewards of the specified denoms for the user, applying the specified reward multiplier.
- **MsgClaimDelegatorReward** takes three arguments: Sender (sdk.AccAddress), Multiplier Name (string), and Denoms To Claim ([]string). This message will claim any available delegation rewards of the specified denoms for the user, applying the specified reward multiplier.
- **MsgClaimSwapReward** takes three arguments: Sender (sdk.AccAddress), Multiplier Name (string), and Denoms To Claim ([]string). This message will claim any available SWAP protocol rewards of the specified denoms for the user, applying the specified reward multiplier.

#### Committee Voting
Voting in the `x/committee` module has been updated to support Yes, No, and Abstain votes.
- **MsgVote** takes three arguments: Proposal ID (uint64), Voter (sdk.AccAddress), and Vote Type (VoteType). Valid Vote Types are “yes”, “y,” “no”, “n”, “abstain”, and “a”.

## New Features
#### SWAP Protocol
Kava-8 introduces SWAP protocol, a decentralized exchange that enables users to swap tokens against liquidity pools. SWAP protocol has several new messages:


- **MsgDeposit** enables liquidity providers to deposit tokens into a liquidity pool. It takes five arguments - the depositor (sdk.AccAddress) that is providing liquidity, token A (sdk.Coin) that is being deposited, token B (sdk.Coin) that is being deposited, minimum acceptable slippage (sdk.Dec), and deadline (int64). After a successful deposit the liquidity provider will be credited deposit shares in the pool.
- **MsgWithdraw** enables liquidity providers to withdraw deposited tokens from a liquidity pool. It takes five arguments - the withdrawer’s address (sdk.AccAddress), the amount of shares (sdk.Int) to be withdrawn, the minimum accepted token A amount (sdk.Coin) to be received by the withdrawer, the minimum accepted token B amount (sdk.Coin) to be received, and the deadline (int64). After a successful withdrawal the liquidity provider will receive tokens.
- **MsgSwapExactForTokens** supports token swaps with an exact amount of input tokens. It takes five arguments - the requester’s address (sdk.AccAddress), the exact input amount of token A (sdk.Coin), the desired output amount of token B (sdk.Coin), the minimum accepted slippage (sdk.Dec) i.e. percentage difference from the output amount, and the deadline (int64).
- **MsgSwapForExactTokens** supports token swaps for an exact amount of output tokens. It takes five arguments - the requester’s address (sdk.AccAddress), the desired input amount of token A (sdk.Coin), the exact output amount of token B (sdk.Coin), the minimum accepted slippage (sdk.Dec) i.e. percentage difference from the input amount, and the deadline (int64).

#### Committee
kava-8 introduces HARD and SWAP protocol governance by token holders via two new committees. Both committees have a proposal voting duration of seven days, a minimum quorum of 33%, and enact proposals that receive over 50% Yes votes.

The **HARD Governance Committee** will have permissions to change the following parameters:
- Hard module permissions
    - *Money markets*: whitelist of supported money markets.
    - *Minimum borrow USD value*: minimum valid borrow amount in USD from a money market.
- Incentive module permissions
    - *Hard supply reward periods*: HARD token rewards for HARD protocol suppliers.
    - *Hard borrow reward periods*: HARD token rewards for HARD protocol borrowers.
    - *Delegator reward periods*: HARD token rewards for KAVA delegators

The **SWP Governance Committee** will have permissions to change the following parameters:
- Swap module permissions
    - *Allowed pools*: whitelist of supported token pairs.
    - *Swap fee*: global trading fee paid by users to liquidity providers.
- Incentive module permissions
    - *Swap reward periods*: SWP token rewards for SWAP protocol liquidity providers.
    - *Delegator reward periods*: SWP token rewards for KAVA delegators.

## Kava REST API
Kava’s REST API supports all Kava-8 changes and features. To minimize compatibility issues, commonly used endpoints are still supported and have not been deprecated. API documentation can be found here.
Testing

Kava-testnet-13000 is a publicly available testnet (http://app.swap-testnet.kava.io/) to test validation and external integrations. Publicly available REST and RPC endpoints are:

#### Pruning nodes:
- Rest:  https://api.testnet.kava.io/node_info
- RPC: https://rpc.testnet.kava.io

#### Archive nodes:
- Rest: https://api.data-testnet.kava.io/node_info
- RPC: https://rpc.data-testnet.kava.io


## Questions/Feedback
Please reach out in your preferred communication channel (Discord, Slack, email) with any questions, or ask in [The Kava Platform Telegram](https://t.me/kavalabs).
