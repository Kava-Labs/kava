 <!-- This file is auto-generated. Please do not modify it yourself. -->

# Protobuf Documentation

<a name="top"></a>

## Table of Contents

- [kava/auction/v1beta1/auction.proto](#kava/auction/v1beta1/auction.proto)

  - [BaseAuction](#kava.auction.v1beta1.BaseAuction)
  - [CollateralAuction](#kava.auction.v1beta1.CollateralAuction)
  - [DebtAuction](#kava.auction.v1beta1.DebtAuction)
  - [SurplusAuction](#kava.auction.v1beta1.SurplusAuction)
  - [WeightedAddresses](#kava.auction.v1beta1.WeightedAddresses)

- [kava/auction/v1beta1/genesis.proto](#kava/auction/v1beta1/genesis.proto)

  - [GenesisState](#kava.auction.v1beta1.GenesisState)
  - [Params](#kava.auction.v1beta1.Params)

- [kava/auction/v1beta1/query.proto](#kava/auction/v1beta1/query.proto)

  - [QueryAuctionRequest](#kava.auction.v1beta1.QueryAuctionRequest)
  - [QueryAuctionResponse](#kava.auction.v1beta1.QueryAuctionResponse)
  - [QueryAuctionsRequest](#kava.auction.v1beta1.QueryAuctionsRequest)
  - [QueryAuctionsResponse](#kava.auction.v1beta1.QueryAuctionsResponse)
  - [QueryNextAuctionIDRequest](#kava.auction.v1beta1.QueryNextAuctionIDRequest)
  - [QueryNextAuctionIDResponse](#kava.auction.v1beta1.QueryNextAuctionIDResponse)
  - [QueryParamsRequest](#kava.auction.v1beta1.QueryParamsRequest)
  - [QueryParamsResponse](#kava.auction.v1beta1.QueryParamsResponse)

  - [Query](#kava.auction.v1beta1.Query)

- [kava/auction/v1beta1/tx.proto](#kava/auction/v1beta1/tx.proto)

  - [MsgPlaceBid](#kava.auction.v1beta1.MsgPlaceBid)
  - [MsgPlaceBidResponse](#kava.auction.v1beta1.MsgPlaceBidResponse)

  - [Msg](#kava.auction.v1beta1.Msg)

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

<a name="kava/auction/v1beta1/auction.proto"></a>

<p align="right"><a href="#top">Top</a></p>

## kava/auction/v1beta1/auction.proto

<a name="kava.auction.v1beta1.BaseAuction"></a>

### BaseAuction

BaseAuction defines common attributes of all auctions

| Field               | Type                                                    | Label | Description |
| ------------------- | ------------------------------------------------------- | ----- | ----------- |
| `id`                | [uint64](#uint64)                                       |       |             |
| `initiator`         | [string](#string)                                       |       |             |
| `lot`               | [bytes](#bytes)                                         |       |             |
| `bidder`            | [string](#string)                                       |       |             |
| `bid`               | [bytes](#bytes)                                         |       |             |
| `has_received_bids` | [bool](#bool)                                           |       |             |
| `end_time`          | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |       |             |
| `max_end_time`      | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |       |             |

<a name="kava.auction.v1beta1.CollateralAuction"></a>

### CollateralAuction

CollateralAuction is a two phase auction.
Initially, in forward auction phase, bids can be placed up to a max bid.
Then it switches to a reverse auction phase, where the initial amount up for auction is bid down.
Unsold Lot is sent to LotReturns, being divided among the addresses by weight.
Collateral auctions are normally used to sell off collateral seized from CDPs.

| Field                | Type                                                         | Label | Description |
| -------------------- | ------------------------------------------------------------ | ----- | ----------- |
| `base_auction`       | [BaseAuction](#kava.auction.v1beta1.BaseAuction)             |       |             |
| `corresponding_debt` | [bytes](#bytes)                                              |       |             |
| `max_bid`            | [bytes](#bytes)                                              |       |             |
| `lot_returns`        | [WeightedAddresses](#kava.auction.v1beta1.WeightedAddresses) |       |             |

<a name="kava.auction.v1beta1.DebtAuction"></a>

### DebtAuction

DebtAuction is a reverse auction that mints what it pays out.
It is normally used to acquire pegged asset to cover the CDP system's debts that were not covered by selling
collateral.

| Field                | Type                                             | Label | Description |
| -------------------- | ------------------------------------------------ | ----- | ----------- |
| `base_auction`       | [BaseAuction](#kava.auction.v1beta1.BaseAuction) |       |             |
| `corresponding_debt` | [bytes](#bytes)                                  |       |             |

<a name="kava.auction.v1beta1.SurplusAuction"></a>

### SurplusAuction

SurplusAuction is a forward auction that burns what it receives from bids.
It is normally used to sell off excess pegged asset acquired by the CDP system.

| Field          | Type                                             | Label | Description |
| -------------- | ------------------------------------------------ | ----- | ----------- |
| `base_auction` | [BaseAuction](#kava.auction.v1beta1.BaseAuction) |       |             |

<a name="kava.auction.v1beta1.WeightedAddresses"></a>

### WeightedAddresses

WeightedAddresses is a type for storing some addresses and associated weights.

| Field       | Type              | Label    | Description |
| ----------- | ----------------- | -------- | ----------- |
| `addresses` | [string](#string) | repeated |             |
| `weights`   | [bytes](#bytes)   | repeated |             |

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

<a name="kava/auction/v1beta1/genesis.proto"></a>

<p align="right"><a href="#top">Top</a></p>

## kava/auction/v1beta1/genesis.proto

<a name="kava.auction.v1beta1.GenesisState"></a>

### GenesisState

GenesisState defines the auction module's genesis state.

| Field             | Type                                        | Label    | Description      |
| ----------------- | ------------------------------------------- | -------- | ---------------- |
| `next_auction_id` | [uint64](#uint64)                           |          |                  |
| `params`          | [Params](#kava.auction.v1beta1.Params)      |          |                  |
| `auctions`        | [google.protobuf.Any](#google.protobuf.Any) | repeated | Genesis auctions |

<a name="kava.auction.v1beta1.Params"></a>

### Params

Params defines the parameters for the issuance module.

| Field                  | Type                                                  | Label | Description |
| ---------------------- | ----------------------------------------------------- | ----- | ----------- |
| `max_auction_duration` | [google.protobuf.Duration](#google.protobuf.Duration) |       |             |
| `bid_duration`         | [google.protobuf.Duration](#google.protobuf.Duration) |       |             |
| `increment_surplus`    | [bytes](#bytes)                                       |       |             |
| `increment_debt`       | [bytes](#bytes)                                       |       |             |
| `increment_collateral` | [bytes](#bytes)                                       |       |             |

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

<a name="kava/auction/v1beta1/query.proto"></a>

<p align="right"><a href="#top">Top</a></p>

## kava/auction/v1beta1/query.proto

<a name="kava.auction.v1beta1.QueryAuctionRequest"></a>

### QueryAuctionRequest

QueryAuctionRequest is the request type for the Query/Auction RPC method.

| Field        | Type              | Label | Description |
| ------------ | ----------------- | ----- | ----------- |
| `auction_id` | [string](#string) |       |             |

<a name="kava.auction.v1beta1.QueryAuctionResponse"></a>

### QueryAuctionResponse

QueryAuctionResponse is the response type for the Query/Auction RPC method.

| Field     | Type                                        | Label | Description |
| --------- | ------------------------------------------- | ----- | ----------- |
| `auction` | [google.protobuf.Any](#google.protobuf.Any) |       |             |

<a name="kava.auction.v1beta1.QueryAuctionsRequest"></a>

### QueryAuctionsRequest

QueryAuctionsRequest is the request type for the Query/Auctions RPC method.

| Field        | Type                                                                            | Label | Description                                                |
| ------------ | ------------------------------------------------------------------------------- | ----- | ---------------------------------------------------------- |
| `type`       | [string](#string)                                                               |       |                                                            |
| `owner`      | [string](#string)                                                               |       |                                                            |
| `denom`      | [string](#string)                                                               |       |                                                            |
| `phase`      | [string](#string)                                                               |       |                                                            |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |       | pagination defines an optional pagination for the request. |

<a name="kava.auction.v1beta1.QueryAuctionsResponse"></a>

### QueryAuctionsResponse

QueryAuctionsResponse is the response type for the Query/Auctions RPC method.

| Field        | Type                                                                              | Label    | Description                                        |
| ------------ | --------------------------------------------------------------------------------- | -------- | -------------------------------------------------- |
| `auction`    | [google.protobuf.Any](#google.protobuf.Any)                                       | repeated |                                                    |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |          | pagination defines the pagination in the response. |

<a name="kava.auction.v1beta1.QueryNextAuctionIDRequest"></a>

### QueryNextAuctionIDRequest

QueryNextAuctionIDRequest defines the request type for querying x/auction next auction ID.

<a name="kava.auction.v1beta1.QueryNextAuctionIDResponse"></a>

### QueryNextAuctionIDResponse

QueryNextAuctionIDResponse defines the response type for querying x/auction next auction ID.

| Field | Type              | Label | Description |
| ----- | ----------------- | ----- | ----------- |
| `id`  | [uint64](#uint64) |       |             |

<a name="kava.auction.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest

QueryParamsRequest defines the request type for querying x/auction parameters.

<a name="kava.auction.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse

QueryParamsResponse defines the response type for querying x/auction parameters.

| Field    | Type                                   | Label | Description |
| -------- | -------------------------------------- | ----- | ----------- |
| `params` | [Params](#kava.auction.v1beta1.Params) |       |             |

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

<a name="kava.auction.v1beta1.Query"></a>

### Query

Query defines the gRPC querier service for auction module

| Method Name     | Request Type                                                                 | Response Type                                                                  | Description                                                                               | HTTP Verb | Endpoint                 |
| --------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------- | --------- | ------------------------ |
| `Params`        | [QueryParamsRequest](#kava.auction.v1beta1.QueryParamsRequest)               | [QueryParamsResponse](#kava.auction.v1beta1.QueryParamsResponse)               | Params queries all parameters of the auction module.                                      | GET       | /auction/params          |
| `Auction`       | [QueryAuctionRequest](#kava.auction.v1beta1.QueryAuctionRequest)             | [QueryAuctionResponse](#kava.auction.v1beta1.QueryAuctionResponse)             | Auction queries an individual Auction by auction ID                                       | GET       | /auction/auction         |
| `Auctions`      | [QueryAuctionsRequest](#kava.auction.v1beta1.QueryAuctionsRequest)           | [QueryAuctionsResponse](#kava.auction.v1beta1.QueryAuctionsResponse)           | Auctions queries auctions filtered by asset denom, owner address, phase, and auction type | GET       | /auction/auctions        |
| `NextAuctionID` | [QueryNextAuctionIDRequest](#kava.auction.v1beta1.QueryNextAuctionIDRequest) | [QueryNextAuctionIDResponse](#kava.auction.v1beta1.QueryNextAuctionIDResponse) | NextAuctionID queries the next auction ID                                                 | GET       | /auction/next-auction-id |

 <!-- end services -->

<a name="kava/auction/v1beta1/tx.proto"></a>

<p align="right"><a href="#top">Top</a></p>

## kava/auction/v1beta1/tx.proto

<a name="kava.auction.v1beta1.MsgPlaceBid"></a>

### MsgPlaceBid

MsgPlaceBid represents a message used by bidders to place bids on auctions

| Field        | Type              | Label | Description |
| ------------ | ----------------- | ----- | ----------- |
| `auction_id` | [uint64](#uint64) |       |             |
| `bidder`     | [string](#string) |       |             |
| `amount`     | [bytes](#bytes)   |       |             |

<a name="kava.auction.v1beta1.MsgPlaceBidResponse"></a>

### MsgPlaceBidResponse

MsgPlaceBidResponse defines the Msg/PlaceBid response type.

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

<a name="kava.auction.v1beta1.Msg"></a>

### Msg

Msg defines the auction Msg service.

| Method Name | Request Type                                     | Response Type                                                    | Description                                                     | HTTP Verb | Endpoint |
| ----------- | ------------------------------------------------ | ---------------------------------------------------------------- | --------------------------------------------------------------- | --------- | -------- |
| `PlaceBid`  | [MsgPlaceBid](#kava.auction.v1beta1.MsgPlaceBid) | [MsgPlaceBidResponse](#kava.auction.v1beta1.MsgPlaceBidResponse) | PlaceBid message type used by bidders to place bids on auctions |           |

 <!-- end services -->

<a name="kava/kavadist/v1beta1/params.proto"></a>

<p align="right"><a href="#top">Top</a></p>

## kava/kavadist/v1beta1/params.proto

<a name="kava.kavadist.v1beta1.Params"></a>

### Params

Params governance parameters for kavadist module

| Field     | Type                                    | Label    | Description |
| --------- | --------------------------------------- | -------- | ----------- |
| `active`  | [bool](#bool)                           |          |             |
| `periods` | [Period](#kava.kavadist.v1beta1.Period) | repeated |             |

<a name="kava.kavadist.v1beta1.Period"></a>

### Period

Period stores the specified start and end dates, and the inflation, expressed as a decimal
representing the yearly APR of KAVA tokens that will be minted during that period

| Field       | Type                                                    | Label | Description                                    |
| ----------- | ------------------------------------------------------- | ----- | ---------------------------------------------- |
| `start`     | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |       | example "2020-03-01T15:20:00Z"                 |
| `end`       | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |       | example "2020-06-01T15:20:00Z"                 |
| `inflation` | [bytes](#bytes)                                         |       | example "1.000000003022265980" - 10% inflation |

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

| Field                 | Type                                                    | Label | Description |
| --------------------- | ------------------------------------------------------- | ----- | ----------- |
| `params`              | [Params](#kava.kavadist.v1beta1.Params)                 |       |             |
| `previous_block_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |       |             |

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

| Field            | Type                                                              | Label    | Description |
| ---------------- | ----------------------------------------------------------------- | -------- | ----------- |
| `title`          | [string](#string)                                                 |          |             |
| `description`    | [string](#string)                                                 |          |             |
| `recipient_list` | [MultiSpendRecipient](#kava.kavadist.v1beta1.MultiSpendRecipient) | repeated |             |

<a name="kava.kavadist.v1beta1.CommunityPoolMultiSpendProposalJSON"></a>

### CommunityPoolMultiSpendProposalJSON

CommunityPoolMultiSpendProposalJSON defines a CommunityPoolMultiSpendProposal with a deposit

| Field            | Type                                                              | Label    | Description |
| ---------------- | ----------------------------------------------------------------- | -------- | ----------- |
| `title`          | [string](#string)                                                 |          |             |
| `description`    | [string](#string)                                                 |          |             |
| `recipient_list` | [MultiSpendRecipient](#kava.kavadist.v1beta1.MultiSpendRecipient) | repeated |             |
| `deposit`        | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin)             | repeated |             |

<a name="kava.kavadist.v1beta1.MultiSpendRecipient"></a>

### MultiSpendRecipient

MultiSpendRecipient defines a recipient and the amount of coins they are receiving

| Field     | Type                                                  | Label    | Description |
| --------- | ----------------------------------------------------- | -------- | ----------- |
| `address` | [string](#string)                                     |          |             |
| `amount`  | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |             |

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

| Field   | Type                                                  | Label    | Description |
| ------- | ----------------------------------------------------- | -------- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |             |

<a name="kava.kavadist.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest

QueryParamsRequest defines the request type for querying x/kavadist parameters.

<a name="kava.kavadist.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse

QueryParamsResponse defines the response type for querying x/kavadist parameters.

| Field    | Type                                    | Label | Description |
| -------- | --------------------------------------- | ----- | ----------- |
| `params` | [Params](#kava.kavadist.v1beta1.Params) |       |             |

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

<a name="kava.kavadist.v1beta1.Query"></a>

### Query

Query defines the gRPC querier service.

| Method Name | Request Type                                                      | Response Type                                                       | Description                                                    | HTTP Verb | Endpoint                          |
| ----------- | ----------------------------------------------------------------- | ------------------------------------------------------------------- | -------------------------------------------------------------- | --------- | --------------------------------- |
| `Params`    | [QueryParamsRequest](#kava.kavadist.v1beta1.QueryParamsRequest)   | [QueryParamsResponse](#kava.kavadist.v1beta1.QueryParamsResponse)   | Params queries the parameters of x/kavadist module.            | GET       | /kava/kavadist/v1beta1/parameters |
| `Balance`   | [QueryBalanceRequest](#kava.kavadist.v1beta1.QueryBalanceRequest) | [QueryBalanceResponse](#kava.kavadist.v1beta1.QueryBalanceResponse) | Balance queries the balance of all coins of x/kavadist module. | GET       | /kava/kavadist/v1beta1/balance    |

 <!-- end services -->

<a name="kava/pricefeed/v1beta1/pricefeed.proto"></a>

<p align="right"><a href="#top">Top</a></p>

## kava/pricefeed/v1beta1/pricefeed.proto

<a name="kava.pricefeed.v1beta1.CurrentPrice"></a>

### CurrentPrice

CurrentPrice defines a current price for a particular market in the pricefeed
module.

| Field       | Type              | Label | Description |
| ----------- | ----------------- | ----- | ----------- |
| `market_id` | [string](#string) |       |             |
| `price`     | [string](#string) |       |             |

<a name="kava.pricefeed.v1beta1.Market"></a>

### Market

Market defines an asset in the pricefeed.

| Field         | Type              | Label    | Description |
| ------------- | ----------------- | -------- | ----------- |
| `market_id`   | [string](#string) |          |             |
| `base_asset`  | [string](#string) |          |             |
| `quote_asset` | [string](#string) |          |             |
| `oracles`     | [string](#string) | repeated |             |
| `active`      | [bool](#bool)     |          |             |

<a name="kava.pricefeed.v1beta1.PostedPrice"></a>

### PostedPrice

PostedPrice defines a price for market posted by a specific oracle.

| Field            | Type                                                    | Label | Description |
| ---------------- | ------------------------------------------------------- | ----- | ----------- |
| `market_id`      | [string](#string)                                       |       |             |
| `oracle_address` | [string](#string)                                       |       |             |
| `price`          | [string](#string)                                       |       |             |
| `expiry`         | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |       |             |

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

| Field           | Type                                               | Label    | Description                                      |
| --------------- | -------------------------------------------------- | -------- | ------------------------------------------------ |
| `params`        | [Params](#kava.pricefeed.v1beta1.Params)           |          | params defines all the paramaters of the module. |
| `posted_prices` | [PostedPrice](#kava.pricefeed.v1beta1.PostedPrice) | repeated |                                                  |

<a name="kava.pricefeed.v1beta1.Params"></a>

### Params

Params defines the parameters for the pricefeed module.

| Field     | Type                                     | Label    | Description |
| --------- | ---------------------------------------- | -------- | ----------- |
| `markets` | [Market](#kava.pricefeed.v1beta1.Market) | repeated |             |

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

| Field     | Type                                     | Label    | Description     |
| --------- | ---------------------------------------- | -------- | --------------- |
| `markets` | [Market](#kava.pricefeed.v1beta1.Market) | repeated | List of markets |

<a name="kava.pricefeed.v1beta1.QueryOraclesRequest"></a>

### QueryOraclesRequest

QueryOraclesRequest is the request type for the Query/Oracles RPC method.

| Field       | Type              | Label | Description |
| ----------- | ----------------- | ----- | ----------- |
| `market_id` | [string](#string) |       |             |

<a name="kava.pricefeed.v1beta1.QueryOraclesResponse"></a>

### QueryOraclesResponse

QueryOraclesResponse is the response type for the Query/Oracles RPC method.

| Field     | Type              | Label    | Description              |
| --------- | ----------------- | -------- | ------------------------ |
| `oracles` | [string](#string) | repeated | List of oracle addresses |

<a name="kava.pricefeed.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest

QueryParamsRequest defines the request type for querying x/pricefeed
parameters.

<a name="kava.pricefeed.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse

QueryParamsResponse defines the response type for querying x/pricefeed
parameters.

| Field    | Type                                     | Label | Description |
| -------- | ---------------------------------------- | ----- | ----------- |
| `params` | [Params](#kava.pricefeed.v1beta1.Params) |       |             |

<a name="kava.pricefeed.v1beta1.QueryPriceRequest"></a>

### QueryPriceRequest

QueryPriceRequest is the request type for the Query/PriceRequest RPC method.

| Field       | Type              | Label | Description |
| ----------- | ----------------- | ----- | ----------- |
| `market_id` | [string](#string) |       |             |

<a name="kava.pricefeed.v1beta1.QueryPriceResponse"></a>

### QueryPriceResponse

QueryPriceResponse is the response type for the Query/Prices RPC method.

| Field   | Type                                                 | Label | Description |
| ------- | ---------------------------------------------------- | ----- | ----------- |
| `price` | [CurrentPrice](#kava.pricefeed.v1beta1.CurrentPrice) |       |             |

<a name="kava.pricefeed.v1beta1.QueryPricesRequest"></a>

### QueryPricesRequest

QueryPricesRequest is the request type for the Query/Prices RPC method.

<a name="kava.pricefeed.v1beta1.QueryPricesResponse"></a>

### QueryPricesResponse

QueryPricesResponse is the response type for the Query/Prices RPC method.

| Field    | Type                                                 | Label    | Description |
| -------- | ---------------------------------------------------- | -------- | ----------- |
| `prices` | [CurrentPrice](#kava.pricefeed.v1beta1.CurrentPrice) | repeated |             |

<a name="kava.pricefeed.v1beta1.QueryRawPricesRequest"></a>

### QueryRawPricesRequest

QueryRawPricesRequest is the request type for the Query/RawPrices RPC method.

| Field       | Type              | Label | Description |
| ----------- | ----------------- | ----- | ----------- |
| `market_id` | [string](#string) |       |             |

<a name="kava.pricefeed.v1beta1.QueryRawPricesResponse"></a>

### QueryRawPricesResponse

QueryRawPricesResponse is the response type for the Query/RawPrices RPC
method.

| Field        | Type                                               | Label    | Description |
| ------------ | -------------------------------------------------- | -------- | ----------- |
| `raw_prices` | [PostedPrice](#kava.pricefeed.v1beta1.PostedPrice) | repeated |             |

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

<a name="kava.pricefeed.v1beta1.Query"></a>

### Query

Query defines the gRPC querier service for pricefeed module

| Method Name | Request Type                                                           | Response Type                                                            | Description                                            | HTTP Verb | Endpoint                                      |
| ----------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------ | ------------------------------------------------------ | --------- | --------------------------------------------- |
| `Params`    | [QueryParamsRequest](#kava.pricefeed.v1beta1.QueryParamsRequest)       | [QueryParamsResponse](#kava.pricefeed.v1beta1.QueryParamsResponse)       | Params queries all parameters of the pricefeed module. | GET       | /kava/pricefeed/v1beta1/params                |
| `Price`     | [QueryPriceRequest](#kava.pricefeed.v1beta1.QueryPriceRequest)         | [QueryPriceResponse](#kava.pricefeed.v1beta1.QueryPriceResponse)         | Price queries price details based on a market          | GET       | /kava/pricefeed/v1beta1/prices/{market_id}    |
| `Prices`    | [QueryPricesRequest](#kava.pricefeed.v1beta1.QueryPricesRequest)       | [QueryPricesResponse](#kava.pricefeed.v1beta1.QueryPricesResponse)       | Prices queries all prices                              | GET       | /kava/pricefeed/v1beta1/prices                |
| `RawPrices` | [QueryRawPricesRequest](#kava.pricefeed.v1beta1.QueryRawPricesRequest) | [QueryRawPricesResponse](#kava.pricefeed.v1beta1.QueryRawPricesResponse) | RawPrices queries all raw prices based on a market     | GET       | /kava/pricefeed/v1beta1/rawprices/{market_id} |
| `Oracles`   | [QueryOraclesRequest](#kava.pricefeed.v1beta1.QueryOraclesRequest)     | [QueryOraclesResponse](#kava.pricefeed.v1beta1.QueryOraclesResponse)     | Oracles queries all oracles based on a market          | GET       | /kava/pricefeed/v1beta1/oracles/{market_id}   |
| `Markets`   | [QueryMarketsRequest](#kava.pricefeed.v1beta1.QueryMarketsRequest)     | [QueryMarketsResponse](#kava.pricefeed.v1beta1.QueryMarketsResponse)     | Markets queries all markets                            | GET       | /kava/pricefeed/v1beta1/markets               |

 <!-- end services -->

<a name="kava/pricefeed/v1beta1/tx.proto"></a>

<p align="right"><a href="#top">Top</a></p>

## kava/pricefeed/v1beta1/tx.proto

<a name="kava.pricefeed.v1beta1.MsgPostPrice"></a>

### MsgPostPrice

MsgPostPrice represents a method for creating a new post price

| Field       | Type                                                    | Label | Description       |
| ----------- | ------------------------------------------------------- | ----- | ----------------- |
| `from`      | [string](#string)                                       |       | address of client |
| `market_id` | [string](#string)                                       |       |                   |
| `price`     | [string](#string)                                       |       |                   |
| `expiry`    | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |       |                   |

<a name="kava.pricefeed.v1beta1.MsgPostPriceResponse"></a>

### MsgPostPriceResponse

MsgPostPriceResponse defines the Msg/PostPrice response type.

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

<a name="kava.pricefeed.v1beta1.Msg"></a>

### Msg

Msg defines the pricefeed Msg service.

| Method Name | Request Type                                         | Response Type                                                        | Description                                              | HTTP Verb | Endpoint |
| ----------- | ---------------------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------------------- | --------- | -------- |
| `PostPrice` | [MsgPostPrice](#kava.pricefeed.v1beta1.MsgPostPrice) | [MsgPostPriceResponse](#kava.pricefeed.v1beta1.MsgPostPriceResponse) | PostPrice defines a method for creating a new post price |           |

 <!-- end services -->

## Scalar Value Types

| .proto Type                    | Notes                                                                                                                                           | C++    | Java       | Python      | Go      | C#         | PHP            | Ruby                           |
| ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ---------- | ----------- | ------- | ---------- | -------------- | ------------------------------ |
| <a name="double" /> double     |                                                                                                                                                 | double | double     | float       | float64 | double     | float          | Float                          |
| <a name="float" /> float       |                                                                                                                                                 | float  | float      | float       | float32 | float      | float          | Float                          |
| <a name="int32" /> int32       | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32  | int        | int         | int32   | int        | integer        | Bignum or Fixnum (as required) |
| <a name="int64" /> int64       | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64  | long       | int/long    | int64   | long       | integer/string | Bignum                         |
| <a name="uint32" /> uint32     | Uses variable-length encoding.                                                                                                                  | uint32 | int        | int/long    | uint32  | uint       | integer        | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64     | Uses variable-length encoding.                                                                                                                  | uint64 | long       | int/long    | uint64  | ulong      | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32     | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s.                            | int32  | int        | int         | int32   | int        | integer        | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64     | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s.                            | int64  | long       | int/long    | int64   | long       | integer/string | Bignum                         |
| <a name="fixed32" /> fixed32   | Always four bytes. More efficient than uint32 if values are often greater than 2^28.                                                            | uint32 | int        | int         | uint32  | uint       | integer        | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64   | Always eight bytes. More efficient than uint64 if values are often greater than 2^56.                                                           | uint64 | long       | int/long    | uint64  | ulong      | integer/string | Bignum                         |
| <a name="sfixed32" /> sfixed32 | Always four bytes.                                                                                                                              | int32  | int        | int         | int32   | int        | integer        | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes.                                                                                                                             | int64  | long       | int/long    | int64   | long       | integer/string | Bignum                         |
| <a name="bool" /> bool         |                                                                                                                                                 | bool   | boolean    | boolean     | bool    | bool       | boolean        | TrueClass/FalseClass           |
| <a name="string" /> string     | A string must always contain UTF-8 encoded or 7-bit ASCII text.                                                                                 | string | String     | str/unicode | string  | string     | string         | String (UTF-8)                 |
| <a name="bytes" /> bytes       | May contain any arbitrary sequence of bytes.                                                                                                    | string | ByteString | str         | []byte  | ByteString | string         | String (ASCII-8BIT)            |
