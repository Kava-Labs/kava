// kava3 contains the suggested genesis parameters for the kava-3 mainnet.
package kava3

import (
	"fmt"
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
	kavaDenom              = "ukava"
	bnbDenom               = "bnb"
	usdxDenom              = "usdx"
	referenceAsset         = "usd"
	bnbSpotMarketID        = bnbDenom + ":" + referenceAsset
	bnbLiquidationMarketID = bnbDenom + ":" + referenceAsset + ":" + "30"
	debtDenom              = "debt"
)

func AddSuggestedParams(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, chainID string, genesisTime time.Time) (tmtypes.GenesisDoc, error) {

	// Add tendermint params

	genDoc.ChainID = chainID
	genDoc.GenesisTime = genesisTime

	// Add app params

	var appState genutil.AppMap
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
		return tmtypes.GenesisDoc{}, err
	}

	addAuctionState(cdc, appState)
	addBep3State(cdc, appState)
	addCDPState(cdc, appState)
	addCommitteeState(cdc, appState)
	addIncentiveState(cdc, appState)
	addKavaDistState(cdc, appState)
	addPricefeedState(cdc, appState)

	marshaledAppState, err := cdc.MarshalJSON(appState)
	if err != nil {
		return tmtypes.GenesisDoc{}, err
	}
	genDoc.AppState = marshaledAppState

	return genDoc, nil
}

func addAuctionState(cdc *codec.Codec, appState genutil.AppMap) {
	appState[auction.ModuleName] = cdc.MustMarshalJSON(auction.NewGenesisState(
		auction.DefaultNextAuctionID,
		auction.NewParams(
			24*time.Hour,
			8*time.Hour,
			sdk.MustNewDecFromStr("0.01"),
			sdk.MustNewDecFromStr("0.01"),
			sdk.MustNewDecFromStr("0.01"),
		),
		auction.GenesisAuctions{},
	))
}

func addBep3State(cdc *codec.Codec, appState genutil.AppMap) {
	appState[bep3.ModuleName] = cdc.MustMarshalJSON(bep3.NewGenesisState(
		bep3.NewParams(
			sdk.AccAddress("address for a deputy"), // TODO pending receipt of deputy address
			1000,
			bep3.DefaultMinBlockLock,
			bep3.DefaultMaxBlockLock,
			bep3.AssetParams{{
				Denom:  bnbDenom,
				CoinID: 714,
				Limit:  sdk.NewInt(4_000_000_000_000),
				Active: true,
			}},
		),
		bep3.AtomicSwaps{},
		bep3.AssetSupplies{},
	))
}

func addCDPState(cdc *codec.Codec, appState genutil.AppMap) {
	appState[cdp.ModuleName] = cdc.MustMarshalJSON(cdp.NewGenesisState(
		cdp.NewParams(
			sdk.NewInt64Coin(usdxDenom, 100_000_000_000),
			cdp.CollateralParams{{
				Denom:               bnbDenom,
				LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
				DebtLimit:           sdk.NewInt64Coin(usdxDenom, 100_000_000_000),
				StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
				LiquidationPenalty:  sdk.MustNewDecFromStr("0.075"),
				AuctionSize:         sdk.NewInt(50_000_000_000),
				Prefix:              0x20,
				ConversionFactor:    sdk.NewInt(8),
				SpotMarketID:        bnbSpotMarketID,
				LiquidationMarketID: bnbLiquidationMarketID,
			}},
			cdp.DebtParam{
				Denom:            usdxDenom,
				ReferenceAsset:   referenceAsset,
				ConversionFactor: sdk.NewInt(6),
				DebtFloor:        sdk.NewInt(10_000_000),
				SavingsRate:      sdk.MustNewDecFromStr("0.9"),
			},
			sdk.NewInt(20_000_000_000),
			sdk.NewInt(10_000_000_000),
			24*time.Hour,
			false,
		),
		cdp.CDPs{},
		cdp.Deposits{},
		cdp.DefaultCdpStartingID,
		debtDenom,
		kavaDenom,
		cdp.DefaultPreviousDistributionTime,
	))
}

