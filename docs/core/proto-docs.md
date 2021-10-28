 <!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [kava/swap/v1beta1/genesis.proto](#kava/swap/v1beta1/genesis.proto)
    - [AllowedPool](#kava.swap.v1beta1.AllowedPool)
    - [GenesisState](#kava.swap.v1beta1.GenesisState)
    - [Params](#kava.swap.v1beta1.Params)
    - [PoolRecord](#kava.swap.v1beta1.PoolRecord)
    - [ShareRecord](#kava.swap.v1beta1.ShareRecord)
  
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



<a name="kava/swap/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/swap/v1beta1/genesis.proto



<a name="kava.swap.v1beta1.AllowedPool"></a>

### AllowedPool
AllowedPool defines a tradable pool.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_a` | [string](#string) |  |  |
| `token_b` | [string](#string) |  |  |






<a name="kava.swap.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the swap module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.swap.v1beta1.Params) |  | params defines all the paramaters of the module. |
| `pool_records` | [PoolRecord](#kava.swap.v1beta1.PoolRecord) | repeated |  |
| `share_records` | [ShareRecord](#kava.swap.v1beta1.ShareRecord) | repeated |  |






<a name="kava.swap.v1beta1.Params"></a>

### Params
Params defines the parameters for the swap module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `allowed_pools` | [AllowedPool](#kava.swap.v1beta1.AllowedPool) | repeated |  |
| `swap_fee` | [bytes](#bytes) |  |  |






<a name="kava.swap.v1beta1.PoolRecord"></a>

### PoolRecord
PoolRecord represents the state of a liquidity pool
and is used to store the state of a denominated pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pool_id` | [string](#string) |  |  |
| `reserves_a` | [bytes](#bytes) |  |  |
| `reserves_b` | [bytes](#bytes) |  |  |
| `total_shares` | [bytes](#bytes) |  |  |






<a name="kava.swap.v1beta1.ShareRecord"></a>

### ShareRecord
ShareRecord stores the shares owned for a depositor and pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `depositor` | [string](#string) |  |  |
| `pool_id` | [string](#string) |  |  |
| `shares_owned` | [bytes](#bytes) |  |  |





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
| `depositor` | [string](#string) |  |  |
| `pool_id` | [string](#string) |  |  |
| `shares_owned` | [bytes](#bytes) |  |  |
| `shares_value` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="kava.swap.v1beta1.PoolResponse"></a>

### PoolResponse
PoolStatsQueryResponse defines the coins and shares of a pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `total_shares` | [bytes](#bytes) |  |  |






<a name="kava.swap.v1beta1.QueryDepositsRequest"></a>

### QueryDepositsRequest
QueryDepositsRequest is the request type for the Query/Deposits RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `pool_id` | [string](#string) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="kava.swap.v1beta1.QueryDepositsResponse"></a>

### QueryDepositsResponse
QueryDepositsResponse is the response type for the Query/Deposits RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deposits` | [DepositResponse](#kava.swap.v1beta1.DepositResponse) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="kava.swap.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/swap parameters.






<a name="kava.swap.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/swap parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.swap.v1beta1.Params) |  |  |






<a name="kava.swap.v1beta1.QueryPoolsRequest"></a>

### QueryPoolsRequest
QueryPoolsRequest is the request type for the Query/Pools RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pool_id` | [string](#string) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="kava.swap.v1beta1.QueryPoolsResponse"></a>

### QueryPoolsResponse
QueryPoolsResponse is the response type for the Query/Pools RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pools` | [PoolResponse](#kava.swap.v1beta1.PoolResponse) | repeated |  |
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
| `Deposits` | [QueryDepositsRequest](#kava.swap.v1beta1.QueryDepositsRequest) | [QueryDepositsResponse](#kava.swap.v1beta1.QueryDepositsResponse) | Deposits queries deposit details based on owner address and pool | GET|/kava/swap/v1beta1/deposits|
| `Pools` | [QueryPoolsRequest](#kava.swap.v1beta1.QueryPoolsRequest) | [QueryPoolsResponse](#kava.swap.v1beta1.QueryPoolsResponse) | Pools queries pools based on pool ID | GET|/kava/swap/v1beta1/pools|

 <!-- end services -->



<a name="kava/swap/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/swap/v1beta1/tx.proto



<a name="kava.swap.v1beta1.MsgDeposit"></a>

### MsgDeposit
MsgDeposit represents a message for depositing liquidity into a pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `depositor` | [string](#string) |  |  |
| `token_a` | [bytes](#bytes) |  |  |
| `token_b` | [bytes](#bytes) |  |  |
| `slippage` | [bytes](#bytes) |  |  |
| `deadline` | [int64](#int64) |  |  |






<a name="kava.swap.v1beta1.MsgDepositResponse"></a>

### MsgDepositResponse
MsgDepositResponse defines the Msg/Deposit response type.






<a name="kava.swap.v1beta1.MsgSwapExactForTokens"></a>

### MsgSwapExactForTokens
MsgSwapExactForTokens represents a message for trading exact coinA for coinB


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `requester` | [string](#string) |  |  |
| `exact_token_a` | [bytes](#bytes) |  |  |
| `token_b` | [bytes](#bytes) |  |  |
| `slippage` | [bytes](#bytes) |  |  |
| `deadline` | [int64](#int64) |  |  |






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
| `requester` | [string](#string) |  |  |
| `token_a` | [bytes](#bytes) |  |  |
| `exact_token_b` | [bytes](#bytes) |  |  |
| `slippage` | [bytes](#bytes) |  |  |
| `deadline` | [int64](#int64) |  |  |






<a name="kava.swap.v1beta1.MsgSwapForExactTokensResponse"></a>

### MsgSwapForExactTokensResponse
MsgSwapForExactTokensResponse defines the Msg/SwapForExactTokensResponse
response type.






<a name="kava.swap.v1beta1.MsgWithdraw"></a>

### MsgWithdraw
MsgWithdraw represents a message for withdrawing liquidity from a pool


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from` | [string](#string) |  |  |
| `shares` | [bytes](#bytes) |  |  |
| `min_token_a` | [bytes](#bytes) |  |  |
| `min_token_b` | [bytes](#bytes) |  |  |
| `deadline` | [int64](#int64) |  |  |






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

