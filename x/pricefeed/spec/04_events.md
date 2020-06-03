<!--
order: 4
-->

# Events

The `x/pricefeed` module emits the following events:

## MsgPostPrice

| Type                 | Attribute Key | Attribute Value    |
|----------------------|---------------|--------------------|
| oracle_updated_price | market_id     | `{market ID}`      |
| oracle_updated_price | oracle        | `{oracle}`         |
| oracle_updated_price | market_price  | `{price}`          |
| oracle_updated_price | expiry        | `{expiry}`         |
| message              | module        | pricefeed          |
| message              | sender        | `{sender address}` |

## BeginBlock

| Type                 | Attribute Key   | Attribute Value  |
|----------------------|-----------------|------------------|
| market_price_updated | market_id       | `{market ID}`    |
| market_price_updated | market_price    | `{price}`        |
| no_valid_prices      | market_id       | `{market ID}`    |
