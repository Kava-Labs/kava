<!--
order: 4
-->

# Events

The `x/issuance` module emits the following events:

## BeginBlock

| Type                 | Attribute Key       | Attribute Value |
|----------------------|---------------------|-----------------|
| issuance             | new_issuance        | `{amount}`      |
| issuance             | redemption          | `{amount}`      |
| issuance             | pause               | `{denom}`       |
