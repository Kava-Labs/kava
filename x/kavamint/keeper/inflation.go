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

// Minter wraps the logic of a single source of inflation. It calculates the amount of coins to be
// minted to match a yearly `rate` of inflating the `basis`.
type Minter struct {
	name         string  // name of the minting. used as event attribute to report how much is minted
	rate         sdk.Dec // the yearly apy of the mint. ex. 20% => 0.2
	basis        sdk.Int // the base amount of coins that is inflated
	destMaccName string  // the destination module account name to mint coins to
	mintDenom    string  // denom of coin this minter is responsible for minting
}

// NewMinter returns creates a new source of minting inflation
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

// AccumulateInflation calculates the number of coins that should be minted to match a yearly `rate`
// for interest compounded each second of the year over `secondsPassed` seconds.
// `basis` is the base amount of coins that is inflated.
func (m Minter) AccumulateInflation(secondsPassed uint64) (sdk.Coin, error) {
	// calculate the rate factor based on apy & seconds passed since last block
	inflationRate, err := m.CalculateInflationRate(secondsPassed)
	if err != nil {
		return sdk.Coin{}, err
	}
	amount := inflationRate.MulInt(m.basis).TruncateInt()
	return sdk.NewCoin(m.mintDenom, amount), nil
}

// CalculateInflationRate converts an APY into the factor corresponding with that APY's accumulation
// over a period of secondsPassed seconds.
func (m Minter) CalculateInflationRate(secondsPassed uint64) (sdk.Dec, error) {
	perSecondInterestRate, err := apyToSpy(m.rate.Add(sdk.OneDec()))
	if err != nil {
		return sdk.ZeroDec(), err
	}
	rate := perSecondInterestRate.Power(secondsPassed)
	return rate.Sub(sdk.OneDec()), nil
}

// AccumulateAndMintInflation defines the sources of inflation, determines the seconds passed since
// the last mint, and then mints each source of inflation to the defined destination.
func (k Keeper) AccumulateAndMintInflation(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	// determine seconds since last mint
	previousBlockTime := k.GetPreviousBlockTime(ctx)
	if previousBlockTime.IsZero() {
		previousBlockTime = ctx.BlockTime()
	}
	secondsPassed := ctx.BlockTime().Sub(previousBlockTime).Seconds()

	// calculate totals before any minting is done to prevent new mints affecting the values
	totalSupply := k.totalSupply(ctx)
	totalBonded := k.totalBondedTokens(ctx)

	bondDenom := k.bondDenom(ctx)

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

		// mint & transfer coins
		inflation := sdk.NewCoins(amount)
		if err := k.mintCoinsToModule(ctx, inflation, minter.destMaccName); err != nil {
			return err
		}

		// add attribute to event
		mintEventAttrs = append(mintEventAttrs, sdk.NewAttribute(
			minter.name,
			inflation.String(),
		))
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
