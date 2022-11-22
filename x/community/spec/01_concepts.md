<!--
order: 1
-->

# Concepts

## Community Pool

The community pool is the module account of the x/community module. It replaces the functionality of the community pool fee collector account of the auth module in the vanilla SDK.

### Funding

The community pool can be funded every block from the community pool inflation of the x/kavamint module.

Additionally, the pool can be funded by any account sending a community/FundCommunityPool message.

### Spending

The community pool funds are spent via government proposals. The x/kavadist module includes a CommunityPoolMultiSpendProposal that, upon approval, distributes funds to a list of accounts.
