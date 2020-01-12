# Events

The `x/auction` module emits the following events:

## EndBlock

| Type          | Attribute Key | Attribute Value |
| ------------- | ------------- | --------------- |
| auction_close | auction_id    | {auction ID}    |

## Handlers

### MsgPlaceBid

| Type        | Attribute Key | Attribute Value     |
| ----------- | ------------- | ------------------- |
| auction_bid | auction_id    | {auction ID}        |
| message     | module        | auction             |
| message     | sender        | {sender address}    |

## Triggered By Other Modules

| Type          | Attribute Key | Attribute Value     |
| ------------- | ------------- | ------------------- |
| auction_start | auction_id    | {auction ID}        |
| auction_start | auction_type  | {auction type}      |
| auction_start | lot_denom     | {auction lot denom} |
