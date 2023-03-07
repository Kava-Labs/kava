<!--
order: 4
-->

# Events

The `x/liquid` module emits the following events:

## MsgMintDerivative

| Type            | Attribute Key     | Attribute Value       |
| --------------- | ----------------- | ------------------    |
| mint_derivative | delegator         | `{delegator address}` |
| mint_derivative | validator         | `{validator address}` |
| mint_derivative | amount            | `{amount}`            |
| mint_derivative | shares_transferred| `{shares transferred}`|

## MsgBurnDerivative

| Type            | Attribute Key     | Attribute Value       |
| --------------- | ----------------- | ------------------    |
| burn_derivative | delegator         | `{delegator address}` |
| burn_derivative | validator         | `{validator address}` |
| burn_derivative | amount            | `{amount}`            |
| burn_derivative | shares_transferred| `{shares transferred}`|