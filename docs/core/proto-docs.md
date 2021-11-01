 <!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [kava/bep3/v1beta1/bep3.proto](#kava/bep3/v1beta1/bep3.proto)
    - [AssetParam](#kava.bep3.v1beta1.AssetParam)
    - [AssetSupply](#kava.bep3.v1beta1.AssetSupply)
    - [AtomicSwap](#kava.bep3.v1beta1.AtomicSwap)
    - [AugmentedAtomicSwap](#kava.bep3.v1beta1.AugmentedAtomicSwap)
    - [Params](#kava.bep3.v1beta1.Params)
    - [SupplyLimit](#kava.bep3.v1beta1.SupplyLimit)
  
    - [SwapDirection](#kava.bep3.v1beta1.SwapDirection)
    - [SwapStatus](#kava.bep3.v1beta1.SwapStatus)
  
- [kava/bep3/v1beta1/genesis.proto](#kava/bep3/v1beta1/genesis.proto)
    - [GenesisState](#kava.bep3.v1beta1.GenesisState)
  
- [kava/bep3/v1beta1/query.proto](#kava/bep3/v1beta1/query.proto)
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
  
- [Scalar Value Types](#scalar-value-types)



<a name="kava/bep3/v1beta1/bep3.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/bep3/v1beta1/bep3.proto



<a name="kava.bep3.v1beta1.AssetParam"></a>

### AssetParam
AssetParam defines parameters for each bep3 asset.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `coin_id` | [int64](#int64) |  |  |
| `supply_limit` | [SupplyLimit](#kava.bep3.v1beta1.SupplyLimit) |  |  |
| `active` | [bool](#bool) |  |  |
| `deputy_address` | [string](#string) |  |  |
| `fixed_fee` | [bytes](#bytes) |  |  |
| `min_swap_amount` | [bytes](#bytes) |  |  |
| `max_swap_amount` | [bytes](#bytes) |  |  |
| `min_block_lock` | [uint64](#uint64) |  |  |
| `max_block_lock` | [uint64](#uint64) |  |  |






<a name="kava.bep3.v1beta1.AssetSupply"></a>

### AssetSupply
AssetSupply defines information about an asset's supply.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `incoming_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `outgoing_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `current_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `time_limited_current_supply` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `time_elapsed` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="kava.bep3.v1beta1.AtomicSwap"></a>

### AtomicSwap
AtomicSwap defines an atomic swap between chains for the pricefeed module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `random_number_hash` | [bytes](#bytes) |  |  |
| `expire_height` | [uint64](#uint64) |  |  |
| `timestamp` | [int64](#int64) |  |  |
| `sender` | [string](#string) |  |  |
| `recipient` | [string](#string) |  |  |
| `sender_other_chain` | [string](#string) |  |  |
| `recipient_other_chain` | [string](#string) |  |  |
| `closed_block` | [int64](#int64) |  |  |
| `status` | [SwapStatus](#kava.bep3.v1beta1.SwapStatus) |  |  |
| `cross_chain` | [bool](#bool) |  |  |
| `direction` | [SwapDirection](#kava.bep3.v1beta1.SwapDirection) |  |  |






<a name="kava.bep3.v1beta1.AugmentedAtomicSwap"></a>

### AugmentedAtomicSwap
AugmentedAtomicSwap defines an AtomicSwap with an ID.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `atomic_swap` | [AtomicSwap](#kava.bep3.v1beta1.AtomicSwap) |  |  |






<a name="kava.bep3.v1beta1.Params"></a>

### Params
Params defines the parameters for the bep3 module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_params` | [AssetParam](#kava.bep3.v1beta1.AssetParam) | repeated |  |






<a name="kava.bep3.v1beta1.SupplyLimit"></a>

### SupplyLimit
SupplyLimit define the absolute and time-based limits for an assets's supply.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `limit` | [bytes](#bytes) |  |  |
| `time_limited` | [bool](#bool) |  |  |
| `time_period` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `time_based_limit` | [bytes](#bytes) |  |  |





 <!-- end messages -->


<a name="kava.bep3.v1beta1.SwapDirection"></a>

### SwapDirection
SwapDirection is the direction of an AtomicSwap

| Name | Number | Description |
| ---- | ------ | ----------- |
| SWAP_DIRECTION_UNSPECIFIED | 0 | SwapDirection is unspecified or invalid |
| SWAP_DIRECTION_INCOMING | 1 | SwapDirection is incoming |
| SWAP_DIRECTION_OUTGOING | 2 | SwapDirection is outgoing |



<a name="kava.bep3.v1beta1.SwapStatus"></a>

### SwapStatus
SwapStatus is the status of an AtomicSwap

| Name | Number | Description |
| ---- | ------ | ----------- |
| SWAP_STATUS_UNSPECIFIED | 0 | AtomicSwap is unspecified |
| SWAP_STATUS_OPEN | 1 | AtomicSwap is open |
| SWAP_STATUS_COMPLETED | 2 | AtomicSwap is completed |
| SWAP_STATUS_EXPIRED | 3 | AtomicSwap is expired |


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
| `atomic_swaps` | [AtomicSwap](#kava.bep3.v1beta1.AtomicSwap) | repeated |  |
| `supplies` | [AssetSupply](#kava.bep3.v1beta1.AssetSupply) | repeated |  |
| `previous_block_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="kava/bep3/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## kava/bep3/v1beta1/query.proto



<a name="kava.bep3.v1beta1.QueryAssetSuppliesRequest"></a>

### QueryAssetSuppliesRequest
QueryAssetSuppliesRequest is the request type for the Query/AssetSupplies RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="kava.bep3.v1beta1.QueryAssetSuppliesResponse"></a>

### QueryAssetSuppliesResponse
QueryAssetSuppliesResponse is the response type for the Query/AssetSupplies RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_supplies` | [AssetSupply](#kava.bep3.v1beta1.AssetSupply) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |






<a name="kava.bep3.v1beta1.QueryAssetSupplyRequest"></a>

### QueryAssetSupplyRequest
QueryAssetSupplyRequest is the request type for the Query/AssetSupply RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="kava.bep3.v1beta1.QueryAssetSupplyResponse"></a>

### QueryAssetSupplyResponse
QueryAssetSupplyResponse is the response type for the Query/AssetSupply RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_supply` | [AssetSupply](#kava.bep3.v1beta1.AssetSupply) |  |  |






<a name="kava.bep3.v1beta1.QueryAtomicSwapRequest"></a>

### QueryAtomicSwapRequest
QueryAtomicSwapRequest is the request type for the Query/AtomicSwap RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `swap_id` | [bytes](#bytes) |  |  |






<a name="kava.bep3.v1beta1.QueryAtomicSwapResponse"></a>

### QueryAtomicSwapResponse
QueryAtomicSwapResponse is the response type for the Query/AtomicSwap RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `atomic_swap` | [AtomicSwap](#kava.bep3.v1beta1.AtomicSwap) |  |  |






<a name="kava.bep3.v1beta1.QueryAtomicSwapsRequest"></a>

### QueryAtomicSwapsRequest
QueryAtomicSwapsRequest is the request type for the Query/AtomicSwaps RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `involve` | [string](#string) |  |  |
| `expiration` | [uint64](#uint64) |  |  |
| `status` | [SwapStatus](#kava.bep3.v1beta1.SwapStatus) |  |  |
| `direction` | [SwapDirection](#kava.bep3.v1beta1.SwapDirection) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="kava.bep3.v1beta1.QueryAtomicSwapsResponse"></a>

### QueryAtomicSwapsResponse
QueryAtomicSwapsResponse is the response type for the Query/AtomicSwaps RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `atomic_swap` | [AugmentedAtomicSwap](#kava.bep3.v1beta1.AugmentedAtomicSwap) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |






<a name="kava.bep3.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/bep3 parameters.






<a name="kava.bep3.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/bep3 parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#kava.bep3.v1beta1.Params) |  |  |





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
| `random_number_hash` | [bytes](#bytes) |  |  |
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

