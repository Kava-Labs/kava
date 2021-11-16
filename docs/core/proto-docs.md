 <!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [kava/bep3/v1beta1/bep3.proto](#kava/bep3/v1beta1/bep3.proto)
    - [AssetParam](#kava.bep3.v1beta1.AssetParam)
    - [AssetSupply](#kava.bep3.v1beta1.AssetSupply)
    - [AtomicSwap](#kava.bep3.v1beta1.AtomicSwap)
    - [Params](#kava.bep3.v1beta1.Params)
    - [SupplyLimit](#kava.bep3.v1beta1.SupplyLimit)
  
    - [SwapDirection](#kava.bep3.v1beta1.SwapDirection)
    - [SwapStatus](#kava.bep3.v1beta1.SwapStatus)
  
- [kava/bep3/v1beta1/genesis.proto](#kava/bep3/v1beta1/genesis.proto)
    - [GenesisState](#kava.bep3.v1beta1.GenesisState)
  
- [kava/bep3/v1beta1/query.proto](#kava/bep3/v1beta1/query.proto)
    - [AssetSupplyResponse](#kava.bep3.v1beta1.AssetSupplyResponse)
    - [AtomicSwapResponse](#kava.bep3.v1beta1.AtomicSwapResponse)
    - [QueryAssetSuppliesRequest](#kava.bep3.v1beta1.QueryAssetSuppliesRequest)
    - [QueryAssetSuppliesResponse](#kava.bep3.v1beta1.QueryAssetSuppliesResponse)
    - [QueryAssetSupplyRequest](#kava.bep3.v1beta1.QueryAssetSupplyRequest)
    - [QueryAssetSupplyResponse](#kava.bep3.v1beta1.QueryAssetSupplyResponse)
    - [QueryAtomicSwapRequest](#kava.bep3.v1beta1.QueryAtomicSwapRequest)
    - [QueryAtomicSwapResponse](#kava.bep3.v1beta1.QueryAtomicSwapResponse)
    - [QueryAtomicSwapsRequest](#kava.bep3.v1beta1.QueryAtomicSwapsRequest)
    - [QueryAtomicSwapsResponse](#kava.bep3.v1beta1.QueryAtomicSwapsResponse)
    - [QueryParamsRequest](#kava.bep3.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#kava.bep3.v1beta1.QueryParamsResponse)
  
    - [Query](#kava.bep3.v1beta1.Query)
  
- [kava/bep3/v1beta1/tx.proto](#kava/bep3/v1beta1/tx.proto)
    - [MsgClaimAtomicSwap](#kava.bep3.v1beta1.MsgClaimAtomicSwap)
    - [MsgClaimAtomicSwapResponse](#kava.bep3.v1beta1.MsgClaimAtomicSwapResponse)
    - [MsgCreateAtomicSwap](#kava.bep3.v1beta1.MsgCreateAtomicSwap)
    - [MsgCreateAtomicSwapResponse](#kava.bep3.v1beta1.MsgCreateAtomicSwapResponse)
    - [MsgRefundAtomicSwap](#kava.bep3.v1beta1.MsgRefundAtomicSwap)
    - [MsgRefundAtomicSwapResponse](#kava.bep3.v1beta1.MsgRefundAtomicSwapResponse)
  
    - [Msg](#kava.bep3.v1beta1.Msg)
  
- [kava/committee/v1beta1/committee.proto](#kava/committee/v1beta1/committee.proto)
    - [BaseCommittee](#kava.committee.v1beta1.BaseCommittee)
    - [MemberCommittee](#kava.committee.v1beta1.MemberCommittee)
    - [TokenCommittee](#kava.committee.v1beta1.TokenCommittee)
  
    - [TallyOption](#kava.committee.v1beta1.TallyOption)
  
- [kava/committee/v1beta1/genesis.proto](#kava/committee/v1beta1/genesis.proto)
    - [GenesisState](#kava.committee.v1beta1.GenesisState)
    - [Proposal](#kava.committee.v1beta1.Proposal)
    - [Vote](#kava.committee.v1beta1.Vote)
  
    - [VoteType](#kava.committee.v1beta1.VoteType)
  
- [kava/committee/v1beta1/permissions.proto](#kava/committee/v1beta1/permissions.proto)
    - [GodPermission](#kava.committee.v1beta1.GodPermission)
    - [SoftwareUpgradePermission](#kava.committee.v1beta1.SoftwareUpgradePermission)
    - [TextPermission](#kava.committee.v1beta1.TextPermission)
  
- [kava/committee/v1beta1/proposal.proto](#kava/committee/v1beta1/proposal.proto)
    - [CommitteeChangeProposal](#kava.committee.v1beta1.CommitteeChangeProposal)
    - [CommitteeDeleteProposal](#kava.committee.v1beta1.CommitteeDeleteProposal)
  
- [kava/committee/v1beta1/tx.proto](#kava/committee/v1beta1/tx.proto)
    - [MsgSubmitProposal](#kava.committee.v1beta1.MsgSubmitProposal)
    - [MsgVote](#kava.committee.v1beta1.MsgVote)
  
- [kava/issuance/v1beta1/genesis.proto](#kava/issuance/v1beta1/genesis.proto)
    - [Asset](#kava.issuance.v1beta1.Asset)
    - [AssetSupply](#kava.issuance.v1beta1.AssetSupply)
    - [GenesisState](#kava.issuance.v1beta1.GenesisState)
    - [Params](#kava.issuance.v1beta1.Params)
    - [RateLimit](#kava.issuance.v1beta1.RateLimit)
  
- [kava/issuance/v1beta1/query.proto](#kava/issuance/v1beta1/query.proto)
    - [QueryParamsRequest](#kava.issuance.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#kava.issuance.v1beta1.QueryParamsResponse)
  
    - [Query](#kava.issuance.v1beta1.Query)
  
- [kava/issuance/v1beta1/tx.proto](#kava/issuance/v1beta1/tx.proto)
    - [MsgBlockAddress](#kava.issuance.v1beta1.MsgBlockAddress)
    - [MsgBlockAddressResponse](#kava.issuance.v1beta1.MsgBlockAddressResponse)
    - [MsgIssueTokens](#kava.issuance.v1beta1.MsgIssueTokens)
    - [MsgIssueTokensResponse](#kava.issuance.v1beta1.MsgIssueTokensResponse)
    - [MsgRedeemTokens](#kava.issuance.v1beta1.MsgRedeemTokens)
    - [MsgRedeemTokensResponse](#kava.issuance.v1beta1.MsgRedeemTokensResponse)
    - [MsgSetPauseStatus](#kava.issuance.v1beta1.MsgSetPauseStatus)
    - [MsgSetPauseStatusResponse](#kava.issuance.v1beta1.MsgSetPauseStatusResponse)
    - [MsgUnblockAddress](#kava.issuance.v1beta1.MsgUnblockAddress)
    - [MsgUnblockAddressResponse](#kava.issuance.v1beta1.MsgUnblockAddressResponse)
  
    - [Msg](#kava.issuance.v1beta1.Msg)
  
- [kava/kavadist/v1beta1/params.proto](#kava/kavadist/v1beta1/params.proto)
    - [Params](#kava.kavadist.v1beta1.Params)
    - [Period](#kava.kavadist.v1beta1.Period)
  
- [kava/kavadist/v1beta1/genesis.proto](#kava/kavadist/v1beta1/genesis.proto)
    - [GenesisState](#kava.kavadist.v1beta1.GenesisState)
  
- [kava/kavadist/v1beta1/proposal.proto](#kava/kavadist/v1beta1/proposal.proto)
    - [CommunityPoolMultiSpendProposal](#kava.kavadist.v1beta1.CommunityPoolMultiSpendProposal)
    - [CommunityPoolMultiSpendProposalJSON](#kava.kavadist.v1beta1.CommunityPoolMultiSpendProposalJSON)
    - [MultiSpendRecipient](#kava.kavadist.v1beta1.MultiSpendRecipient)
  
- [kava/kavadist/v1beta1/query.proto](#kava/kavadist/v1beta1/query.proto)
    - [QueryBalanceRequest](#kava.kavadist.v1beta1.QueryBalanceRequest)
    - [QueryBalanceResponse](#kava.kavadist.v1beta1.QueryBalanceResponse)
    - [QueryParamsRequest](#kava.kavadist.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#kava.kavadist.v1beta1.QueryParamsResponse)
  
    - [Query](#kava.kavadist.v1beta1.Query)
  
- [kava/pricefeed/v1beta1/pricefeed.proto](#kava/pricefeed/v1beta1/pricefeed.proto)
    - [CurrentPrice](#kava.pricefeed.v1beta1.CurrentPrice)
    - [Market](#kava.pricefeed.v1beta1.Market)
    - [PostedPrice](#kava.pricefeed.v1beta1.PostedPrice)
  
- [kava/pricefeed/v1beta1/genesis.proto](#kava/pricefeed/v1beta1/genesis.proto)
    - [GenesisState](#kava.pricefeed.v1beta1.GenesisState)
    - [Params](#kava.pricefeed.v1beta1.Params)
  
- [kava/pricefeed/v1beta1/query.proto](#kava/pricefeed/v1beta1/query.proto)
    - [QueryMarketsRequest](#kava.pricefeed.v1beta1.QueryMarketsRequest)
    - [QueryMarketsResponse](#kava.pricefeed.v1beta1.QueryMarketsResponse)
    - [QueryOraclesRequest](#kava.pricefeed.v1beta1.QueryOraclesRequest)
    - [QueryOraclesResponse](#kava.pricefeed.v1beta1.QueryOraclesResponse)
    - [QueryParamsRequest](#kava.pricefeed.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#kava.pricefeed.v1beta1.QueryParamsResponse)
    - [QueryPriceRequest](#kava.pricefeed.v1beta1.QueryPriceRequest)
    - [QueryPriceResponse](#kava.pricefeed.v1beta1.QueryPriceResponse)
    - [QueryPricesRequest](#kava.pricefeed.v1beta1.QueryPricesRequest)
    - [QueryPricesResponse](#kava.pricefeed.v1beta1.QueryPricesResponse)
    - [QueryRawPricesRequest](#kava.pricefeed.v1beta1.QueryRawPricesRequest)
    - [QueryRawPricesResponse](#kava.pricefeed.v1beta1.QueryRawPricesResponse)
  
    - [Query](#kava.pricefeed.v1beta1.Query)
  
- [kava/pricefeed/v1beta1/tx.proto](#kava/pricefeed/v1beta1/tx.proto)
    - [MsgPostPrice](#kava.pricefeed.v1beta1.MsgPostPrice)
    - [MsgPostPriceResponse](#kava.pricefeed.v1beta1.MsgPostPriceResponse)
  
    - [Msg](#kava.pricefeed.v1beta1.Msg)
  
- [kava/swap/v1beta1/swap.proto](#kava/swap/v1beta1/swap.proto)
    - [AllowedPool](#kava.swap.v1beta1.AllowedPool)
    - [Params](#kava.swap.v1beta1.Params)
    - [PoolRecord](#kava.swap.v1beta1.PoolRecord)
    - [ShareRecord](#kava.swap.v1beta1.ShareRecord)
  
- [kava/swap/v1beta1/genesis.proto](#kava/swap/v1beta1/genesis.proto)
    - [GenesisState](#kava.swap.v1beta1.GenesisState)
  
- [kava/swap/v1beta1/query.proto](#kava/swap/v1beta1/query.proto)
    - [DepositResponse](#kava.swap.v1beta1.DepositResponse)
    - [PoolResponse](#kava.swap.v1beta1.PoolResponse)
    - [QueryDepositsRequest](#kava.swap.v1beta1.QueryDepositsRequest)
    - [QueryDepositsResponse](#kava.swap.v1beta1.QueryDepositsResponse)
    - [QueryParamsRequest](#kava.swap.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#kava.swap.v1beta1.QueryParamsResponse)
    - [QueryPoolsRequest](#kava.swap.v1beta1.QueryPoolsRequest)
    - [QueryPoolsResponse](#kava.swap.v1beta1.QueryPoolsResponse)
  
    - [Query](#kava.swap.v1beta1.Query)
  
- [kava/swap/v1beta1/tx.proto](#kava/swap/v1beta1/tx.proto)
    - [MsgDeposit](#kava.swap.v1beta1.MsgDeposit)
    - [MsgDepositResponse](#kava.swap.v1beta1.MsgDepositResponse)
    - [MsgSwapExactForTokens](#kava.swap.v1beta1.MsgSwapExactForTokens)
    - [MsgSwapExactForTokensResponse](#kava.swap.v1beta1.MsgSwapExactForTokensResponse)
    - [MsgSwapForExactTokens](#kava.swap.v1beta1.MsgSwapForExactTokens)
    - [MsgSwapForExactTokensResponse](#kava.swap.v1beta1.MsgSwapForExactTokensResponse)
    - [MsgWithdraw](#kava.swap.v1beta1.MsgWithdraw)
    - [MsgWithdrawResponse](#kava.swap.v1beta1.MsgWithdrawResponse)
  
    - [Msg](#kava.swap.v1beta1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="kava/bep3/v1beta1/bep3.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/bep3/v1beta1/bep3.proto



<a name="kava.bep3.v1beta1.AssetParam"></a>

### AssetParam
AssetParam defines parameters for each bep3 asset.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | denom represents the denominatin for this asset |
| `coin_id` | [int64](#int64) |  | coin_id represents the registered coin type to use (https://github.com/satoshilabs/slips/blob/master/slip-0044.md) |
| `supply_limit` | [SupplyLimit](#kava.bep3.v1beta1.SupplyLimit) |  | supply_limit defines the maximum supply allowed for the asset - a total or time based rate limit |
| `active` | [bool](#bool) |  | active specifies if the asset is live or paused |
| `deputy_address` | [bytes](#bytes) |  | deputy_address the kava address of the deputy |
| `fixed_fee` | [string](#string) |  | fixed_fee defines the fee for incoming swaps |
| `min_swap_amount` | [string](#string) |  | min_swap_amount defines the minimum amount able to be swapped in a single message |
| `max_swap_amount` | [string](#string) |  | max_swap_amount defines the maximum amount able to be swapped in a single message |
| `min_block_lock` | [uint64](#uint64) |  | min_block_lock defined the minimum blocks to lock |
| `max_block_lock` | [uint64](#uint64) |  | min_block_lock defined the maximum blocks to lock |






<a name="kava.bep3.v1beta1.AssetSupply"></a>

### AssetSupply
AssetSupply defines information about an asset's supply.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `incoming_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | incoming_supply represents the incoming supply of an asset |
| `outgoing_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | outgoing_supply represents the outgoing supply of an asset |
| `current_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | current_supply represents the current on-chain supply of an asset |
| `time_limited_current_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | time_limited_current_supply represents the time limited current supply of an asset |
| `time_elapsed` | [google.protobuf.Duration](#google.protobuf.Duration) |  | time_elapsed represents the time elapsed |






<a name="kava.bep3.v1beta1.AtomicSwap"></a>

### AtomicSwap
AtomicSwap defines an atomic swap between chains for the pricefeed module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | amount represents the amount being swapped |
| `random_number_hash` | [bytes](#bytes) |  | random_number_hash represents the hash of the random number |
| `expire_height` | [uint64](#uint64) |  | expire_height represents the height when the swap expires |
| `timestamp` | [int64](#int64) |  | timestamp represents the timestamp of the swap |
| `sender` | [bytes](#bytes) |  | sender is the kava chain sender of the swap |
| `recipient` | [bytes](#bytes) |  | recipient is the kava chain recipient of the swap |
| `sender_other_chain` | [string](#string) |  | sender_other_chain is the sender on the other chain |
| `recipient_other_chain` | [string](#string) |  | recipient_other_chain is the recipient on the other chain |
| `closed_block` | [int64](#int64) |  | closed_block is the block when the swap is closed |
| `status` | [SwapStatus](#kava.bep3.v1beta1.SwapStatus) |  | status represents the current status of the swap |
| `cross_chain` | [bool](#bool) |  | cross_chain identifies whether the atomic swap is cross chain |
| `direction` | [SwapDirection](#kava.bep3.v1beta1.SwapDirection) |  | direction identifies if the swap is incoming or outgoing |






<a name="kava.bep3.v1beta1.Params"></a>

### Params
Params defines the parameters for the bep3 module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_params` | [AssetParam](#kava.bep3.v1beta1.AssetParam) | repeated | asset_params define the parameters for each bep3 asset |






<a name="kava.bep3.v1beta1.SupplyLimit"></a>

### SupplyLimit
SupplyLimit define the absolute and time-based limits for an assets's supply.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limit` | [string](#string) |  | limit defines the total supply allowed |
| `time_limited` | [bool](#bool) |  | time_limited enables or disables time based supply limiting |
| `time_period` | [google.protobuf.Duration](#google.protobuf.Duration) |  | time_period specifies the duration that time_based_limit is evalulated |
| `time_based_limit` | [string](#string) |  | time_based_limit defines the maximum supply that can be swapped within time_period |





 <!-- end messages -->


<a name="kava.bep3.v1beta1.SwapDirection"></a>

### SwapDirection
SwapDirection is the direction of an AtomicSwap

| Name | Number | Description |
| ---- | ------ | ----------- |
| SWAP_DIRECTION_UNSPECIFIED | 0 | SWAP_DIRECTION_UNSPECIFIED represents unspecified or invalid swap direcation |
| SWAP_DIRECTION_INCOMING | 1 | SWAP_DIRECTION_INCOMING represents is incoming swap (to the kava chain) |
| SWAP_DIRECTION_OUTGOING | 2 | SWAP_DIRECTION_OUTGOING represents an outgoing swap (from the kava chain) |



<a name="kava.bep3.v1beta1.SwapStatus"></a>

### SwapStatus
SwapStatus is the status of an AtomicSwap

| Name | Number | Description |
| ---- | ------ | ----------- |
| SWAP_STATUS_UNSPECIFIED | 0 | SWAP_STATUS_UNSPECIFIED represents an unspecified status |
| SWAP_STATUS_OPEN | 1 | SWAP_STATUS_OPEN represents an open swap |
| SWAP_STATUS_COMPLETED | 2 | SWAP_STATUS_COMPLETED represents a completed swap |
| SWAP_STATUS_EXPIRED | 3 | SWAP_STATUS_EXPIRED represents an expired swap |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/bep3/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/bep3/v1beta1/genesis.proto



<a name="kava.bep3.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the pricefeed module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.bep3.v1beta1.Params) |  | params defines all the paramaters of the module. |
| `atomic_swaps` | [AtomicSwap](#kava.bep3.v1beta1.AtomicSwap) | repeated | atomic_swaps represents the state of stored atomic swaps |
| `supplies` | [AssetSupply](#kava.bep3.v1beta1.AssetSupply) | repeated | supplies represents the supply information of each atomic swap |
| `previous_block_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | previous_block_time represents the time of the previous block |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/bep3/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/bep3/v1beta1/query.proto



<a name="kava.bep3.v1beta1.AssetSupplyResponse"></a>

### AssetSupplyResponse
AssetSupplyResponse defines information about an asset's supply.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `incoming_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | incoming_supply represents the incoming supply of an asset |
| `outgoing_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | outgoing_supply represents the outgoing supply of an asset |
| `current_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | current_supply represents the current on-chain supply of an asset |
| `time_limited_current_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | time_limited_current_supply represents the time limited current supply of an asset |
| `time_elapsed` | [google.protobuf.Duration](#google.protobuf.Duration) |  | time_elapsed represents the time elapsed |






<a name="kava.bep3.v1beta1.AtomicSwapResponse"></a>

### AtomicSwapResponse
AtomicSwapResponse represents the returned atomic swap properties


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | id represents the id of the atomic swap |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | amount represents the amount being swapped |
| `random_number_hash` | [string](#string) |  | random_number_hash represents the hash of the random number |
| `expire_height` | [uint64](#uint64) |  | expire_height represents the height when the swap expires |
| `timestamp` | [int64](#int64) |  | timestamp represents the timestamp of the swap |
| `sender` | [string](#string) |  | sender is the kava chain sender of the swap |
| `recipient` | [string](#string) |  | recipient is the kava chain recipient of the swap |
| `sender_other_chain` | [string](#string) |  | sender_other_chain is the sender on the other chain |
| `recipient_other_chain` | [string](#string) |  | recipient_other_chain is the recipient on the other chain |
| `closed_block` | [int64](#int64) |  | closed_block is the block when the swap is closed |
| `status` | [SwapStatus](#kava.bep3.v1beta1.SwapStatus) |  | status represents the current status of the swap |
| `cross_chain` | [bool](#bool) |  | cross_chain identifies whether the atomic swap is cross chain |
| `direction` | [SwapDirection](#kava.bep3.v1beta1.SwapDirection) |  | direction identifies if the swap is incoming or outgoing |






<a name="kava.bep3.v1beta1.QueryAssetSuppliesRequest"></a>

### QueryAssetSuppliesRequest
QueryAssetSuppliesRequest is the request type for the Query/AssetSupplies RPC method.






<a name="kava.bep3.v1beta1.QueryAssetSuppliesResponse"></a>

### QueryAssetSuppliesResponse
QueryAssetSuppliesResponse is the response type for the Query/AssetSupplies RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_supplies` | [AssetSupplyResponse](#kava.bep3.v1beta1.AssetSupplyResponse) | repeated | asset_supplies represents the supplies of returned assets |






<a name="kava.bep3.v1beta1.QueryAssetSupplyRequest"></a>

### QueryAssetSupplyRequest
QueryAssetSupplyRequest is the request type for the Query/AssetSupply RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | denom filters the asset response for the specified denom |






<a name="kava.bep3.v1beta1.QueryAssetSupplyResponse"></a>

### QueryAssetSupplyResponse
QueryAssetSupplyResponse is the response type for the Query/AssetSupply RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_supply` | [AssetSupplyResponse](#kava.bep3.v1beta1.AssetSupplyResponse) |  | asset_supply represents the supply of the asset |






<a name="kava.bep3.v1beta1.QueryAtomicSwapRequest"></a>

### QueryAtomicSwapRequest
QueryAtomicSwapRequest is the request type for the Query/AtomicSwap RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `swap_id` | [string](#string) |  | swap_id represents the id of the swap to query |






<a name="kava.bep3.v1beta1.QueryAtomicSwapResponse"></a>

### QueryAtomicSwapResponse
QueryAtomicSwapResponse is the response type for the Query/AtomicSwap RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `atomic_swap` | [AtomicSwapResponse](#kava.bep3.v1beta1.AtomicSwapResponse) |  |  |






<a name="kava.bep3.v1beta1.QueryAtomicSwapsRequest"></a>

### QueryAtomicSwapsRequest
QueryAtomicSwapsRequest is the request type for the Query/AtomicSwaps RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `involve` | [string](#string) |  | involve filters by address |
| `expiration` | [uint64](#uint64) |  | expiration filters by expiration block height |
| `status` | [SwapStatus](#kava.bep3.v1beta1.SwapStatus) |  | status filters by swap status |
| `direction` | [SwapDirection](#kava.bep3.v1beta1.SwapDirection) |  | direction fitlers by swap direction |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="kava.bep3.v1beta1.QueryAtomicSwapsResponse"></a>

### QueryAtomicSwapsResponse
QueryAtomicSwapsResponse is the response type for the Query/AtomicSwaps RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `atomic_swaps` | [AtomicSwapResponse](#kava.bep3.v1beta1.AtomicSwapResponse) | repeated | atomic_swap represents the returned atomic swaps for the request |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |






<a name="kava.bep3.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/bep3 parameters.






<a name="kava.bep3.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/bep3 parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.bep3.v1beta1.Params) |  | params represents the parameters of the module |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.bep3.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service for bep3 module

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#kava.bep3.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#kava.bep3.v1beta1.QueryParamsResponse) | Params queries module params | GET|/kava/bep3/v1beta1/params|
| `AssetSupply` | [QueryAssetSupplyRequest](#kava.bep3.v1beta1.QueryAssetSupplyRequest) | [QueryAssetSupplyResponse](#kava.bep3.v1beta1.QueryAssetSupplyResponse) | AssetSupply queries info about an asset's supply | GET|/kava/bep3/v1beta1/assetsupply/{denom}|
| `AssetSupplies` | [QueryAssetSuppliesRequest](#kava.bep3.v1beta1.QueryAssetSuppliesRequest) | [QueryAssetSuppliesResponse](#kava.bep3.v1beta1.QueryAssetSuppliesResponse) | AssetSupplies queries a list of asset supplies | GET|/kava/bep3/v1beta1/assetsupplies|
| `AtomicSwap` | [QueryAtomicSwapRequest](#kava.bep3.v1beta1.QueryAtomicSwapRequest) | [QueryAtomicSwapResponse](#kava.bep3.v1beta1.QueryAtomicSwapResponse) | AtomicSwap queries info about an atomic swap | GET|/kava/bep3/v1beta1/atomicswap/{swap_id}|
| `AtomicSwaps` | [QueryAtomicSwapsRequest](#kava.bep3.v1beta1.QueryAtomicSwapsRequest) | [QueryAtomicSwapsResponse](#kava.bep3.v1beta1.QueryAtomicSwapsResponse) | AtomicSwaps queries a list of atomic swaps | GET|/kava/bep3/v1beta1/atomicswaps|

 <!-- end services -->



<a name="kava/bep3/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/bep3/v1beta1/tx.proto



<a name="kava.bep3.v1beta1.MsgClaimAtomicSwap"></a>

### MsgClaimAtomicSwap
MsgClaimAtomicSwap defines the Msg/ClaimAtomicSwap request type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  |  |
| `swap_id` | [bytes](#bytes) |  |  |
| `random_number` | [bytes](#bytes) |  |  |






<a name="kava.bep3.v1beta1.MsgClaimAtomicSwapResponse"></a>

### MsgClaimAtomicSwapResponse
MsgClaimAtomicSwapResponse defines the Msg/ClaimAtomicSwap response type.






<a name="kava.bep3.v1beta1.MsgCreateAtomicSwap"></a>

### MsgCreateAtomicSwap
MsgCreateAtomicSwap defines the Msg/CreateAtomicSwap request type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  |  |
| `to` | [string](#string) |  |  |
| `recipient_other_chain` | [string](#string) |  |  |
| `sender_other_chain` | [string](#string) |  |  |
| `random_number_hash` | [string](#string) |  |  |
| `timestamp` | [int64](#int64) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `height_span` | [uint64](#uint64) |  |  |






<a name="kava.bep3.v1beta1.MsgCreateAtomicSwapResponse"></a>

### MsgCreateAtomicSwapResponse
MsgCreateAtomicSwapResponse defines the Msg/CreateAtomicSwap response type.






<a name="kava.bep3.v1beta1.MsgRefundAtomicSwap"></a>

### MsgRefundAtomicSwap
MsgRefundAtomicSwap defines the Msg/RefundAtomicSwap request type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  |  |
| `swap_id` | [bytes](#bytes) |  |  |






<a name="kava.bep3.v1beta1.MsgRefundAtomicSwapResponse"></a>

### MsgRefundAtomicSwapResponse
MsgRefundAtomicSwapResponse defines the Msg/RefundAtomicSwap response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.bep3.v1beta1.Msg"></a>

### Msg
Msg defines the bep3 Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateAtomicSwap` | [MsgCreateAtomicSwap](#kava.bep3.v1beta1.MsgCreateAtomicSwap) | [MsgCreateAtomicSwapResponse](#kava.bep3.v1beta1.MsgCreateAtomicSwapResponse) | CreateAtomicSwap defines a method for creating an atomic swap | |
| `ClaimAtomicSwap` | [MsgClaimAtomicSwap](#kava.bep3.v1beta1.MsgClaimAtomicSwap) | [MsgClaimAtomicSwapResponse](#kava.bep3.v1beta1.MsgClaimAtomicSwapResponse) | ClaimAtomicSwap defines a method for claiming an atomic swap | |
| `RefundAtomicSwap` | [MsgRefundAtomicSwap](#kava.bep3.v1beta1.MsgRefundAtomicSwap) | [MsgRefundAtomicSwapResponse](#kava.bep3.v1beta1.MsgRefundAtomicSwapResponse) | RefundAtomicSwap defines a method for refunding an atomic swap | |

 <!-- end services -->



<a name="kava/committee/v1beta1/committee.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/committee/v1beta1/committee.proto



<a name="kava.committee.v1beta1.BaseCommittee"></a>

### BaseCommittee
BaseCommittee is a common type shared by all Committees


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `description` | [string](#string) |  |  |
| `members` | [bytes](#bytes) | repeated |  |
| `permissions` | [google.protobuf.Any](#google.protobuf.Any) | repeated |  |
| `vote_threshold` | [string](#string) |  | Smallest percentage that must vote for a proposal to pass |
| `proposal_duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  | The length of time a proposal remains active for. Proposals will close earlier if they get enough votes. |
| `tally_option` | [TallyOption](#kava.committee.v1beta1.TallyOption) |  |  |






<a name="kava.committee.v1beta1.MemberCommittee"></a>

### MemberCommittee
MemberCommittee is an alias of BaseCommittee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_committee` | [BaseCommittee](#kava.committee.v1beta1.BaseCommittee) |  |  |






<a name="kava.committee.v1beta1.TokenCommittee"></a>

### TokenCommittee
TokenCommittee supports voting on proposals by token holders


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_committee` | [BaseCommittee](#kava.committee.v1beta1.BaseCommittee) |  |  |
| `quorum` | [string](#string) |  |  |
| `tally_denom` | [string](#string) |  |  |





 <!-- end messages -->


<a name="kava.committee.v1beta1.TallyOption"></a>

### TallyOption
TallyOption enumerates the valid types of a tally.

| Name | Number | Description |
| ---- | ------ | ----------- |
| TALLY_OPTION_UNSPECIFIED | 0 | TALLY_OPTION_UNSPECIFIED defines a null tally option. |
| TALLY_OPTION_FIRST_PAST_THE_POST | 1 | Votes are tallied each block and the proposal passes as soon as the vote threshold is reached |
| TALLY_OPTION_DEADLINE | 2 | Votes are tallied exactly once, when the deadline time is reached |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/committee/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/committee/v1beta1/genesis.proto



<a name="kava.committee.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the committee module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `next_proposal_id` | [uint64](#uint64) |  |  |
| `committees` | [google.protobuf.Any](#google.protobuf.Any) | repeated |  |
| `proposals` | [Proposal](#kava.committee.v1beta1.Proposal) | repeated |  |
| `votes` | [Vote](#kava.committee.v1beta1.Vote) | repeated |  |






<a name="kava.committee.v1beta1.Proposal"></a>

### Proposal
Proposal is an internal record of a governance proposal submitted to a committee.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `content` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `id` | [uint64](#uint64) |  |  |
| `committee_id` | [uint64](#uint64) |  |  |
| `deadline` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="kava.committee.v1beta1.Vote"></a>

### Vote
Vote is an internal record of a single governance vote.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `proposal_id` | [uint64](#uint64) |  |  |
| `voter` | [bytes](#bytes) |  |  |
| `vote_type` | [VoteType](#kava.committee.v1beta1.VoteType) |  |  |





 <!-- end messages -->


<a name="kava.committee.v1beta1.VoteType"></a>

### VoteType
VoteType enumerates the valid types of a vote.

| Name | Number | Description |
| ---- | ------ | ----------- |
| VOTE_TYPE_UNSPECIFIED | 0 | VOTE_TYPE_UNSPECIFIED defines a no-op vote option. |
| VOTE_TYPE_YES | 1 | VOTE_TYPE_YES defines a yes vote option. |
| VOTE_TYPE_NO | 2 | VOTE_TYPE_NO defines a no vote option. |
| VOTE_TYPE_ABSTAIN | 3 | VOTE_TYPE_ABSTAIN defines an abstain vote option. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/committee/v1beta1/permissions.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/committee/v1beta1/permissions.proto



<a name="kava.committee.v1beta1.GodPermission"></a>

### GodPermission
GodPermission allows any governance proposal. It is used mainly for testing.






<a name="kava.committee.v1beta1.SoftwareUpgradePermission"></a>

### SoftwareUpgradePermission
SoftwareUpgradePermission permission type for software upgrade proposals






<a name="kava.committee.v1beta1.TextPermission"></a>

### TextPermission
TextPermission allows any text governance proposal.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/committee/v1beta1/proposal.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/committee/v1beta1/proposal.proto



<a name="kava.committee.v1beta1.CommitteeChangeProposal"></a>

### CommitteeChangeProposal
CommitteeChangeProposal is a gov proposal for creating a new committee or modifying an existing one.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `new_committee` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="kava.committee.v1beta1.CommitteeDeleteProposal"></a>

### CommitteeDeleteProposal
CommitteeDeleteProposal is a gov proposal for removing a committee.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `committee_id` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/committee/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/committee/v1beta1/tx.proto



<a name="kava.committee.v1beta1.MsgSubmitProposal"></a>

### MsgSubmitProposal
MsgSubmitProposal is used by committee members to create a new proposal that they can vote on.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pub_proposal` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `proposer` | [bytes](#bytes) |  |  |
| `committee_id` | [uint64](#uint64) |  |  |






<a name="kava.committee.v1beta1.MsgVote"></a>

### MsgVote
MsgVote is submitted by committee members to vote on proposals.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `proposal_id` | [uint64](#uint64) |  |  |
| `voter` | [bytes](#bytes) |  |  |
| `vote_type` | [VoteType](#kava.committee.v1beta1.VoteType) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/issuance/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/issuance/v1beta1/genesis.proto



<a name="kava.issuance.v1beta1.Asset"></a>

### Asset
Asset type for assets in the issuance module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `blocked_addresses` | [string](#string) | repeated |  |
| `paused` | [bool](#bool) |  |  |
| `blockable` | [bool](#bool) |  |  |
| `rate_limit` | [RateLimit](#kava.issuance.v1beta1.RateLimit) |  |  |






<a name="kava.issuance.v1beta1.AssetSupply"></a>

### AssetSupply
AssetSupply contains information about an asset's rate-limited supply (the
total supply of the asset is tracked in the top-level supply module)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `current_supply` | [bytes](#bytes) |  |  |
| `time_elapsed` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="kava.issuance.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the issuance module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.issuance.v1beta1.Params) |  | params defines all the paramaters of the module. |
| `supplies` | [AssetSupply](#kava.issuance.v1beta1.AssetSupply) | repeated |  |






<a name="kava.issuance.v1beta1.Params"></a>

### Params
Params defines the parameters for the issuance module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assets` | [Asset](#kava.issuance.v1beta1.Asset) | repeated |  |






<a name="kava.issuance.v1beta1.RateLimit"></a>

### RateLimit
RateLimit parameters for rate-limiting the supply of an issued asset


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `active` | [bool](#bool) |  |  |
| `limit` | [bytes](#bytes) |  |  |
| `time_period` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/issuance/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/issuance/v1beta1/query.proto



<a name="kava.issuance.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/issuance parameters.






<a name="kava.issuance.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/issuance parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.issuance.v1beta1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.issuance.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service for issuance module

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#kava.issuance.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#kava.issuance.v1beta1.QueryParamsResponse) | Params queries all parameters of the issuance module. | GET|/kava/issuance/v1beta1/params|

 <!-- end services -->



<a name="kava/issuance/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/issuance/v1beta1/tx.proto



<a name="kava.issuance.v1beta1.MsgBlockAddress"></a>

### MsgBlockAddress
MsgBlockAddress represents a message used by the issuer to block an address from holding or transferring tokens


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `blocked_address` | [string](#string) |  |  |






<a name="kava.issuance.v1beta1.MsgBlockAddressResponse"></a>

### MsgBlockAddressResponse
MsgBlockAddressResponse defines the Msg/BlockAddress response type.






<a name="kava.issuance.v1beta1.MsgIssueTokens"></a>

### MsgIssueTokens
MsgIssueTokens represents a message used by the issuer to issue new tokens


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `tokens` | [bytes](#bytes) |  |  |
| `receiver` | [string](#string) |  |  |






<a name="kava.issuance.v1beta1.MsgIssueTokensResponse"></a>

### MsgIssueTokensResponse
MsgIssueTokensResponse defines the Msg/IssueTokens response type.






<a name="kava.issuance.v1beta1.MsgRedeemTokens"></a>

### MsgRedeemTokens
MsgRedeemTokens represents a message used by the issuer to redeem (burn) tokens


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `tokens` | [bytes](#bytes) |  |  |






<a name="kava.issuance.v1beta1.MsgRedeemTokensResponse"></a>

### MsgRedeemTokensResponse
MsgRedeemTokensResponse defines the Msg/RedeemTokens response type.






<a name="kava.issuance.v1beta1.MsgSetPauseStatus"></a>

### MsgSetPauseStatus
MsgSetPauseStatus message type used by the issuer to pause or unpause status


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `status` | [bool](#bool) |  |  |






<a name="kava.issuance.v1beta1.MsgSetPauseStatusResponse"></a>

### MsgSetPauseStatusResponse
MsgSetPauseStatusResponse defines the Msg/SetPauseStatus response type.






<a name="kava.issuance.v1beta1.MsgUnblockAddress"></a>

### MsgUnblockAddress
MsgUnblockAddress message type used by the issuer to unblock an address from holding or transferring tokens


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `blocked_address` | [string](#string) |  |  |






<a name="kava.issuance.v1beta1.MsgUnblockAddressResponse"></a>

### MsgUnblockAddressResponse
MsgUnblockAddressResponse defines the Msg/UnblockAddress response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.issuance.v1beta1.Msg"></a>

### Msg
Msg defines the issuance Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `IssueTokens` | [MsgIssueTokens](#kava.issuance.v1beta1.MsgIssueTokens) | [MsgIssueTokensResponse](#kava.issuance.v1beta1.MsgIssueTokensResponse) | IssueTokens message type used by the issuer to issue new tokens | |
| `RedeemTokens` | [MsgRedeemTokens](#kava.issuance.v1beta1.MsgRedeemTokens) | [MsgRedeemTokensResponse](#kava.issuance.v1beta1.MsgRedeemTokensResponse) | RedeemTokens message type used by the issuer to redeem (burn) tokens | |
| `BlockAddress` | [MsgBlockAddress](#kava.issuance.v1beta1.MsgBlockAddress) | [MsgBlockAddressResponse](#kava.issuance.v1beta1.MsgBlockAddressResponse) | BlockAddress message type used by the issuer to block an address from holding or transferring tokens | |
| `UnblockAddress` | [MsgUnblockAddress](#kava.issuance.v1beta1.MsgUnblockAddress) | [MsgUnblockAddressResponse](#kava.issuance.v1beta1.MsgUnblockAddressResponse) | UnblockAddress message type used by the issuer to unblock an address from holding or transferring tokens | |
| `SetPauseStatus` | [MsgSetPauseStatus](#kava.issuance.v1beta1.MsgSetPauseStatus) | [MsgSetPauseStatusResponse](#kava.issuance.v1beta1.MsgSetPauseStatusResponse) | SetPauseStatus message type used to pause or unpause status | |

 <!-- end services -->



<a name="kava/kavadist/v1beta1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/kavadist/v1beta1/params.proto



<a name="kava.kavadist.v1beta1.Params"></a>

### Params
Params governance parameters for kavadist module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `active` | [bool](#bool) |  |  |
| `periods` | [Period](#kava.kavadist.v1beta1.Period) | repeated |  |






<a name="kava.kavadist.v1beta1.Period"></a>

### Period
Period stores the specified start and end dates, and the inflation, expressed as a decimal
representing the yearly APR of KAVA tokens that will be minted during that period


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `start` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | example "2020-03-01T15:20:00Z" |
| `end` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | example "2020-06-01T15:20:00Z" |
| `inflation` | [bytes](#bytes) |  | example "1.000000003022265980" - 10% inflation |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/kavadist/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/kavadist/v1beta1/genesis.proto



<a name="kava.kavadist.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the kavadist module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.kavadist.v1beta1.Params) |  |  |
| `previous_block_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/kavadist/v1beta1/proposal.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/kavadist/v1beta1/proposal.proto



<a name="kava.kavadist.v1beta1.CommunityPoolMultiSpendProposal"></a>

### CommunityPoolMultiSpendProposal
CommunityPoolMultiSpendProposal spends from the community pool by sending to one or more
addresses


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `recipient_list` | [MultiSpendRecipient](#kava.kavadist.v1beta1.MultiSpendRecipient) | repeated |  |






<a name="kava.kavadist.v1beta1.CommunityPoolMultiSpendProposalJSON"></a>

### CommunityPoolMultiSpendProposalJSON
CommunityPoolMultiSpendProposalJSON defines a CommunityPoolMultiSpendProposal with a deposit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `recipient_list` | [MultiSpendRecipient](#kava.kavadist.v1beta1.MultiSpendRecipient) | repeated |  |
| `deposit` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="kava.kavadist.v1beta1.MultiSpendRecipient"></a>

### MultiSpendRecipient
MultiSpendRecipient defines a recipient and the amount of coins they are receiving


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/kavadist/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/kavadist/v1beta1/query.proto



<a name="kava.kavadist.v1beta1.QueryBalanceRequest"></a>

### QueryBalanceRequest
QueryBalanceRequest defines the request type for querying x/kavadist balance.






<a name="kava.kavadist.v1beta1.QueryBalanceResponse"></a>

### QueryBalanceResponse
QueryBalanceResponse defines the response type for querying x/kavadist balance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="kava.kavadist.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/kavadist parameters.






<a name="kava.kavadist.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/kavadist parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.kavadist.v1beta1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.kavadist.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#kava.kavadist.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#kava.kavadist.v1beta1.QueryParamsResponse) | Params queries the parameters of x/kavadist module. | GET|/kava/kavadist/v1beta1/parameters|
| `Balance` | [QueryBalanceRequest](#kava.kavadist.v1beta1.QueryBalanceRequest) | [QueryBalanceResponse](#kava.kavadist.v1beta1.QueryBalanceResponse) | Balance queries the balance of all coins of x/kavadist module. | GET|/kava/kavadist/v1beta1/balance|

 <!-- end services -->



<a name="kava/pricefeed/v1beta1/pricefeed.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/pricefeed/v1beta1/pricefeed.proto



<a name="kava.pricefeed.v1beta1.CurrentPrice"></a>

### CurrentPrice
CurrentPrice defines a current price for a particular market in the pricefeed
module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |
| `price` | [string](#string) |  |  |






<a name="kava.pricefeed.v1beta1.Market"></a>

### Market
Market defines an asset in the pricefeed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |
| `base_asset` | [string](#string) |  |  |
| `quote_asset` | [string](#string) |  |  |
| `oracles` | [string](#string) | repeated |  |
| `active` | [bool](#bool) |  |  |






<a name="kava.pricefeed.v1beta1.PostedPrice"></a>

### PostedPrice
PostedPrice defines a price for market posted by a specific oracle.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |
| `oracle_address` | [string](#string) |  |  |
| `price` | [string](#string) |  |  |
| `expiry` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/pricefeed/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/pricefeed/v1beta1/genesis.proto



<a name="kava.pricefeed.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the pricefeed module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.pricefeed.v1beta1.Params) |  | params defines all the paramaters of the module. |
| `posted_prices` | [PostedPrice](#kava.pricefeed.v1beta1.PostedPrice) | repeated |  |






<a name="kava.pricefeed.v1beta1.Params"></a>

### Params
Params defines the parameters for the pricefeed module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `markets` | [Market](#kava.pricefeed.v1beta1.Market) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/pricefeed/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/pricefeed/v1beta1/query.proto



<a name="kava.pricefeed.v1beta1.QueryMarketsRequest"></a>

### QueryMarketsRequest
QueryMarketsRequest is the request type for the Query/Markets RPC method.






<a name="kava.pricefeed.v1beta1.QueryMarketsResponse"></a>

### QueryMarketsResponse
QueryMarketsResponse is the response type for the Query/Markets RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `markets` | [Market](#kava.pricefeed.v1beta1.Market) | repeated | List of markets |






<a name="kava.pricefeed.v1beta1.QueryOraclesRequest"></a>

### QueryOraclesRequest
QueryOraclesRequest is the request type for the Query/Oracles RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |






<a name="kava.pricefeed.v1beta1.QueryOraclesResponse"></a>

### QueryOraclesResponse
QueryOraclesResponse is the response type for the Query/Oracles RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `oracles` | [string](#string) | repeated | List of oracle addresses |






<a name="kava.pricefeed.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/pricefeed
parameters.






<a name="kava.pricefeed.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/pricefeed
parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.pricefeed.v1beta1.Params) |  |  |






<a name="kava.pricefeed.v1beta1.QueryPriceRequest"></a>

### QueryPriceRequest
QueryPriceRequest is the request type for the Query/PriceRequest RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |






<a name="kava.pricefeed.v1beta1.QueryPriceResponse"></a>

### QueryPriceResponse
QueryPriceResponse is the response type for the Query/Prices RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `price` | [CurrentPrice](#kava.pricefeed.v1beta1.CurrentPrice) |  |  |






<a name="kava.pricefeed.v1beta1.QueryPricesRequest"></a>

### QueryPricesRequest
QueryPricesRequest is the request type for the Query/Prices RPC method.






<a name="kava.pricefeed.v1beta1.QueryPricesResponse"></a>

### QueryPricesResponse
QueryPricesResponse is the response type for the Query/Prices RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `prices` | [CurrentPrice](#kava.pricefeed.v1beta1.CurrentPrice) | repeated |  |






<a name="kava.pricefeed.v1beta1.QueryRawPricesRequest"></a>

### QueryRawPricesRequest
QueryRawPricesRequest is the request type for the Query/RawPrices RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |






<a name="kava.pricefeed.v1beta1.QueryRawPricesResponse"></a>

### QueryRawPricesResponse
QueryRawPricesResponse is the response type for the Query/RawPrices RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `raw_prices` | [PostedPrice](#kava.pricefeed.v1beta1.PostedPrice) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.pricefeed.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service for pricefeed module

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#kava.pricefeed.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#kava.pricefeed.v1beta1.QueryParamsResponse) | Params queries all parameters of the pricefeed module. | GET|/kava/pricefeed/v1beta1/params|
| `Price` | [QueryPriceRequest](#kava.pricefeed.v1beta1.QueryPriceRequest) | [QueryPriceResponse](#kava.pricefeed.v1beta1.QueryPriceResponse) | Price queries price details based on a market | GET|/kava/pricefeed/v1beta1/prices/{market_id}|
| `Prices` | [QueryPricesRequest](#kava.pricefeed.v1beta1.QueryPricesRequest) | [QueryPricesResponse](#kava.pricefeed.v1beta1.QueryPricesResponse) | Prices queries all prices | GET|/kava/pricefeed/v1beta1/prices|
| `RawPrices` | [QueryRawPricesRequest](#kava.pricefeed.v1beta1.QueryRawPricesRequest) | [QueryRawPricesResponse](#kava.pricefeed.v1beta1.QueryRawPricesResponse) | RawPrices queries all raw prices based on a market | GET|/kava/pricefeed/v1beta1/rawprices/{market_id}|
| `Oracles` | [QueryOraclesRequest](#kava.pricefeed.v1beta1.QueryOraclesRequest) | [QueryOraclesResponse](#kava.pricefeed.v1beta1.QueryOraclesResponse) | Oracles queries all oracles based on a market | GET|/kava/pricefeed/v1beta1/oracles/{market_id}|
| `Markets` | [QueryMarketsRequest](#kava.pricefeed.v1beta1.QueryMarketsRequest) | [QueryMarketsResponse](#kava.pricefeed.v1beta1.QueryMarketsResponse) | Markets queries all markets | GET|/kava/pricefeed/v1beta1/markets|

 <!-- end services -->



<a name="kava/pricefeed/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/pricefeed/v1beta1/tx.proto



<a name="kava.pricefeed.v1beta1.MsgPostPrice"></a>

### MsgPostPrice
MsgPostPrice represents a method for creating a new post price


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  | address of client |
| `market_id` | [string](#string) |  |  |
| `price` | [string](#string) |  |  |
| `expiry` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="kava.pricefeed.v1beta1.MsgPostPriceResponse"></a>

### MsgPostPriceResponse
MsgPostPriceResponse defines the Msg/PostPrice response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.pricefeed.v1beta1.Msg"></a>

### Msg
Msg defines the pricefeed Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `PostPrice` | [MsgPostPrice](#kava.pricefeed.v1beta1.MsgPostPrice) | [MsgPostPriceResponse](#kava.pricefeed.v1beta1.MsgPostPriceResponse) | PostPrice defines a method for creating a new post price | |

 <!-- end services -->



<a name="kava/swap/v1beta1/swap.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/swap/v1beta1/swap.proto



<a name="kava.swap.v1beta1.AllowedPool"></a>

### AllowedPool
AllowedPool defines a pool that is allowed to be created


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_a` | [string](#string) |  | token_a represents the a token allowed |
| `token_b` | [string](#string) |  | token_b represents the b token allowed |






<a name="kava.swap.v1beta1.Params"></a>

### Params
Params defines the parameters for the swap module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `allowed_pools` | [AllowedPool](#kava.swap.v1beta1.AllowedPool) | repeated | allowed_pools defines that pools that are allowed to be created |
| `swap_fee` | [string](#string) |  | swap_fee defines the swap fee for all pools |






<a name="kava.swap.v1beta1.PoolRecord"></a>

### PoolRecord
PoolRecord represents the state of a liquidity pool
and is used to store the state of a denominated pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pool_id` | [string](#string) |  | pool_id represents the unique id of the pool |
| `reserves_a` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | reserves_a is the a token coin reserves |
| `reserves_b` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | reserves_b is the a token coin reserves |
| `total_shares` | [string](#string) |  | total_shares is the total distrubuted shares of the pool |






<a name="kava.swap.v1beta1.ShareRecord"></a>

### ShareRecord
ShareRecord stores the shares owned for a depositor and pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `depositor` | [bytes](#bytes) |  | depositor represents the owner of the shares |
| `pool_id` | [string](#string) |  | pool_id represents the pool the shares belong to |
| `shares_owned` | [string](#string) |  | shares_owned represents the number of shares owned by depsoitor for the pool_id |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/swap/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/swap/v1beta1/genesis.proto



<a name="kava.swap.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the swap module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.swap.v1beta1.Params) |  | params defines all the paramaters related to swap |
| `pool_records` | [PoolRecord](#kava.swap.v1beta1.PoolRecord) | repeated | pool_records defines the available pools |
| `share_records` | [ShareRecord](#kava.swap.v1beta1.ShareRecord) | repeated | share_records defines the owned shares of each pool |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/swap/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/swap/v1beta1/query.proto



<a name="kava.swap.v1beta1.DepositResponse"></a>

### DepositResponse
DepositResponse defines a single deposit query response type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `depositor` | [string](#string) |  | depositor represents the owner of the deposit |
| `pool_id` | [string](#string) |  | pool_id represents the pool the deposit is for |
| `shares_owned` | [string](#string) |  | shares_owned presents the shares owned by the depositor for the pool |
| `shares_value` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | shares_value represents the coin value of the shares_owned |






<a name="kava.swap.v1beta1.PoolResponse"></a>

### PoolResponse
Pool represents the state of a single pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name represents the name of the pool |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | coins represents the total reserves of the pool |
| `total_shares` | [string](#string) |  | total_shares represents the total shares of the pool |






<a name="kava.swap.v1beta1.QueryDepositsRequest"></a>

### QueryDepositsRequest
QueryDepositsRequest is the request type for the Query/Deposits RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner optionally filters deposits by owner |
| `pool_id` | [string](#string) |  | pool_id optionally fitlers deposits by pool id |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="kava.swap.v1beta1.QueryDepositsResponse"></a>

### QueryDepositsResponse
QueryDepositsResponse is the response type for the Query/Deposits RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposits` | [DepositResponse](#kava.swap.v1beta1.DepositResponse) | repeated | deposits returns the deposits matching the requested parameters |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="kava.swap.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/swap parameters.






<a name="kava.swap.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/swap parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.swap.v1beta1.Params) |  | params represents the swap module parameters |






<a name="kava.swap.v1beta1.QueryPoolsRequest"></a>

### QueryPoolsRequest
QueryPoolsRequest is the request type for the Query/Pools RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pool_id` | [string](#string) |  | pool_id filters pools by id |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="kava.swap.v1beta1.QueryPoolsResponse"></a>

### QueryPoolsResponse
QueryPoolsResponse is the response type for the Query/Pools RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pools` | [PoolResponse](#kava.swap.v1beta1.PoolResponse) | repeated | pools represents returned pools |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.swap.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service for swap module

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#kava.swap.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#kava.swap.v1beta1.QueryParamsResponse) | Params queries all parameters of the swap module. | GET|/kava/swap/v1beta1/params|
| `Pools` | [QueryPoolsRequest](#kava.swap.v1beta1.QueryPoolsRequest) | [QueryPoolsResponse](#kava.swap.v1beta1.QueryPoolsResponse) | Pools queries pools based on pool ID | GET|/kava/swap/v1beta1/pools|
| `Deposits` | [QueryDepositsRequest](#kava.swap.v1beta1.QueryDepositsRequest) | [QueryDepositsResponse](#kava.swap.v1beta1.QueryDepositsResponse) | Deposits queries deposit details based on owner address and pool | GET|/kava/swap/v1beta1/deposits|

 <!-- end services -->



<a name="kava/swap/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/swap/v1beta1/tx.proto



<a name="kava.swap.v1beta1.MsgDeposit"></a>

### MsgDeposit
MsgDeposit represents a message for depositing liquidity into a pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `depositor` | [string](#string) |  | depositor represents the address to deposit funds from |
| `token_a` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | token_a represents one token of deposit pair |
| `token_b` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | token_b represents one token of deposit pair |
| `slippage` | [string](#string) |  | slippage represents the max decimal percentage price change |
| `deadline` | [int64](#int64) |  | deadline represents the unix timestamp to complete the deposit by |






<a name="kava.swap.v1beta1.MsgDepositResponse"></a>

### MsgDepositResponse
MsgDepositResponse defines the Msg/Deposit response type.






<a name="kava.swap.v1beta1.MsgSwapExactForTokens"></a>

### MsgSwapExactForTokens
MsgSwapExactForTokens represents a message for trading exact coinA for coinB


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `requester` | [string](#string) |  | represents the address swaping the tokens |
| `exact_token_a` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | exact_token_a represents the exact amount to swap for token_b |
| `token_b` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | token_b represents the desired token_b to swap for |
| `slippage` | [string](#string) |  | slippage represents the maximum change in token_b allowed |
| `deadline` | [int64](#int64) |  | deadline represents the unix timestamp to complete the swap by |






<a name="kava.swap.v1beta1.MsgSwapExactForTokensResponse"></a>

### MsgSwapExactForTokensResponse
MsgSwapExactForTokensResponse defines the Msg/SwapExactForTokens response
type.






<a name="kava.swap.v1beta1.MsgSwapForExactTokens"></a>

### MsgSwapForExactTokens
MsgSwapForExactTokens represents a message for trading coinA for an exact
coinB


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `requester` | [string](#string) |  | represents the address swaping the tokens |
| `token_a` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | token_a represents the desired token_a to swap for |
| `exact_token_b` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | exact_token_b represents the exact token b amount to swap for token a |
| `slippage` | [string](#string) |  | slippage represents the maximum change in token_a allowed |
| `deadline` | [int64](#int64) |  | deadline represents the unix timestamp to complete the swap by |






<a name="kava.swap.v1beta1.MsgSwapForExactTokensResponse"></a>

### MsgSwapForExactTokensResponse
MsgSwapForExactTokensResponse defines the Msg/SwapForExactTokensResponse
response type.






<a name="kava.swap.v1beta1.MsgWithdraw"></a>

### MsgWithdraw
MsgWithdraw represents a message for withdrawing liquidity from a pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  | from represents the address we are withdrawing for |
| `shares` | [string](#string) |  | shares represents the amount of shares to withdraw |
| `min_token_a` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | min_token_a represents the minimum a token to withdraw |
| `min_token_b` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | min_token_a represents the minimum a token to withdraw |
| `deadline` | [int64](#int64) |  | deadline represents the unix timestamp to complete the withdraw by |






<a name="kava.swap.v1beta1.MsgWithdrawResponse"></a>

### MsgWithdrawResponse
MsgWithdrawResponse defines the Msg/Withdraw response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.swap.v1beta1.Msg"></a>

### Msg
Msg defines the swap Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Deposit` | [MsgDeposit](#kava.swap.v1beta1.MsgDeposit) | [MsgDepositResponse](#kava.swap.v1beta1.MsgDepositResponse) | Deposit defines a method for depositing liquidity into a pool | |
| `Withdraw` | [MsgWithdraw](#kava.swap.v1beta1.MsgWithdraw) | [MsgWithdrawResponse](#kava.swap.v1beta1.MsgWithdrawResponse) | Withdraw defines a method for withdrawing liquidity into a pool | |
| `SwapExactForTokens` | [MsgSwapExactForTokens](#kava.swap.v1beta1.MsgSwapExactForTokens) | [MsgSwapExactForTokensResponse](#kava.swap.v1beta1.MsgSwapExactForTokensResponse) | SwapExactForTokens represents a message for trading exact coinA for coinB | |
| `SwapForExactTokens` | [MsgSwapForExactTokens](#kava.swap.v1beta1.MsgSwapForExactTokens) | [MsgSwapForExactTokensResponse](#kava.swap.v1beta1.MsgSwapForExactTokensResponse) | SwapForExactTokens represents a message for trading coinA for an exact coinB | |

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

