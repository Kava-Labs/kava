 <!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [kava/cdp/v1beta1/cdp.proto](#kava/cdp/v1beta1/cdp.proto)
    - [AugmentedCDP](#kava.cdp.v1beta1.AugmentedCDP)
    - [CDP](#kava.cdp.v1beta1.CDP)
    - [Deposit](#kava.cdp.v1beta1.Deposit)
    - [TotalCollateral](#kava.cdp.v1beta1.TotalCollateral)
    - [TotalPrincipal](#kava.cdp.v1beta1.TotalPrincipal)
  
- [kava/cdp/v1beta1/genesis.proto](#kava/cdp/v1beta1/genesis.proto)
    - [CollateralParam](#kava.cdp.v1beta1.CollateralParam)
    - [DebtParam](#kava.cdp.v1beta1.DebtParam)
    - [GenesisAccumulationTime](#kava.cdp.v1beta1.GenesisAccumulationTime)
    - [GenesisState](#kava.cdp.v1beta1.GenesisState)
    - [GenesisTotalPrincipal](#kava.cdp.v1beta1.GenesisTotalPrincipal)
    - [Params](#kava.cdp.v1beta1.Params)
  
- [kava/cdp/v1beta1/query.proto](#kava/cdp/v1beta1/query.proto)
    - [QueryAccountsRequest](#kava.cdp.v1beta1.QueryAccountsRequest)
    - [QueryAccountsResponse](#kava.cdp.v1beta1.QueryAccountsResponse)
    - [QueryCdpRequest](#kava.cdp.v1beta1.QueryCdpRequest)
    - [QueryCdpResponse](#kava.cdp.v1beta1.QueryCdpResponse)
    - [QueryCdpsByCollateralTypeRequest](#kava.cdp.v1beta1.QueryCdpsByCollateralTypeRequest)
    - [QueryCdpsByCollateralTypeResponse](#kava.cdp.v1beta1.QueryCdpsByCollateralTypeResponse)
    - [QueryCdpsByRatioRequest](#kava.cdp.v1beta1.QueryCdpsByRatioRequest)
    - [QueryCdpsByRatioResponse](#kava.cdp.v1beta1.QueryCdpsByRatioResponse)
    - [QueryCdpsRequest](#kava.cdp.v1beta1.QueryCdpsRequest)
    - [QueryCdpsResponse](#kava.cdp.v1beta1.QueryCdpsResponse)
    - [QueryDepositsRequest](#kava.cdp.v1beta1.QueryDepositsRequest)
    - [QueryDepositsResponse](#kava.cdp.v1beta1.QueryDepositsResponse)
    - [QueryParamsRequest](#kava.cdp.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#kava.cdp.v1beta1.QueryParamsResponse)
    - [QueryTotalCollateralRequest](#kava.cdp.v1beta1.QueryTotalCollateralRequest)
    - [QueryTotalCollateralResponse](#kava.cdp.v1beta1.QueryTotalCollateralResponse)
    - [QueryTotalPrincipalRequest](#kava.cdp.v1beta1.QueryTotalPrincipalRequest)
    - [QueryTotalPrincipalResponse](#kava.cdp.v1beta1.QueryTotalPrincipalResponse)
  
    - [Query](#kava.cdp.v1beta1.Query)
  
- [kava/cdp/v1beta1/tx.proto](#kava/cdp/v1beta1/tx.proto)
    - [MsgCreateCDP](#kava.cdp.v1beta1.MsgCreateCDP)
    - [MsgCreateCDPResponse](#kava.cdp.v1beta1.MsgCreateCDPResponse)
    - [MsgDeposit](#kava.cdp.v1beta1.MsgDeposit)
    - [MsgDepositResponse](#kava.cdp.v1beta1.MsgDepositResponse)
    - [MsgDrawDebt](#kava.cdp.v1beta1.MsgDrawDebt)
    - [MsgDrawDebtResponse](#kava.cdp.v1beta1.MsgDrawDebtResponse)
    - [MsgLiquidate](#kava.cdp.v1beta1.MsgLiquidate)
    - [MsgLiquidateResponse](#kava.cdp.v1beta1.MsgLiquidateResponse)
    - [MsgRepayDebt](#kava.cdp.v1beta1.MsgRepayDebt)
    - [MsgRepayDebtResponse](#kava.cdp.v1beta1.MsgRepayDebtResponse)
    - [MsgWithdraw](#kava.cdp.v1beta1.MsgWithdraw)
    - [MsgWithdrawResponse](#kava.cdp.v1beta1.MsgWithdrawResponse)
  
    - [Msg](#kava.cdp.v1beta1.Msg)
  
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
  
- [Scalar Value Types](#scalar-value-types)



<a name="kava/cdp/v1beta1/cdp.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/cdp/v1beta1/cdp.proto



<a name="kava.cdp.v1beta1.AugmentedCDP"></a>

### AugmentedCDP
AugmentedCDP defines additional information about an active CDP


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cdp` | [CDP](#kava.cdp.v1beta1.CDP) |  |  |
| `collateral_value` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `collateralization_ratio` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.CDP"></a>

### CDP
CDP defines the state of a single collateralized debt position.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  |  |
| `owner` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |
| `collateral` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `principal` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `accumulated_fees` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `fees_updated` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `interest_factor` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.Deposit"></a>

### Deposit
Deposit defines an amount of coins deposited by an account to a cdp


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cdp_id` | [uint64](#uint64) |  |  |
| `depositor` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="kava.cdp.v1beta1.TotalCollateral"></a>

### TotalCollateral
TotalCollateral defines the total collateral of a given collateral type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="kava.cdp.v1beta1.TotalPrincipal"></a>

### TotalPrincipal
TotalPrincipal defines the total principal of a given collateral type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/cdp/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/cdp/v1beta1/genesis.proto



<a name="kava.cdp.v1beta1.CollateralParam"></a>

### CollateralParam
CollateralParam defines governance parameters for each collateral type within the cdp module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |
| `liquidation_ratio` | [string](#string) |  |  |
| `debt_limit` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `stability_fee` | [string](#string) |  |  |
| `auction_size` | [string](#string) |  |  |
| `liquidation_penalty` | [string](#string) |  |  |
| `prefix` | [uint32](#uint32) |  | No single byte type, use a uint32 |
| `spot_market_id` | [string](#string) |  |  |
| `liquidation_market_id` | [string](#string) |  |  |
| `keeper_reward_percentage` | [string](#string) |  |  |
| `check_collateralization_index_count` | [string](#string) |  |  |
| `conversion_factor` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.DebtParam"></a>

### DebtParam
DebtParam defines governance params for debt assets


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `reference_asset` | [string](#string) |  |  |
| `conversion_factor` | [string](#string) |  |  |
| `debt_floor` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.GenesisAccumulationTime"></a>

### GenesisAccumulationTime
GenesisAccumulationTime defines the previous distribution time and its corresponding denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |
| `previous_accumulation_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `interest_factor` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the cdp module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.cdp.v1beta1.Params) |  | params defines all the paramaters of the module. |
| `cdps` | [CDP](#kava.cdp.v1beta1.CDP) | repeated |  |
| `deposits` | [Deposit](#kava.cdp.v1beta1.Deposit) | repeated |  |
| `starting_cdp_id` | [uint64](#uint64) |  |  |
| `debt_denom` | [string](#string) |  |  |
| `gov_denom` | [string](#string) |  |  |
| `previous_accumulation_times` | [GenesisAccumulationTime](#kava.cdp.v1beta1.GenesisAccumulationTime) | repeated |  |
| `total_principals` | [GenesisTotalPrincipal](#kava.cdp.v1beta1.GenesisTotalPrincipal) | repeated |  |






<a name="kava.cdp.v1beta1.GenesisTotalPrincipal"></a>

### GenesisTotalPrincipal
GenesisTotalPrincipal defines the total principal and its corresponding collateral type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |
| `total_principal` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.Params"></a>

### Params
Params defines the parameters for the cdp module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_params` | [CollateralParam](#kava.cdp.v1beta1.CollateralParam) | repeated |  |
| `debt_param` | [DebtParam](#kava.cdp.v1beta1.DebtParam) |  |  |
| `global_debt_limit` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `surplus_auction_threshold` | [string](#string) |  |  |
| `surplus_auction_lot` | [string](#string) |  |  |
| `debt_auction_threshold` | [string](#string) |  |  |
| `debt_auction_lot` | [string](#string) |  |  |
| `circuit_breaker` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/cdp/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/cdp/v1beta1/query.proto



<a name="kava.cdp.v1beta1.QueryAccountsRequest"></a>

### QueryAccountsRequest
QueryAccountsRequest defines the request type for the Query/Accounts RPC method.






<a name="kava.cdp.v1beta1.QueryAccountsResponse"></a>

### QueryAccountsResponse
QueryAccountsResponse defines the response type for the Query/Accounts RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [cosmos.auth.v1beta1.ModuleAccount](#cosmos.auth.v1beta1.ModuleAccount) | repeated |  |






<a name="kava.cdp.v1beta1.QueryCdpRequest"></a>

### QueryCdpRequest
QueryCdpRequest defines the request type for the Query/Cdp RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.QueryCdpResponse"></a>

### QueryCdpResponse
QueryCdpResponse defines the response type for the Query/Cdp RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cdp` | [AugmentedCDP](#kava.cdp.v1beta1.AugmentedCDP) |  |  |






<a name="kava.cdp.v1beta1.QueryCdpsByCollateralTypeRequest"></a>

### QueryCdpsByCollateralTypeRequest
QueryCdpsByCollateralTypeRequest defines the request type for the Query/CdpsByCollateralType RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.QueryCdpsByCollateralTypeResponse"></a>

### QueryCdpsByCollateralTypeResponse
QueryCdpsByCollateralTypeResponse defines the response type for the Query/CdpsByCollateralType RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cdps` | [AugmentedCDP](#kava.cdp.v1beta1.AugmentedCDP) | repeated |  |






<a name="kava.cdp.v1beta1.QueryCdpsByRatioRequest"></a>

### QueryCdpsByRatioRequest
QueryCdpsByCollateralTypeRequest defines the request type for the Query/CdpsByRatio RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |
| `ratio` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.QueryCdpsByRatioResponse"></a>

### QueryCdpsByRatioResponse
QueryCdpsByRatioResponse defines the response type for the Query/CdpsByRatio RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cdps` | [AugmentedCDP](#kava.cdp.v1beta1.AugmentedCDP) | repeated |  |






<a name="kava.cdp.v1beta1.QueryCdpsRequest"></a>

### QueryCdpsRequest
QueryCdpsRequest is the params for a filtered CDP query, the request type for the Query/Cdps RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |
| `id` | [uint64](#uint64) |  |  |
| `ratio` | [string](#string) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="kava.cdp.v1beta1.QueryCdpsResponse"></a>

### QueryCdpsResponse
QueryCdpsResponse defines the response type for the Query/Cdps RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `cdps` | [AugmentedCDP](#kava.cdp.v1beta1.AugmentedCDP) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |






<a name="kava.cdp.v1beta1.QueryDepositsRequest"></a>

### QueryDepositsRequest
QueryDepositsRequest defines the request type for the Query/Deposits RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.QueryDepositsResponse"></a>

### QueryDepositsResponse
QueryDepositsResponse defines the response type for the Query/Deposits RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposits` | [Deposit](#kava.cdp.v1beta1.Deposit) | repeated |  |






<a name="kava.cdp.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for the Query/Params RPC method.






<a name="kava.cdp.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.cdp.v1beta1.Params) |  |  |






<a name="kava.cdp.v1beta1.QueryTotalCollateralRequest"></a>

### QueryTotalCollateralRequest
QueryTotalCollateralRequest defines the request type for the Query/TotalCollateral RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.QueryTotalCollateralResponse"></a>

### QueryTotalCollateralResponse
QueryTotalCollateralResponse defines the response type for the Query/TotalCollateral RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `total_collateral` | [TotalCollateral](#kava.cdp.v1beta1.TotalCollateral) | repeated |  |






<a name="kava.cdp.v1beta1.QueryTotalPrincipalRequest"></a>

### QueryTotalPrincipalRequest
QueryTotalPrincipalRequest defines the request type for the Query/TotalPrincipal RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `collateral_type` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.QueryTotalPrincipalResponse"></a>

### QueryTotalPrincipalResponse
QueryTotalPrincipalResponse defines the response type for the Query/TotalPrincipal RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `total_principal` | [TotalPrincipal](#kava.cdp.v1beta1.TotalPrincipal) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.cdp.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service for cdp module

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#kava.cdp.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#kava.cdp.v1beta1.QueryParamsResponse) | Params queries all parameters of the cdp module. | GET|/kava/cdp/v1beta1/params|
| `Accounts` | [QueryAccountsRequest](#kava.cdp.v1beta1.QueryAccountsRequest) | [QueryAccountsResponse](#kava.cdp.v1beta1.QueryAccountsResponse) | Accounts queries the CDP module accounts. | GET|/kava/cdp/v1beta1/accounts|
| `TotalPrincipal` | [QueryTotalPrincipalRequest](#kava.cdp.v1beta1.QueryTotalPrincipalRequest) | [QueryTotalPrincipalResponse](#kava.cdp.v1beta1.QueryTotalPrincipalResponse) | TotalPrincipal queries the total principal of a given collateral type. | GET|/kava/cdp/v1beta1/totalPrincipal|
| `TotalCollateral` | [QueryTotalCollateralRequest](#kava.cdp.v1beta1.QueryTotalCollateralRequest) | [QueryTotalCollateralResponse](#kava.cdp.v1beta1.QueryTotalCollateralResponse) | TotalCollateral queries the total collateral of a given collateral type. | GET|/kava/cdp/v1beta1/totalCollateral|
| `Cdps` | [QueryCdpsRequest](#kava.cdp.v1beta1.QueryCdpsRequest) | [QueryCdpsResponse](#kava.cdp.v1beta1.QueryCdpsResponse) | Cdps queries all active CDPs. | GET|/kava/cdp/v1beta1/cdps|
| `Cdp` | [QueryCdpRequest](#kava.cdp.v1beta1.QueryCdpRequest) | [QueryCdpResponse](#kava.cdp.v1beta1.QueryCdpResponse) | Cdp queries a CDP with the input owner address and collateral type. | GET|/kava/cdp/v1beta1/cdps/{owner}/{collateral_type}|
| `Deposits` | [QueryDepositsRequest](#kava.cdp.v1beta1.QueryDepositsRequest) | [QueryDepositsResponse](#kava.cdp.v1beta1.QueryDepositsResponse) | Deposits queries deposits associated with the CDP owned by an address for a collateral type. | GET|/kava/cdp/v1beta1/cdps/deposits/{owner}/{collateral_type}|
| `CdpsByCollateralType` | [QueryCdpsByCollateralTypeRequest](#kava.cdp.v1beta1.QueryCdpsByCollateralTypeRequest) | [QueryCdpsByCollateralTypeResponse](#kava.cdp.v1beta1.QueryCdpsByCollateralTypeResponse) | CdpsByCollateralType queries all CDPs with the collateral type equal to the input collateral type. | GET|/kava/cdp/v1beta1/cdps/collateralType/{collateral_type}|
| `CdpsByRatio` | [QueryCdpsByRatioRequest](#kava.cdp.v1beta1.QueryCdpsByRatioRequest) | [QueryCdpsByRatioResponse](#kava.cdp.v1beta1.QueryCdpsByRatioResponse) | CdpsByRatio queries all CDPs with the collateral type equal to the input colalteral type and collateralization ratio strictly less than the input ratio. | GET|/kava/cdp/v1beta1/cdps/ratio/{collateral_type}|

 <!-- end services -->



<a name="kava/cdp/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/cdp/v1beta1/tx.proto



<a name="kava.cdp.v1beta1.MsgCreateCDP"></a>

### MsgCreateCDP
MsgCreateCDP defines a message to create a new CDP.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `collateral` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `principal` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `collateral_type` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.MsgCreateCDPResponse"></a>

### MsgCreateCDPResponse
MsgCreateCDPResponse defines the Msg/CreateCDP response type.






<a name="kava.cdp.v1beta1.MsgDeposit"></a>

### MsgDeposit
MsgDeposit defines a message to deposit to a CDP.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `depositor` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |
| `collateral` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `collateral_type` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.MsgDepositResponse"></a>

### MsgDepositResponse
MsgDepositResponse defines the Msg/Deposit response type.






<a name="kava.cdp.v1beta1.MsgDrawDebt"></a>

### MsgDrawDebt
MsgDrawDebt defines a message to draw debt from a CDP.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `collateral_type` | [string](#string) |  |  |
| `principal` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="kava.cdp.v1beta1.MsgDrawDebtResponse"></a>

### MsgDrawDebtResponse
MsgDrawDebtResponse defines the Msg/DrawDebt response type.






<a name="kava.cdp.v1beta1.MsgLiquidate"></a>

### MsgLiquidate
MsgLiquidate defines a message to attempt to liquidate a CDP whos
collateralization ratio is under its liquidation ratio.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keeper` | [string](#string) |  |  |
| `borrower` | [string](#string) |  |  |
| `collateral_type` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.MsgLiquidateResponse"></a>

### MsgLiquidateResponse
MsgLiquidateResponse defines the Msg/Liquidate response type.






<a name="kava.cdp.v1beta1.MsgRepayDebt"></a>

### MsgRepayDebt
MsgRepayDebt defines a message to repay debt from a CDP.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `collateral_type` | [string](#string) |  |  |
| `payment` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="kava.cdp.v1beta1.MsgRepayDebtResponse"></a>

### MsgRepayDebtResponse
MsgRepayDebtResponse defines the Msg/RepayDebt response type.






<a name="kava.cdp.v1beta1.MsgWithdraw"></a>

### MsgWithdraw
MsgWithdraw defines a message to withdraw collateral from a CDP.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `depositor` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |
| `collateral` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `collateral_type` | [string](#string) |  |  |






<a name="kava.cdp.v1beta1.MsgWithdrawResponse"></a>

### MsgWithdrawResponse
MsgWithdrawResponse defines the Msg/Withdraw response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="kava.cdp.v1beta1.Msg"></a>

### Msg
Msg defines the cdp Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateCDP` | [MsgCreateCDP](#kava.cdp.v1beta1.MsgCreateCDP) | [MsgCreateCDPResponse](#kava.cdp.v1beta1.MsgCreateCDPResponse) | CreateCDP defines a method to create a new CDP. | |
| `Deposit` | [MsgDeposit](#kava.cdp.v1beta1.MsgDeposit) | [MsgDepositResponse](#kava.cdp.v1beta1.MsgDepositResponse) | Deposit defines a method to deposit to a CDP. | |
| `Withdraw` | [MsgWithdraw](#kava.cdp.v1beta1.MsgWithdraw) | [MsgWithdrawResponse](#kava.cdp.v1beta1.MsgWithdrawResponse) | Withdraw defines a method to withdraw collateral from a CDP. | |
| `DrawDebt` | [MsgDrawDebt](#kava.cdp.v1beta1.MsgDrawDebt) | [MsgDrawDebtResponse](#kava.cdp.v1beta1.MsgDrawDebtResponse) | DrawDebt defines a method to draw debt from a CDP. | |
| `RepayDebt` | [MsgRepayDebt](#kava.cdp.v1beta1.MsgRepayDebt) | [MsgRepayDebtResponse](#kava.cdp.v1beta1.MsgRepayDebtResponse) | RepayDebt defines a method to repay debt from a CDP. | |
| `Liquidate` | [MsgLiquidate](#kava.cdp.v1beta1.MsgLiquidate) | [MsgLiquidateResponse](#kava.cdp.v1beta1.MsgLiquidateResponse) | Liquidate defines a method to attempt to liquidate a CDP whos collateralization ratio is under its liquidation ratio. | |

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

