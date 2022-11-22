<!--
order: 3
-->

# Messages

## FundCommunityPool

This message sends coins directly from the sender to the community module account.

The transaction fails if the amount cannot be transferred from the sender to the community module account.

```go
func (k Keeper) FundCommunityPool(ctx sdk.Context, sender sdk.AccAddress, amount sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleAccountName, amount)
}
```
