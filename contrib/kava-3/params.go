// kava3 contains the suggested genesis parameters for the kava-3 mainnet.
package kava3

import (
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/kavadist"
	"github.com/kava-labs/kava/x/pricefeed"
)

const (
	kavaDenom      = "ukava"
	bnbDenom       = "bnb"
	usdxDenom      = "usdx"
	referenceAsset = "usd"
	bnbMarketID    = bnbDenom + ":" + referenceAsset // TODO is ':' safe in cdp rest urls?
	debtDenom      = "debt"
)

var testAddress = sdk.AccAddress("test address: len 20")

func AddSuggestedParams(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, chainID string, genesisTime time.Time) (tmtypes.GenesisDoc, error) {

	// Add tendermint params

	genDoc.ChainID = chainID
	genDoc.GenesisTime = genesisTime

	// Add app params

	var appState genutil.AppMap
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
		return tmtypes.GenesisDoc{}, err
	}

	previousBlockTime := time.Date(2020, time.June, 1, 12, 0, 0, 0, time.UTC) // TODO

	appState[auction.ModuleName] = cdc.MustMarshalJSON(auction.NewGenesisState(
		auction.DefaultNextAuctionID,
		auction.NewParams(
			48*time.Hour,
			6*time.Hour,
			sdk.MustNewDecFromStr("0.01"),
			sdk.MustNewDecFromStr("0.01"),
			sdk.MustNewDecFromStr("0.01"),
		),
		auction.GenesisAuctions{},
	))

	appState[bep3.ModuleName] = cdc.MustMarshalJSON(bep3.NewGenesisState(
		bep3.NewParams(
			testAddress, // TODO need deputy,
			0,
			bep3.DefaultMinBlockLock,
			bep3.DefaultMaxBlockLock,
			bep3.AssetParams{{
				Denom:  bnbDenom,
				CoinID: 714,
				Limit:  sdk.NewInt(100_000_000),
				Active: true,
			}},
		),
		bep3.AtomicSwaps{},
		bep3.AssetSupplies{}, // TODO should be populated?
	))

	appState[cdp.ModuleName] = cdc.MustMarshalJSON(cdp.NewGenesisState(
		cdp.NewParams(
			sdk.NewInt64Coin(usdxDenom, 100_000_000_000),
			cdp.CollateralParams{{
				Denom:              bnbDenom,
				LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
				DebtLimit:          sdk.NewInt64Coin(usdxDenom, 100_000_000_000),
				StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
				LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
				AuctionSize:        sdk.NewInt(100),
				Prefix:             0x20, // TODO ?
				ConversionFactor:   sdk.NewInt(8),
				MarketID:           bnbMarketID,
			}},
			cdp.DebtParam{
				Denom:            usdxDenom,
				ReferenceAsset:   referenceAsset,
				ConversionFactor: sdk.NewInt(6),
				DebtFloor:        sdk.NewInt(10_000_000),
				SavingsRate:      sdk.MustNewDecFromStr("0.9"),
			},
			sdk.NewInt(1000_000_000),
			sdk.NewInt(1000_000_000),
			24*time.Hour,
			false,
		),
		cdp.CDPs{},
		cdp.Deposits{},
		cdp.DefaultCdpStartingID,
		debtDenom,
		kavaDenom,
		genesisTime, // TODO this cannot be zero
	))

	appState[committee.ModuleName] = cdc.MustMarshalJSON(committee.NewGenesisState(
		committee.DefaultNextProposalID,
		[]committee.Committee{committee.NewCommittee(
			0,
			"This committee is for adjusting parameters of the cdp system.",
			[]sdk.AccAddress{testAddress}, // TODO add members
			[]committee.Permission{
				committee.SubParamChangePermission{
					AllowedParams: committee.AllowedParams{
						{
							Subspace: "auction",
							Key:      "BidDuration", // TODO snake_case ?
						},
						{
							Subspace: "auction",
							Key:      "IncrementSurplus",
						},
						{
							Subspace: "auction",
							Key:      "IncrementDebt",
						},
						{
							Subspace: "auction",
							Key:      "IncrementCollateral",
						},
						{
							Subspace: "bep3",
							Key:      "SupportedAssets",
						},
						{
							Subspace: "cdp",
							Key:      "GlobalDebtLimit",
						},
						{
							Subspace: "cdp",
							Key:      "SurplusThreshold",
						},
						{
							Subspace: "cdp",
							Key:      "DebtThreshold",
						},
						{
							Subspace: "cdp",
							Key:      "DistributionFrequency",
						},
						{
							Subspace: "cdp",
							Key:      "CircuitBreaker",
						},
						{
							Subspace: "cdp",
							Key:      "CollateralParams",
						},
						{
							Subspace: "cdp",
							Key:      "DebtParam",
						},
						{
							Subspace: "incentive",
							Key:      "Active",
						},
						{
							Subspace: "kavadist",
							Key:      "Active",
						},
						{
							Subspace: "pricefeed",
							Key:      "Markets",
						},
					},
					AllowedCollateralParams: committee.AllowedCollateralParams{{
						Denom:              bnbDenom,
						LiquidationRatio:   false,
						DebtLimit:          true,
						StabilityFee:       true,
						AuctionSize:        true,
						LiquidationPenalty: false,
						Prefix:             false,
						MarketID:           false,
						ConversionFactor:   false,
					}},
					AllowedDebtParam: committee.AllowedDebtParam{
						Denom:            false,
						ReferenceAsset:   false,
						ConversionFactor: false,
						DebtFloor:        false,
						SavingsRate:      true,
					},
					AllowedAssetParams: committee.AllowedAssetParams{{
						Denom:  bnbDenom,
						CoinID: false,
						Limit:  true,
						Active: true,
					}},
					AllowedMarkets: committee.AllowedMarkets{{
						MarketID:   bnbMarketID,
						BaseAsset:  false,
						QuoteAsset: false,
						Oracles:    false,
						Active:     true,
					}},
				},
				committee.TextPermission{},
			},
			sdk.MustNewDecFromStr("0.75"),
			7*24*time.Hour,
		),
			committee.NewCommittee(
				1,
				"emergency shutdown committee",
				[]sdk.AccAddress{testAddress}, // TODO
				[]committee.Permission{committee.SoftwareUpgradePermission{}},
				sdk.MustNewDecFromStr("0.75"),
				7*24*time.Hour,
			),
		},
		[]committee.Proposal{},
		[]committee.Vote{},
	))

	appState[incentive.ModuleName] = cdc.MustMarshalJSON(incentive.NewGenesisState(
		incentive.NewParams(
			true,
			incentive.Rewards{incentive.NewReward(
				true,
				kavaDenom,
				sdk.NewInt64Coin(kavaDenom, 100_000_000_000),
				2*7*24*time.Hour,
				2*365*24*time.Hour,
				2*7*24*time.Hour,
			)},
		),
		previousBlockTime,
		incentive.RewardPeriods{},
		incentive.ClaimPeriods{},
		incentive.Claims{},
		incentive.GenesisClaimPeriodIDs{},
	))

	appState[kavadist.ModuleName] = cdc.MustMarshalJSON(kavadist.NewGenesisState(
		kavadist.NewParams(
			true,
			kavadist.Periods{ // TODO what are our periods?
				{
					Start:     time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
					End:       time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
					Inflation: sdk.MustNewDecFromStr("1.00000001"),
				},
				{
					Start:     time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
					End:       time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC),
					Inflation: sdk.MustNewDecFromStr("1.00000001"),
				},
			},
		),
		previousBlockTime,
	))

	appState[pricefeed.ModuleName] = cdc.MustMarshalJSON(pricefeed.NewGenesisState(
		pricefeed.NewParams(
			pricefeed.Markets{{
				MarketID:   bnbMarketID,
				BaseAsset:  bnbDenom,
				QuoteAsset: referenceAsset,
				Oracles:    []sdk.AccAddress{}, // TODO need the oracles
				Active:     true,
			}},
		),
		pricefeed.PostedPrices{},
	))

	// TODO validator-vesting previous blockTime?

	// TODO sdk modules
	// crisis fee? minting.blocks_per_year mint inflation rate?

	marshaledAppState, err := cdc.MarshalJSON(appState)
	if err != nil {
		return tmtypes.GenesisDoc{}, err
	}
	genDoc.AppState = marshaledAppState

	return genDoc, nil
}
