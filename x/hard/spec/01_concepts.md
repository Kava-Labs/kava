<!--
order: 1
-->

# Concepts

## Automated, Cross-Chain Money Markets

The hard module provides for functionality and governance of a two-sided money market protocol with autonomous interest rates. The main state transitions in the hard module are composed of deposit, withdraw, borrow and repay actions. Borrow positions can be liquidated by an external party called a "keeper". Keepers receive a fee in exchange for liquidating risk positions, and the fee rate is determined by governance. Internally, all funds are stored in a module account (the cosmos-sdk equivalent of the `address` portion of a smart contract), and can be accessed via the above actions. Each money market has governance parameters which are controlled by token-holder governance. Of particular note are the interest rate model, which determines (using a static formula) what the prevailing rate of interest will be for each block, and the loan-to-value (LTV), which determines how much borrowing power each unit of deposited collateral will count for. Initial parameterization of the hard module will stipulate that all markets are over-collateralized and that overall borrow limits for each collateral will start small and rise gradually.

## HARD Token distribution

[See Incentive Module](../../incentive/spec/01_concepts.md)
