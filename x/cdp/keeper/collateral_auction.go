package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

func (k Keeper) AuctionCollateral(ctx sdk.Context, deposits types.Deposits, debt sdk.Int, auctionDenom string) {
	auctionSize := k.getAuctionSize(ctx, deposits[0].Amount[0].Denom)
	for i, dep := range deposits {
		// create auctions from individual deposits that are larger than the auction size
		k.CreateAuctionsFromDeposit(ctx, &dep, &debt, auctionSize, auctionDenom)
		if dep.Amount[0].Amount.IsZero() {
			// remove the deposit from the slice if it is empty
			deposits = append(deposits[:i], deposits[i+1:]...)
		}
	}
	k.CreateAuctionsFromDeposits(ctx, deposits)

}

// CreateAuctionsFromDeposit creates auctions from the input deposit until there is less than auctionSize left on the deposit
func (k Keeper) CreateAuctionsFromDeposit(ctx sdk.Context, dep *types.Deposit, debt *sdk.Int, auctionSize sdk.Int, principalDenom string) {
	for dep.Amount[0].Amount.GTE(auctionSize) {
		// figure out how much debt is covered by one lots worth of collateral
		depositDebtAmount := (auctionSize.Quo(dep.Amount[0].Amount)).Mul(*debt)
		// subtract one lot's worth of debt from the total debt covered by this deposit
		*debt = debt.Sub(depositDebtAmount)
		// start an auction for one lot, attempting to raise depositDebtAmount
		_, err := k.auctionKeeper.StartForwardReverseAuction(
			ctx, types.ModuleName, sdk.NewCoin(dep.Amount[0].Denom, auctionSize),
			sdk.NewCoin(principalDenom, depositDebtAmount), []sdk.AccAddress{dep.Depositor},
			[]sdk.Int{auctionSize})
		if err != nil {
			panic(err)
		}
		dep.Amount[0].Amount = dep.Amount[0].Amount.Sub(auctionSize)
	}
}
