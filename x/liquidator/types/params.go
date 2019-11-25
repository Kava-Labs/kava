package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Parameter keys
var (
	KeyDebtAuctionSize  = []byte("DebtAuctionSize")
	KeyCollateralParams = []byte("CollateralParams")
)

// LiquidatorParams store params for the liquidator module
type LiquidatorParams struct {
	DebtAuctionSize sdk.Int
	//SurplusAuctionSize sdk.Int
	CollateralParams []CollateralParams
}

// NewLiquidatorParams returns a new params object for the liquidator module
func NewLiquidatorParams(debtAuctionSize sdk.Int, collateralParams []CollateralParams) LiquidatorParams {
	return LiquidatorParams{
		DebtAuctionSize:  debtAuctionSize,
		CollateralParams: collateralParams,
	}
}

// String implements fmt.Stringer
func (p LiquidatorParams) String() string {
	out := fmt.Sprintf(`Params:
		Debt Auction Size: %s
		Collateral Params: `,
		p.DebtAuctionSize,
	)
	for _, cp := range p.CollateralParams {
		out += fmt.Sprintf(`
		%s`, cp.String())
	}
	return out
}

// CollateralParams params storing information about each collateral for the liquidator module
type CollateralParams struct {
	Denom       string  // Coin name of collateral type
	AuctionSize sdk.Int // Max amount of collateral to sell off in any one auction. Known as lump in Maker.
	// LiquidationPenalty
}

// String implements stringer interface
func (cp CollateralParams) String() string {
	return fmt.Sprintf(`
  Denom:        %s
  AuctionSize: %s`, cp.Denom, cp.AuctionSize)
}

// ParamKeyTable for the liquidator module
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&LiquidatorParams{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of liquidator module's parameters.
// nolint
func (p *LiquidatorParams) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		subspace.NewParamSetPair(KeyDebtAuctionSize, &p.DebtAuctionSize),
		subspace.NewParamSetPair(KeyCollateralParams, &p.CollateralParams),
	}
}

// DefaultParams for the liquidator module
func DefaultParams() LiquidatorParams {
	return LiquidatorParams{
		DebtAuctionSize:  sdk.NewInt(1000),
		CollateralParams: []CollateralParams{},
	}
}

func (p LiquidatorParams) Validate() error {
	if p.DebtAuctionSize.IsNegative() {
		return fmt.Errorf("debt auction size should be positive, is %s", p.DebtAuctionSize)
	}
	denomDupMap := make(map[string]int)
	for _, cp := range p.CollateralParams {
		_, found := denomDupMap[cp.Denom]
		if found {
			return fmt.Errorf("duplicate denom: %s", cp.Denom)
		}
		denomDupMap[cp.Denom] = 1
		if cp.AuctionSize.IsNegative() {
			return fmt.Errorf(
				"auction size for each collateral should be positive, is %s for %s", cp.AuctionSize, cp.Denom,
			)
		}
	}
	return nil
}
