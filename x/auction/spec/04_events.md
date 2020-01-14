# Events

The `x/auction` module emits the following events:

## Triggered By Other Modules

| Type          | Attribute Key | Attribute Value     |
|---------------|---------------|---------------------|
| auction_start | auction_id    | {auction ID}        |
| auction_start | auction_type  | {auction type}      |
| auction_start | lot_denom     | {auction lot denom} |
| auction_start | bid_denom     | {auction bid denom} |

## Handlers

### MsgPlaceBid

| Type        | Attribute Key | Attribute Value    |
|-------------|---------------|--------------------|
| auction_bid | auction_id    | {auction ID}       |
| auction_bid | bidder        | {latest bidder}    |
| auction_bid | bid_amount    | {coin amount}      |
| auction_bid | lot_amount    | {coin amount}      |
| auction_bid | end_time      | {auction end time} |
| message     | module        | auction            |
| message     | sender        | {sender address}   |

## EndBlock

| Type          | Attribute Key | Attribute Value |
|---------------|---------------|-----------------|
| auction_close | auction_id    | {auction ID}    |
