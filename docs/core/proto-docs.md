 <!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

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
  
- [kava/pricefeed/v1beta1/store.proto](#kava/pricefeed/v1beta1/store.proto)
    - [CurrentPrice](#kava.pricefeed.v1beta1.CurrentPrice)
    - [Market](#kava.pricefeed.v1beta1.Market)
    - [Params](#kava.pricefeed.v1beta1.Params)
    - [PostedPrice](#kava.pricefeed.v1beta1.PostedPrice)
  
- [kava/pricefeed/v1beta1/genesis.proto](#kava/pricefeed/v1beta1/genesis.proto)
    - [GenesisState](#kava.pricefeed.v1beta1.GenesisState)
  
- [kava/pricefeed/v1beta1/query.proto](#kava/pricefeed/v1beta1/query.proto)
    - [CurrentPriceResponse](#kava.pricefeed.v1beta1.CurrentPriceResponse)
    - [MarketResponse](#kava.pricefeed.v1beta1.MarketResponse)
    - [PostedPriceResponse](#kava.pricefeed.v1beta1.PostedPriceResponse)
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



<a name="kava/pricefeed/v1beta1/store.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/pricefeed/v1beta1/store.proto



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
| `oracles` | [bytes](#bytes) | repeated |  |
| `active` | [bool](#bool) |  |  |






<a name="kava.pricefeed.v1beta1.Params"></a>

### Params
Params defines the parameters for the pricefeed module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `markets` | [Market](#kava.pricefeed.v1beta1.Market) | repeated |  |






<a name="kava.pricefeed.v1beta1.PostedPrice"></a>

### PostedPrice
PostedPrice defines a price for market posted by a specific oracle.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |
| `oracle_address` | [bytes](#bytes) |  |  |
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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/pricefeed/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/pricefeed/v1beta1/query.proto



<a name="kava.pricefeed.v1beta1.CurrentPriceResponse"></a>

### CurrentPriceResponse
CurrentPriceResponse defines a current price for a particular market in the pricefeed
module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |
| `price` | [string](#string) |  |  |






<a name="kava.pricefeed.v1beta1.MarketResponse"></a>

### MarketResponse
MarketResponse defines an asset in the pricefeed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |
| `base_asset` | [string](#string) |  |  |
| `quote_asset` | [string](#string) |  |  |
| `oracles` | [string](#string) | repeated |  |
| `active` | [bool](#bool) |  |  |






<a name="kava.pricefeed.v1beta1.PostedPriceResponse"></a>

### PostedPriceResponse
PostedPriceResponse defines a price for market posted by a specific oracle.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [string](#string) |  |  |
| `oracle_address` | [string](#string) |  |  |
| `price` | [string](#string) |  |  |
| `expiry` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="kava.pricefeed.v1beta1.QueryMarketsRequest"></a>

### QueryMarketsRequest
QueryMarketsRequest is the request type for the Query/Markets RPC method.






<a name="kava.pricefeed.v1beta1.QueryMarketsResponse"></a>

### QueryMarketsResponse
QueryMarketsResponse is the response type for the Query/Markets RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `markets` | [MarketResponse](#kava.pricefeed.v1beta1.MarketResponse) | repeated | List of markets |






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
| `price` | [CurrentPriceResponse](#kava.pricefeed.v1beta1.CurrentPriceResponse) |  |  |






<a name="kava.pricefeed.v1beta1.QueryPricesRequest"></a>

### QueryPricesRequest
QueryPricesRequest is the request type for the Query/Prices RPC method.






<a name="kava.pricefeed.v1beta1.QueryPricesResponse"></a>

### QueryPricesResponse
QueryPricesResponse is the response type for the Query/Prices RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `prices` | [CurrentPriceResponse](#kava.pricefeed.v1beta1.CurrentPriceResponse) | repeated |  |






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
| `raw_prices` | [PostedPriceResponse](#kava.pricefeed.v1beta1.PostedPriceResponse) | repeated |  |





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

