<!--
order: 4
-->

# Events

The `x/bep3` module emits the following events:

## Handlers

### MsgCreateAtomicSwap

| Type               | Attribute Key      | Attribute Value           |
|--------------------|--------------------|---------------------------|
| create_atomic_swap | sender             | `{sender address}`        |
| create_atomic_swap | recipient          | `{recipient address}`     |
| create_atomic_swap | atomic_swap_id     | `{swap ID}`               |
| create_atomic_swap | random_number_hash | `{random number hash}`    |
| create_atomic_swap | timestamp          | `{timestamp}`             |
| create_atomic_swap | sender_other_chain | `{sender other chain}`    |
| create_atomic_swap | expire_height      | `{swap expiration block}` |
| create_atomic_swap | amount             | `{coin amount}`           |
| create_atomic_swap | direction          | `{incoming or outgoing}`  |
| message            | module             | bep3                      |
| message            | sender             | `{sender address}`        |

### MsgClaimAtomicSwap

| Type               | Attribute Key      | Attribute Value           |
|--------------------|--------------------|---------------------------|
| claim_atomic_swap  | claim_sender       | `{sender address}`        |
| claim_atomic_swap  | recipient          | `{recipient address}`     |
| claim_atomic_swap  | atomic_swap_id     | `{swap ID}`               |
| claim_atomic_swap  | random_number_hash | `{random number hash}`    |
| claim_atomic_swap  | random_number      | `{secret random number}`  |
| message            | module             | bep3                      |
| message            | sender             | `{sender address}`        |

## MsgRefundAtomicSwap

| Type               | Attribute Key      | Attribute Value           |
|--------------------|--------------------|---------------------------|
| refund_atomic_swap | refund_sender      | `{sender address}`        |
| refund_atomic_swap | sender             | `{swap creator address}`  |
| refund_atomic_swap | atomic_swap_id     | `{swap ID}`               |
| refund_atomic_swap | random_number_hash | `{random number hash}`    |
| message            | module             | bep3                      |
| message            | sender             | `{sender address}`        |

## BeginBlock

| Type          | Attribute Key    | Attribute Value                  |
|---------------|------------------|----------------------------------|
| swaps_expired | atomic_swap_ids  | `{array of swap IDs}`            |
| swaps_expired | expiration_block | `{block height at expiration}`   |