func addCommitteeState(cdc *codec.Codec, appState genutil.AppMap) {
	appState[committee.ModuleName] = cdc.MustMarshalJSON(committee.NewGenesisState(
		committee.DefaultNextProposalID,
		[]committee.Committee{
			committee.NewCommittee(
				1,
				"Kava Stability Committee",
				[]sdk.AccAddress{
					// addresses from governance proposal: https://ipfs.io/ipfs/QmSiQexKNixztPgLCe2cRSJ8ZLRjetRgzHPDTuBRCm9DZb/committee-nominations.pdf
					mustAccAddressFromBech32("kava1gru35up50ql2wxhegr880qy6ynl63ujlv8gum2"),
					mustAccAddressFromBech32("kava1sc3mh3pkas5e7xd269am4xm5mp6zweyzmhjagj"),
					mustAccAddressFromBech32("kava1c9ye54e3pzwm3e0zpdlel6pnavrj9qqv6e8r4h"),
					mustAccAddressFromBech32("kava1m7p6sjqrz6mylz776ct48wj6lpnpcd0z82209d"),
					mustAccAddressFromBech32("kava1a9pmkzk570egv3sflu3uwdf3gejl7qfy9hghzl"),
				},
				[]committee.Permission{
					committee.SubParamChangePermission{
						AllowedParams: committee.AllowedParams{
							{
								Subspace: auction.ModuleName,
								Key:      string(auction.KeyBidDuration),
							},
							{
								Subspace: auction.ModuleName,
								Key:      string(auction.KeyIncrementSurplus),
							},
							{
								Subspace: auction.ModuleName,
								Key:      string(auction.KeyIncrementDebt),
							},
							{
								Subspace: auction.ModuleName,
								Key:      string(auction.KeyIncrementCollateral),
							},
							{
								Subspace: bep3.ModuleName,
								Key:      string(bep3.KeySupportedAssets),
							},
							{
								Subspace: cdp.ModuleName,
								Key:      string(cdp.KeyGlobalDebtLimit),
							},
							{
								Subspace: cdp.ModuleName,
								Key:      string(cdp.KeySurplusThreshold),
							},
							{
								Subspace: cdp.ModuleName,
								Key:      string(cdp.KeyDebtThreshold),
							},
							{
								Subspace: cdp.ModuleName,
								Key:      string(cdp.KeyDistributionFrequency),
							},
							{
								Subspace: cdp.ModuleName,
								Key:      string(cdp.KeyCollateralParams),
							},
							{
								Subspace: cdp.ModuleName,
								Key:      string(cdp.KeyDebtParam),
							},
							{
								Subspace: incentive.ModuleName,
								Key:      string(incentive.KeyActive),
							},
							{
								Subspace: kavadist.ModuleName,
								Key:      string(kavadist.KeyActive),
							},
							{
								Subspace: pricefeed.ModuleName,
								Key:      string(pricefeed.KeyMarkets),
							},
						},
						AllowedCollateralParams: committee.AllowedCollateralParams{{
							Denom:               bnbDenom,
							LiquidationRatio:    false,
							DebtLimit:           true,
							StabilityFee:        true,
							AuctionSize:         true,
							LiquidationPenalty:  false,
							Prefix:              false,
							SpotMarketID:        false,
							LiquidationMarketID: false,
							ConversionFactor:    false,
						}},
						AllowedDebtParam: committee.AllowedDebtParam{
							Denom:            false,
							ReferenceAsset:   false,
							ConversionFactor: false,
							DebtFloor:        true,
							SavingsRate:      true,
						},
						AllowedAssetParams: committee.AllowedAssetParams{{
							Denom:  bnbDenom,
							CoinID: false,
							Limit:  true,
							Active: true,
						}},
						AllowedMarkets: committee.AllowedMarkets{
							{
								MarketID:   bnbSpotMarketID,
								BaseAsset:  false,
								QuoteAsset: false,
								Oracles:    false,
								Active:     true,
							},
							{
								MarketID:   bnbLiquidationMarketID,
								BaseAsset:  false,
								QuoteAsset: false,
								Oracles:    false,
								Active:     true,
							},
						},
					},
					committee.TextPermission{},
				},
				sdk.MustNewDecFromStr("0.5"), // 3 of 5
				7*24*time.Hour,
			),
			committee.NewCommittee(
				2,
				"Kava Safety Committee",
				[]sdk.AccAddress{
					// address from governance proposal: https://ipfs.io/ipfs/QmPqfP1Fa8EyzubmctL5uT5TAcWTB7HBQd8pvrmSTG8yS1/safety-nominations.pdf
					mustAccAddressFromBech32("kava1e0agyg6eug9r62fly9sls77ycjgw8ax6xk73es"),
				},
				[]committee.Permission{committee.SoftwareUpgradePermission{}},
				sdk.MustNewDecFromStr("0.5"),
				7*24*time.Hour,
			),
		},
		[]committee.Proposal{},
		[]committee.Vote{},
	))
}

func addIncentiveState(cdc *codec.Codec, appState genutil.AppMap) {
	appState[incentive.ModuleName] = cdc.MustMarshalJSON(incentive.NewGenesisState(
		incentive.NewParams(
			true,
			incentive.Rewards{incentive.NewReward(
				false,
				kavaDenom,
				sdk.NewInt64Coin(kavaDenom, 74_000_000_000),
				1*7*24*time.Hour,
				1*365*24*time.Hour,
				1*7*24*time.Hour,
			)},
		),
		incentive.DefaultPreviousBlockTime,
		incentive.RewardPeriods{},
		incentive.ClaimPeriods{},
		incentive.Claims{},
		incentive.GenesisClaimPeriodIDs{},
	))
}

