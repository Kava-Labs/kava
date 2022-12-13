package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavamint/types"
)

// this is the same value used in the x/hard
const (
	SecondsPerYear = uint64(31536000)
)

type Minter struct {
	name         string
	rate         sdk.Dec
	basis        sdk.Int
	destMaccName string
	mintDenom    string
}

func NewMinter(
	name string, rate sdk.Dec, basis sdk.Int, mintDenom string, destMaccName string,
) Minter {
	return Minter{
		name:         name,
		rate:         rate,
		basis:        basis,
		destMaccName: destMaccName,
		mintDenom:    mintDenom,
	}
}

func (m Minter) AccumulateInflation(secondsPassed uint64) (sdk.Coin, error) {
	// calculate the rate factor based on apy & seconds passed since last block
	inflationRate, err := m.CalculateInflationRate(secondsPassed)
	if err != nil {
		return sdk.Coin{}, err
	}
	amount := inflationRate.MulInt(m.basis).TruncateInt()
	return sdk.NewCoin(m.mintDenom, amount), nil
}

func (m Minter) CalculateInflationRate(secondsPassed uint64) (sdk.Dec, error) {
	perSecondInterestRate, err := apyToSpy(m.rate.Add(sdk.OneDec()))
	if err != nil {
		return sdk.ZeroDec(), err
	}
	rate := perSecondInterestRate.Power(secondsPassed)
	return rate.Sub(sdk.OneDec()), nil
}

func (k Keeper) AccumulateAndMintInflation(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	// determine seconds since last mint
	previousBlockTime := k.GetPreviousBlockTime(ctx)
	if previousBlockTime.IsZero() {
		previousBlockTime = ctx.BlockTime()
	}
	secondsPassed := ctx.BlockTime().Sub(previousBlockTime).Seconds()

	// calculate totals before any minting is done to prevent new mints affecting the values
	totalSupply := k.TotalSupply(ctx)
	totalBonded := k.TotalBondedTokens(ctx)

	bondDenom := k.BondDenom(ctx)

	// define minters for each contributing piece of inflation
	minters := []Minter{
		// staking rewards
		NewMinter(
			types.AttributeKeyStakingRewardMint,
			params.StakingRewardsApy,
			totalBonded,
			bondDenom,
			k.stakingRewardsFeeCollectorName,
		),
		// community pool inflation
		NewMinter(
			types.AttributeKeyCommunityPoolMint,
			params.CommunityPoolInflation,
			totalSupply,
			bondDenom,
			k.communityPoolModuleAccountName,
		),
	}

	mintEventAttrs := make([]sdk.Attribute, 0, len(minters))
	for _, minter := range minters {
		// calculate amount for time passed
		amount, err := minter.AccumulateInflation(uint64(secondsPassed))
		if err != nil {
			return err
		}

		inflation := sdk.NewCoins(amount)
		mintEventAttrs = append(mintEventAttrs, sdk.NewAttribute(
			minter.name,
			inflation.String(),
		))

		if inflation.IsZero() {
			continue
		}

		// mint the coins
		if err := k.MintCoins(ctx, inflation); err != nil {
			return err
		}
		// transfer to destination module
		err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, minter.destMaccName, inflation)
		if err != nil {
			return err
		}
	}

	// ------------- Bookkeeping -------------
	// bookkeep the previous block time
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyTotalSupply, totalSupply.String()),
			sdk.NewAttribute(types.AttributeKeyTotalBonded, totalBonded.String()),
			sdk.NewAttribute(types.AttributeSecondsPassed, fmt.Sprintf("%f", secondsPassed)),
		).AppendAttributes(mintEventAttrs...),
	)

	return nil
}

// AccumulateInflation calculates the number of coins that should be minted to match a yearly `rate`
// for interest compounded each second of the year over `secondsSinceLastMint` seconds.
// `basis` is the base amount of coins that is inflated.
func (k Keeper) AccumulateInflation(
	ctx sdk.Context,
	rate sdk.Dec,
	basis sdk.Int,
	secondsSinceLastMint float64,
) (sdk.Coins, error) {
	bondDenom := k.BondDenom(ctx)

	// calculate the rate factor based on apy & seconds passed since last block
	inflationRate, err := CalculateInflationRate(rate, uint64(secondsSinceLastMint))
	if err != nil {
		return sdk.NewCoins(), err
	}

	amount := inflationRate.MulInt(basis).TruncateInt()

	return sdk.NewCoins(sdk.NewCoin(bondDenom, amount)), nil
}

// CalculateInflationRate converts an APY into the factor corresponding with that APY's accumulation
// over a period of secondsPassed seconds.
func CalculateInflationRate(apy sdk.Dec, secondsPassed uint64) (sdk.Dec, error) {
	perSecondInterestRate, err := apyToSpy(apy.Add(sdk.OneDec()))
	if err != nil {
		return sdk.ZeroDec(), err
	}
	rate := perSecondInterestRate.Power(secondsPassed)
	return rate.Sub(sdk.OneDec()), nil
}

// apyToSpy converts the input annual interest rate. For example, 10% apy would be passed as 1.10.
// SPY = Per second compounded interest rate is how cosmos mathematically represents APY.
func apyToSpy(apy sdk.Dec) (sdk.Dec, error) {
	// Note: any APY greater than 176.5 will cause an out-of-bounds error
	root, err := apy.ApproxRoot(SecondsPerYear)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return root, nil
}
