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
| `asset_supplies` | [AssetSupply](#kava.issuance.v1beta1.AssetSupply) | repeated |  |






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