func addKavaDistState(cdc *codec.Codec, appState genutil.AppMap) {
	appState[kavadist.ModuleName] = cdc.MustMarshalJSON(kavadist.NewGenesisState(
		kavadist.NewParams(
			true,
			kavadist.Periods{
				{
					Start:     time.Date(2020, 6, 1, 14, 0, 0, 0, time.UTC),
					End:       time.Date(2021, 6, 1, 14, 0, 0, 0, time.UTC),
					Inflation: sdk.MustNewDecFromStr("1.000000004431822130"), // 15%
				},
				{
					Start:     time.Date(2021, 6, 1, 14, 0, 0, 0, time.UTC),
					End:       time.Date(2022, 6, 1, 14, 0, 0, 0, time.UTC),
					Inflation: sdk.MustNewDecFromStr("1.000000002293273137"), // 7.5%
				},
				{
					Start:     time.Date(2022, 6, 1, 14, 0, 0, 0, time.UTC),
					End:       time.Date(2023, 6, 1, 14, 0, 0, 0, time.UTC),
					Inflation: sdk.MustNewDecFromStr("1.000000001167363430"), // 3.75%
				},
				{
					Start:     time.Date(2023, 6, 1, 14, 0, 0, 0, time.UTC),
					End:       time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC),
					Inflation: sdk.MustNewDecFromStr("1.000000000782997609"), // 2.5%
				},
			},
		),
		kavadist.DefaultPreviousBlockTime,
	))
}

func addPricefeedState(cdc *codec.Codec, appState genutil.AppMap) {
	appState[pricefeed.ModuleName] = cdc.MustMarshalJSON(pricefeed.NewGenesisState(
		pricefeed.NewParams(
			pricefeed.Markets{
				{
					MarketID:   bnbSpotMarketID,
					BaseAsset:  bnbDenom,
					QuoteAsset: referenceAsset,
					Oracles: []sdk.AccAddress{
						// addresses from governance proposal: https://ipfs.io/ipfs/QmXgSJ4Dcji8msKpDwYHLmfPSLjRxCEGX6egXQU9DzmFMK/oracle-nominations.pdf
						mustAccAddressFromBech32("kava12dyshua9nkvx9w8ywp72wdnzrc4t4mnnycz0dl"),
						mustAccAddressFromBech32("kava1tuxyepdrkwraa22k99w04c0wa64tgh70mv87fs"),
						mustAccAddressFromBech32("kava1ueak7nzesm3pnev6lngp6lgk0ry02djz8pjpcg"),
						mustAccAddressFromBech32("kava1sl62nqm89c780yxm3m9lp3tacmpnfljq6tytvl"),
						mustAccAddressFromBech32("kava1ujfrlcd0ted58mzplnyxzklsw0sqevlgxndanp"),
						mustAccAddressFromBech32("kava1266f45d6te0wlkswp24phvqrnf0ddpyzhv6ycp"),
						mustAccAddressFromBech32("kava19rjk5qmmwywnzfccwzyn02jywgpwjqf60afj92"),
						mustAccAddressFromBech32("kava1xd39avn2f008jmvua0eupg39zsp2xn3wf802vn"),
						mustAccAddressFromBech32("kava1pt6q4kdmwawr3thm9cd82pq7hml8u84rd0f3jy"),
						mustAccAddressFromBech32("kava13tpwqygswyzupqfggfgh9dmtgthgucn5wpfksh"),
					},
					Active: true,
				},
				{
					MarketID:   bnbLiquidationMarketID,
					BaseAsset:  bnbDenom,
					QuoteAsset: referenceAsset,
					Oracles: []sdk.AccAddress{
						// addresses from governance proposal: https://ipfs.io/ipfs/QmXgSJ4Dcji8msKpDwYHLmfPSLjRxCEGX6egXQU9DzmFMK/oracle-nominations.pdf
						mustAccAddressFromBech32("kava12dyshua9nkvx9w8ywp72wdnzrc4t4mnnycz0dl"),
						mustAccAddressFromBech32("kava1tuxyepdrkwraa22k99w04c0wa64tgh70mv87fs"),
						mustAccAddressFromBech32("kava1ueak7nzesm3pnev6lngp6lgk0ry02djz8pjpcg"),
						mustAccAddressFromBech32("kava1sl62nqm89c780yxm3m9lp3tacmpnfljq6tytvl"),
						mustAccAddressFromBech32("kava1ujfrlcd0ted58mzplnyxzklsw0sqevlgxndanp"),
						mustAccAddressFromBech32("kava1266f45d6te0wlkswp24phvqrnf0ddpyzhv6ycp"),
						mustAccAddressFromBech32("kava19rjk5qmmwywnzfccwzyn02jywgpwjqf60afj92"),
						mustAccAddressFromBech32("kava1xd39avn2f008jmvua0eupg39zsp2xn3wf802vn"),
						mustAccAddressFromBech32("kava1pt6q4kdmwawr3thm9cd82pq7hml8u84rd0f3jy"),
						mustAccAddressFromBech32("kava13tpwqygswyzupqfggfgh9dmtgthgucn5wpfksh"),
					},
					Active: true,
				},
			},
		),
		pricefeed.PostedPrices{},
	))
}

func mustAccAddressFromBech32(addrBech32 string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(addrBech32)
	if err != nil {
		panic(fmt.Errorf("couldn't decode address: %w", err))
	}
	return addr
}
