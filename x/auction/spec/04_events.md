<!--
order: 4
-->

# Events

The `x/auction` module emits the following events:

## Triggered By Other Modules

| Type          | Attribute Key | Attribute Value   |
|---------------|---------------|-------------------|
| auction_start | auction_id    | `{auction ID}`    |
| auction_start | auction_type  | `{auction type}`  |
| auction_start | lot           | `{coin amount}`   |
| auction_start | bid           | `{coin amount}`   |
| auction_start | max_bid       | `{coin amount}`   |

## Handlers

### MsgPlaceBid

| Type        | Attribute Key | Attribute Value      |
|-------------|---------------|----------------------|
| auction_bid | auction_id    | `{auction ID}`       |
| auction_bid | bidder        | `{latest bidder}`    |
| auction_bid | bid           | `{coin amount}`      |
| auction_bid | lot           | `{coin amount}`      |
| auction_bid | end_time      | `{auction end time}` |
| message     | module        | auction              |
| message     | sender        | `{sender address}`   |

## BeginBlock

| Type          | Attribute Key | Attribute Value   |
|---------------|---------------|-------------------|
| auction_close | auction_id    | `{auction ID}`    |
| auction_close | close_block   | `{block height}`  |
