# State Transitions

State changes when:

- msg submitted
- end blocker runs
- keeper methods called by another module (usually liquidator)

Msgs and end blocker state changes are detailed in their own files.

## State Change Through External Keeper Method Calls

This module exposes this interface to a liquidator module:

```go
type CdpKeeper interface {
    // non state changing methods
    GetCDP(ctx sdk.Context, collateralDenom string, cdpID uint64) (cdptypes.CDP, bool)
    GetAllLiquidatedDeposits(ctx sdk.Context) (deposits cdptypes.Deposits)
    GetDebtDenom(ctx sdk.Context) (denom string)
    GetDeposits(ctx sdk.Context, cdpID uint64) (deposits cdptypes.Deposits)
    GetCdpByOwnerAndDenom(ctx sdk.Context, owner sdk.AccAddress, denom string) (cdptypes.CDP, bool)
    // state changing methods
    DeleteCDP(ctx sdk.Context, cdp cdptypes.CDP)
    RemoveCdpOwnerIndex(ctx sdk.Context, cdp cdptypes.CDP)
    DeleteDeposit(ctx sdk.Context, status cdptypes.DepositStatus, cdpID uint64, depositor sdk.AccAddress)
}
```

All it can do is delete CDPs and Deposits.
