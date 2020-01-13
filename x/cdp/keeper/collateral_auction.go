package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

type partialDeposit struct {
	types.Deposit

	DebtShare sdk.Int
}

type partialDeposits []partialDeposit

func (pd partialDeposits) SumCollateral() (sum sdk.Int) {
	sum = sdk.ZeroInt()
	for _, d := range pd {
		sum = sum.Add(d.Amount[0].Amount)
	}
	return
}

func (pd partialDeposits) SumDebt() (sum sdk.Int) {
	sum = sdk.ZeroInt()
	for _, d := range pd {
		sum = sum.Add(d.DebtShare)
	}
	return
}

func (k Keeper) AuctionCollateral(ctx sdk.Context, deposits types.Deposits, debt sdk.Int, bidDenom string) {
	auctionSize := k.getAuctionSize(ctx, deposits[0].Amount[0].Denom)
	partialAuctionDeposits := partialDeposits{}
	totalCollateral := deposits.SumCollateral()
	for deposits.SumCollateral().GTE(auctionSize) {
		for i, dep := range deposits {
			// create auctions from individual deposits that are larger than the auction size
			k.CreateAuctionsFromDeposit(ctx, &dep, &debt, &totalCollateral, auctionSize, bidDenom)
			debtCoveredByDeposit := (dep.Amount[0].Amount.Quo(totalCollateral)).Mul(debt)
			if !dep.Amount[0].Amount.IsZero() {
				if (partialAuctionDeposits.SumCollateral().Add(dep.Amount[0].Amount)).LT(auctionSize) {
					partialAuctionDeposits = append(partialAuctionDeposits, partialDeposit{dep, debtCoveredByDeposit})
					dep.Amount[0].Amount = sdk.ZeroInt()
				} else {
					partialCollateral := auctionSize.Sub(partialAuctionDeposits.SumCollateral())
					partialDebt := (partialCollateral.Quo(dep.Amount[0].Amount)).Mul(debtCoveredByDeposit)

					partialDep := partialDeposit{
						Deposit: types.Deposit{
							CdpID:     dep.CdpID,
							Depositor: dep.Depositor,
							Amount:    sdk.NewCoins(sdk.NewCoin(dep.Amount[0].Denom, partialCollateral)),
						},
						DebtShare: partialDebt,
					}
					partialAuctionDeposits = append(partialAuctionDeposits, partialDep)
					k.CreateAuctionFromPartialDeposits(ctx, partialAuctionDeposits, &debt, &totalCollateral, auctionSize, bidDenom)
					partialAuctionDeposits = partialDeposits{}
				}
			}
			if dep.Amount[0].Amount.IsZero() {
				// remove the deposit from the slice if it is empty
				deposits = append(deposits[:i], deposits[i+1:]...)
				i--
			} else {
				deposits[i] = dep
			}
		}
	}
	if partialAuctionDeposits.SumCollateral().GT(sdk.ZeroInt()) {
		k.CreateAuctionFromPartialDeposits(ctx, partialAuctionDeposits, &debt, &totalCollateral, partialAuctionDeposits.SumCollateral(), bidDenom)
	}
}

// deposit -1000xrp
// auction size 500xrp
// total collateral 2000xrp
// total debt = 2000debt
// debt accounted for by one auction from deposit:
// (500xrp/2000xrp) * 2000debt = 500debt

// CreateAuctionsFromDeposit creates auctions from the input deposit until there is less than auctionSize left on the deposit
func (k Keeper) CreateAuctionsFromDeposit(ctx sdk.Context, dep *types.Deposit, debt *sdk.Int, totalCollateral *sdk.Int, auctionSize sdk.Int, principalDenom string) {
	for dep.Amount[0].Amount.GTE(auctionSize) {
		// figure out how much debt is covered by one lots worth of collateral
		depositDebtAmount := (auctionSize.Quo(*totalCollateral)).Mul(*debt)
		// start an auction for one lot, attempting to raise depositDebtAmount
		_, err := k.auctionKeeper.StartForwardReverseAuction(
			ctx, types.LiquidatorMacc, sdk.NewCoin(dep.Amount[0].Denom, auctionSize),
			sdk.NewCoin(principalDenom, depositDebtAmount), []sdk.AccAddress{dep.Depositor},
			[]sdk.Int{auctionSize})
		if err != nil {
			panic(err)
		}
		// subtract one lot's worth of debt from the total debt covered by this deposit
		*debt = debt.Sub(depositDebtAmount)
		// update the deposits collateral amount
		dep.Amount[0].Amount = dep.Amount[0].Amount.Sub(auctionSize)
		// update the total collateral
		*totalCollateral = totalCollateral.Sub(auctionSize)
	}
}

// totalDebt 1000
//

func (k Keeper) CreateAuctionFromPartialDeposits(ctx sdk.Context, partialDeps partialDeposits, debt *sdk.Int, collateral *sdk.Int, auctionSize sdk.Int, bidDenom string) {
	returnAddrs := []sdk.AccAddress{}
	returnWeights := []sdk.Int{}
	for _, pd := range partialDeps {
		returnAddrs = append(returnAddrs, pd.Depositor)
		returnWeights = append(returnWeights, pd.DebtShare)
	}
	_, err := k.auctionKeeper.StartForwardReverseAuction(ctx, types.LiquidatorMacc, sdk.NewCoin(partialDeps[0].Amount[0].Denom, auctionSize), sdk.NewCoin(bidDenom, partialDeps.SumDebt()), returnAddrs, returnWeights)
	if err != nil {
		panic(err)
	}
	*debt = debt.Sub(partialDeps.SumDebt())
	*collateral = collateral.Sub(partialDeps.SumCollateral())
}
