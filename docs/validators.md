# Validator Introduction

Kava is building a decentralized and secure fast finality currency designed to accelerate upcoming interoperability solutions.

We’re focused on providing a fast and resilient base currency that counters the problems existing currencies face when combined with interoperability solutions. This will enable small light clients to connect securely and quickly create and tear down payment channels on demand. Designed to keep transaction volume within acceptable bounds and not limit the global throughput of the system.

A strong set of validators is essential to this vision. The Kava blockchain is built on tendermint using cosmos proof of stake. Unlike other consensus models this places higher security and uptime demands on the validators. Validators have the opportunity to collect fees and inflation payouts, but also must ensure the network stays live and functioning well. They are ultimately responsible for how well the network runs.


## What is a validator?

Validators collect transactions submitted to the network, process them into blocks, and sign off on those blocks to add them to the global blockchain. For this work they are rewarded in tokens.

To keep blocks fast, there are a limited number of validators. The validators are picked by the protocol according to how much tokens they lock up. These can come from their own supplies or from others through a process known as delegation - where others lock up their own tokens with a validator. Validators receive rewards based on the total amount of stake they have, and are free to choose to distribute a portion of these rewards back to delegators.

To keep validators' behavior in line with the protocol, their locked tokens can be deducted if certain behaviors are detected. Tokens are deducted evenly from both their personal tokens and delegates' tokens.

Validators want to be delegated to (as it increases their rewards), so should persuade delegators that they offer a good deal for doing so. This comes down to both distributing rewards back and by maintaining a reliable validator setup that will not result in delegator's stake being slashed.


## A Validation Setup

The correct operation of the network requires overcoming several challenges. These act as the baseline in deciding how validators should ideally be setup and operated. The top challenges are listed here with the recommended methods of mitigation.

#### Chain Halting
Tendermint will halt if enough validators go offline (by malicious or accidental means). Therefore a validator's stake will be slashed if they do not sign blocks.
Validators should maintain a high availability compute setup (ie with redundant failover, located on high availability infrastructure). They should also be resilient to DoS attacks. The recommended pattern is to use a [sentry node architecture](https://forum.cosmos.network/t/sentry-node-architecture-overview/454), where the real validator (and its IP address) is shielded from the open internet.

#### Double Signing
A validator should never sign more than one block for a given height. This is indicative of byzantine behavior and will be slashed harshly.
Using a key management system is suggested to make sure that failover nodes don't result in accidental signatures. A KMS is being developed for tendermint [here](https://github.com/tendermint/kms).

#### Private Key Storage
A validator's private key is used to sign blocks and must not be compromised. Signatures from two thirds of validators defines truth in the blockchain. The network must be secure against hostile take over of validators.
The industry recommendation for secure private key storage is to use a hardware security module; a dedicated hardware device for both storing keys and signing data. Dedicated server hardware is also recommended for running a validator, along with controlled physical access. A secure collocation facility is recommended.

#### Further Reading
For a security analysis, see the work kindly provided to the community by [Bubowerks](https://bubowerks.io):
 - [Risk Assessment](https://bubowerks.io/blog/2018/08/03/risk-assessment-of-cosmos-tendermint-validators/)
 - [Risk Treatment](https://bubowerks.io/blog/2018/08/27/risk-treatments-for-cosmos-hub-tendermint-validator-risks/)

Validators can also checkout the [cosmos documentation](https://cosmos.network/docs/validators/overview.html#introduction) and [cosmos forum]() for additional details.

 ## Current Kava Testnet

The previous section covers what to expect as a long term validator. Right now Kava is in early testnet with the goal of establishing community coordination and communication channels. We're looking for validators to join using whatever compute they like (AWS is fine).

Our vision requires a wide and reliable set of validators to process blocks and keep the network running. To achieve this we are incentivizing the setup of this network to ensure it reaches the decentralization and security requirements a viable currency needs. An incentive program will be released in the near future.

We'll be running tests and ramping up validator security over the course of our testnets, with mainnet launching after a satisfactory level of validator resilience has been achieved.