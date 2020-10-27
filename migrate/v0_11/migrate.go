package v0_11

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v39_1auth "github.com/cosmos/cosmos-sdk/x/auth"
	v39_1authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	v39_1vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	v39_genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	v39_1gov "github.com/cosmos/cosmos-sdk/x/gov"
	v39_1supply "github.com/cosmos/cosmos-sdk/x/supply"

	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	v38_5auth "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/auth"
	v38_5supply "github.com/kava-labs/kava/migrate/v0_11/legacy/cosmos-sdk/v0.38.5/supply"
	v0_11bep3 "github.com/kava-labs/kava/x/bep3"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11cdp "github.com/kava-labs/kava/x/cdp"
	v0_9cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_9"
	v0_11committee "github.com/kava-labs/kava/x/committee"
	v0_9committee "github.com/kava-labs/kava/x/committee/legacy/v0_9"
	v0_11harvest "github.com/kava-labs/kava/x/harvest"
	v0_11incentive "github.com/kava-labs/kava/x/incentive"
	v0_9incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_9"
	v0_11issuance "github.com/kava-labs/kava/x/issuance"
	v0_11pricefeed "github.com/kava-labs/kava/x/pricefeed"
	v0_9pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_9"
	v0_11validator_vesting "github.com/kava-labs/kava/x/validator-vesting"
	v0_9validator_vesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_9"
)

var deputyBnbBalance sdk.Coin
var hardBalance sdk.Coin

// Migrate translates a genesis file from kava v0.9 (or v0.10) format to kava v0.11.x format.
func Migrate(genDoc tmtypes.GenesisDoc) tmtypes.GenesisDoc {
	// migrate app state
	var appStateMap v39_genutil.AppMap
	cdc := codec.New()
	cryptoAmino.RegisterAmino(cdc)
	tmtypes.RegisterEvidences(cdc)

	if err := cdc.UnmarshalJSON(genDoc.AppState, &appStateMap); err != nil {
		panic(err)
	}
	newAppState := MigrateAppState(appStateMap)
	v0_11Codec := app.MakeCodec()
	marshaledNewAppState, err := v0_11Codec.MarshalJSON(newAppState)
	if err != nil {
		panic(err)
	}
	genDoc.AppState = marshaledNewAppState
	genDoc.GenesisTime = time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC)
	genDoc.ChainID = "kava-4"
	return genDoc
}

// MigrateAppState migrates application state from v0.9 (or v0.10) format to a kava v0.11.x format
func MigrateAppState(v0_9AppState v39_genutil.AppMap) v39_genutil.AppMap {
	v0_11AppState := v0_9AppState
	v0_11Codec := app.MakeCodec()
	if v0_9AppState[v38_5auth.ModuleName] != nil {
		v0_9cdc := codec.New()
		codec.RegisterCrypto(v0_9cdc)
		v38_5auth.RegisterCodec(v0_9cdc)
		v38_5auth.RegisterCodecVesting(v0_9cdc)
		v38_5supply.RegisterCodec(v0_9cdc)
		v0_9validator_vesting.RegisterCodec(v0_9cdc)
		var authGenState v38_5auth.GenesisState
		v0_9cdc.MustUnmarshalJSON(v0_9AppState[v38_5auth.ModuleName], &authGenState)
		delete(v0_9AppState, v38_5auth.ModuleName)
		newAuthGS := MigrateAuth(authGenState)
		v0_11AppState[v39_1auth.ModuleName] = v0_11Codec.MustMarshalJSON(newAuthGS)
	}
	if v0_9AppState[v39_1supply.ModuleName] != nil {
		var supplyGenstate v39_1supply.GenesisState
		v0_11Codec.MustUnmarshalJSON(v0_9AppState[v39_1supply.ModuleName], &supplyGenstate)
		delete(v0_9AppState, v39_1supply.ModuleName)
		v0_11AppState[v39_1supply.ModuleName] = v0_11Codec.MustMarshalJSON(
			MigrateSupply(
				supplyGenstate, deputyBnbBalance, sdk.NewCoin("hard", sdk.NewInt(200000000000000))))

	}
	if v0_9AppState[v39_1gov.ModuleName] != nil {
		var govGenstate v39_1gov.GenesisState
		v0_11Codec.MustUnmarshalJSON(v0_9AppState[v39_1gov.ModuleName], &govGenstate)
		delete(v0_9AppState, v39_1gov.ModuleName)
		v0_11AppState[v39_1gov.ModuleName] = v0_11Codec.MustMarshalJSON(
			MigrateGov(govGenstate))

	}
	if v0_9AppState[v0_9bep3.ModuleName] != nil {
		var bep3GenState v0_9bep3.GenesisState
		v0_11Codec.MustUnmarshalJSON(v0_9AppState[v0_9bep3.ModuleName], &bep3GenState)
		delete(v0_9AppState, v0_9bep3.ModuleName)
		v0_11AppState[v0_9bep3.ModuleName] = v0_11Codec.MustMarshalJSON(MigrateBep3(bep3GenState))
	}
	if v0_9AppState[v0_9cdp.ModuleName] != nil {
		var cdpGenState v0_9cdp.GenesisState
		v0_11Codec.MustUnmarshalJSON(v0_9AppState[v0_9cdp.ModuleName], &cdpGenState)
		delete(v0_9AppState, v0_9cdp.ModuleName)
		v0_11AppState[v0_9cdp.ModuleName] = v0_11Codec.MustMarshalJSON(MigrateCDP(cdpGenState))
	}
	if v0_9AppState[v0_9committee.ModuleName] != nil {
		var committeeGenState v0_9committee.GenesisState
		cdc := codec.New()
		sdk.RegisterCodec(cdc)
		v0_9committee.RegisterCodec(cdc)
		cdc.MustUnmarshalJSON(v0_9AppState[v0_9committee.ModuleName], &committeeGenState)
		delete(v0_9AppState, v0_9committee.ModuleName)
		v0_11AppState[v0_9committee.ModuleName] = v0_11Codec.MustMarshalJSON(MigrateCommittee(committeeGenState))
	}
	if v0_9AppState[v0_9incentive.ModuleName] != nil {
		var incentiveGenState v0_9incentive.GenesisState
		v0_11Codec.MustUnmarshalJSON(v0_9AppState[v0_9incentive.ModuleName], &incentiveGenState)
		delete(v0_9AppState, v0_9incentive.ModuleName)
		v0_11AppState[v0_9incentive.ModuleName] = v0_11Codec.MustMarshalJSON(MigrateIncentive(incentiveGenState))
	}
	if v0_9AppState[v0_9pricefeed.ModuleName] != nil {
		var pricefeedGenState v0_9pricefeed.GenesisState
		v0_11Codec.MustUnmarshalJSON(v0_9AppState[v0_9pricefeed.ModuleName], &pricefeedGenState)
		delete(v0_9AppState, v0_9pricefeed.ModuleName)
		v0_11AppState[v0_9pricefeed.ModuleName] = v0_11Codec.MustMarshalJSON(MigratePricefeed(pricefeedGenState))
	}
	// v0_11AppState[v0_11harvest.ModuleName] = v0_11Codec.MustMarshalJSON(MigrateHarvest())
	v0_11AppState[v0_11issuance.ModuleName] = v0_11Codec.MustMarshalJSON(v0_11issuance.DefaultGenesisState())
	return v0_11AppState
}

// MigrateBep3 migrates from a v0.9 (or v0.10) bep3 genesis state to a v0.11 bep3 genesis state
func MigrateBep3(oldGenState v0_9bep3.GenesisState) v0_11bep3.GenesisState {
	var assetParams v0_11bep3.AssetParams
	var assetSupplies v0_11bep3.AssetSupplies
	v0_9Params := oldGenState.Params

	for _, asset := range v0_9Params.SupportedAssets {
		v11AssetParam := v0_11bep3.AssetParam{
			Active:        asset.Active,
			Denom:         asset.Denom,
			CoinID:        asset.CoinID,
			DeputyAddress: v0_9Params.BnbDeputyAddress,
			FixedFee:      v0_9Params.BnbDeputyFixedFee,
			MinSwapAmount: v0_9Params.BnbDeputyFixedFee.Add(sdk.OneInt()), // set min swap to one (after fees)- prevents accounts that hold zero bnb from creating spam txs
			MaxSwapAmount: v0_9Params.MaxAmount,
			MinBlockLock:  v0_9Params.MinBlockLock,
			MaxBlockLock:  v0_9Params.MaxBlockLock,
			SupplyLimit: v0_11bep3.SupplyLimit{
				Limit:          asset.Limit,
				TimeLimited:    false,
				TimePeriod:     time.Duration(0),
				TimeBasedLimit: sdk.ZeroInt(),
			},
		}
		assetParams = append(assetParams, v11AssetParam)
	}
	for _, supply := range oldGenState.AssetSupplies {
		newSupply := v0_11bep3.NewAssetSupply(supply.IncomingSupply, supply.OutgoingSupply, supply.CurrentSupply, sdk.NewCoin(supply.CurrentSupply.Denom, sdk.ZeroInt()), time.Duration(0))
		assetSupplies = append(assetSupplies, newSupply)
	}
	var swaps v0_11bep3.AtomicSwaps
	for _, oldSwap := range oldGenState.AtomicSwaps {
		newSwap := v0_11bep3.AtomicSwap{
			Amount:              oldSwap.Amount,
			RandomNumberHash:    oldSwap.RandomNumberHash,
			ExpireHeight:        oldSwap.ExpireHeight,
			Timestamp:           oldSwap.Timestamp,
			Sender:              oldSwap.Sender,
			Recipient:           oldSwap.Recipient,
			SenderOtherChain:    oldSwap.SenderOtherChain,
			RecipientOtherChain: oldSwap.RecipientOtherChain,
			ClosedBlock:         oldSwap.ClosedBlock,
			Status:              v0_11bep3.SwapStatus(oldSwap.Status),
			CrossChain:          oldSwap.CrossChain,
			Direction:           v0_11bep3.SwapDirection(oldSwap.Direction),
		}
		swaps = append(swaps, newSwap)
	}

	// -------------- ADD BTCB To BEP3 params --------------------
	btcbAssetParam := v0_11bep3.NewAssetParam(
		"btcb",
		0,
		v0_11bep3.SupplyLimit{
			Limit:          sdk.NewInt(100000000), // 1 BTC limit at launch
			TimeLimited:    false,
			TimePeriod:     time.Duration(0),
			TimeBasedLimit: sdk.ZeroInt()},
		true,
		mustAccAddressFromBech32("kava14qsmvzprqvhwmgql9fr0u3zv9n2qla8zhnm5pc"),
		sdk.NewInt(2), // 2 satoshi fee
		sdk.NewInt(3),
		sdk.NewInt(1000000000),
		220,
		270,
	)
	btcbAssetSupply := v0_11bep3.NewAssetSupply(
		sdk.NewCoin("btcb", sdk.ZeroInt()),
		sdk.NewCoin("btcb", sdk.ZeroInt()),
		sdk.NewCoin("btcb", sdk.ZeroInt()),
		sdk.NewCoin("btcb", sdk.ZeroInt()),
		time.Duration(0))
	assetParams = append(assetParams, btcbAssetParam)
	assetSupplies = append(assetSupplies, btcbAssetSupply)

	// -------------- ADD XRPB To BEP3 params --------------------
	xrpbAssetParam := v0_11bep3.NewAssetParam(
		"xrpb", // NOTE: XRPB has 8 decimals on binance chain, whereas XRP has 6 decimals natively
		144,
		v0_11bep3.SupplyLimit{
			Limit:          sdk.NewInt(1000000000000), // 10,000 XRP limit at launch
			TimeLimited:    false,
			TimePeriod:     time.Duration(0),
			TimeBasedLimit: sdk.ZeroInt()},
		true,
		mustAccAddressFromBech32("kava1c0ju5vnwgpgxnrktfnkccuth9xqc68dcdpzpas"),
		sdk.NewInt(100000), // 0.001 XRP fee
		sdk.NewInt(100001),
		sdk.NewInt(10000000000000),
		220,
		270,
	)
	xrpbAssetSupply := v0_11bep3.NewAssetSupply(
		sdk.NewCoin("xrpb", sdk.ZeroInt()),
		sdk.NewCoin("xrpb", sdk.ZeroInt()),
		sdk.NewCoin("xrpb", sdk.ZeroInt()),
		sdk.NewCoin("xrpb", sdk.ZeroInt()),
		time.Duration(0))
	assetParams = append(assetParams, xrpbAssetParam)
	assetSupplies = append(assetSupplies, xrpbAssetSupply)

	// -------------- ADD BUSD To BEP3 params --------------------
	busdAssetParam := v0_11bep3.NewAssetParam(
		"busd",
		727, // note - no official SLIP 44 ID
		v0_11bep3.SupplyLimit{
			Limit:          sdk.NewInt(100000000000), // 1,000 BUSD limit at launch
			TimeLimited:    false,
			TimePeriod:     time.Duration(0),
			TimeBasedLimit: sdk.ZeroInt()},
		true,
		mustAccAddressFromBech32("kava1hh4x3a4suu5zyaeauvmv7ypf7w9llwlfufjmuu"),
		sdk.NewInt(20000),
		sdk.NewInt(20001),
		sdk.NewInt(1000000000000),
		220,
		270,
	)
	busdAssetSupply := v0_11bep3.NewAssetSupply(
		sdk.NewCoin("busd", sdk.ZeroInt()),
		sdk.NewCoin("busd", sdk.ZeroInt()),
		sdk.NewCoin("busd", sdk.ZeroInt()),
		sdk.NewCoin("busd", sdk.ZeroInt()),
		time.Duration(0))
	assetParams = append(assetParams, busdAssetParam)
	assetSupplies = append(assetSupplies, busdAssetSupply)
	return v0_11bep3.GenesisState{
		Params:            v0_11bep3.NewParams(assetParams),
		AtomicSwaps:       swaps,
		Supplies:          assetSupplies,
		PreviousBlockTime: v0_11bep3.DefaultPreviousBlockTime,
	}
}

// MigrateCommittee migrates from a v0.9 (or v0.10) committee genesis state to a v0.11 committee genesis state
func MigrateCommittee(oldGenState v0_9committee.GenesisState) v0_11committee.GenesisState {
	var newCommittees []v0_11committee.Committee
	var newStabilityCommittee v0_11committee.Committee
	var newSafetyCommittee v0_11committee.Committee
	var newProposals []v0_11committee.Proposal
	var newVotes []v0_11committee.Vote

	for _, committee := range oldGenState.Committees {
		if committee.ID == 1 {
			newStabilityCommittee.Description = committee.Description
			newStabilityCommittee.ID = committee.ID
			newStabilityCommittee.Members = committee.Members
			newStabilityCommittee.VoteThreshold = committee.VoteThreshold
			newStabilityCommittee.ProposalDuration = committee.ProposalDuration
			var newStabilityPermissions []v0_11committee.Permission
			var newStabilitySubParamPermissions v0_11committee.SubParamChangePermission
			for _, permission := range committee.Permissions {
				subPermission, ok := permission.(v0_9committee.SubParamChangePermission)
				if ok {
					oldCollateralParam := subPermission.AllowedCollateralParams[0]
					newCollateralParam := v0_11committee.AllowedCollateralParam{
						Type:                "bnb-a",
						Denom:               false,
						AuctionSize:         oldCollateralParam.AuctionSize,
						ConversionFactor:    oldCollateralParam.ConversionFactor,
						DebtLimit:           oldCollateralParam.DebtLimit,
						LiquidationMarketID: oldCollateralParam.LiquidationMarketID,
						SpotMarketID:        oldCollateralParam.SpotMarketID,
						LiquidationPenalty:  oldCollateralParam.LiquidationPenalty,
						LiquidationRatio:    oldCollateralParam.LiquidationRatio,
						Prefix:              oldCollateralParam.Prefix,
						StabilityFee:        oldCollateralParam.StabilityFee,
					}
					oldDebtParam := subPermission.AllowedDebtParam
					newDebtParam := v0_11committee.AllowedDebtParam{
						ConversionFactor: oldDebtParam.ConversionFactor,
						DebtFloor:        oldDebtParam.DebtFloor,
						Denom:            oldDebtParam.Denom,
						ReferenceAsset:   oldDebtParam.ReferenceAsset,
						SavingsRate:      oldDebtParam.SavingsRate,
					}
					oldAssetParam := subPermission.AllowedAssetParams[0]
					newAssetParam := v0_11committee.AllowedAssetParam{
						Active:        oldAssetParam.Active,
						CoinID:        oldAssetParam.CoinID,
						Denom:         oldAssetParam.Denom,
						Limit:         oldAssetParam.Limit,
						MaxSwapAmount: true,
						MinBlockLock:  true,
					}
					oldMarketParams := subPermission.AllowedMarkets
					var newMarketParams v0_11committee.AllowedMarkets
					for _, oldMarketParam := range oldMarketParams {
						newMarketParam := v0_11committee.AllowedMarket(oldMarketParam)
						newMarketParams = append(newMarketParams, newMarketParam)
					}
					// add btc, xrp, busd markets to committee
					btcMarketParam := v0_11committee.AllowedMarket{
						MarketID:   "btc:usd",
						BaseAsset:  false,
						QuoteAsset: false,
						Oracles:    false,
						Active:     true,
					}
					btc30MarketParam := v0_11committee.AllowedMarket{
						MarketID:   "btc:usd:30",
						BaseAsset:  false,
						QuoteAsset: false,
						Oracles:    false,
						Active:     true,
					}
					xrpMarketParam := v0_11committee.AllowedMarket{
						MarketID:   "xrp:usd",
						BaseAsset:  false,
						QuoteAsset: false,
						Oracles:    false,
						Active:     true,
					}
					xrp30MarketParam := v0_11committee.AllowedMarket{
						MarketID:   "xrp:usd:30",
						BaseAsset:  false,
						QuoteAsset: false,
						Oracles:    false,
						Active:     true,
					}
					busdMarketParam := v0_11committee.AllowedMarket{
						MarketID:   "busd:usd",
						BaseAsset:  false,
						QuoteAsset: false,
						Oracles:    false,
						Active:     true,
					}
					busd30MarketParam := v0_11committee.AllowedMarket{
						MarketID:   "busd:usd:30",
						BaseAsset:  false,
						QuoteAsset: false,
						Oracles:    false,
						Active:     true,
					}
					newMarketParams = append(newMarketParams, btcMarketParam, btc30MarketParam, xrpMarketParam, xrp30MarketParam, busdMarketParam, busd30MarketParam)
					oldAllowedParams := subPermission.AllowedParams
					var newAllowedParams v0_11committee.AllowedParams
					for _, oldAllowedParam := range oldAllowedParams {
						newAllowedParam := v0_11committee.AllowedParam(oldAllowedParam)
						if oldAllowedParam.Subspace == "bep3" && oldAllowedParam.Key == "SupportedAssets" {
							newAllowedParam.Key = "AssetParams"
						}
						harvestParam := v0_11committee.AllowedParam{Subspace: "harvest", Key: "Active"}

						newAllowedParams = append(newAllowedParams, newAllowedParam, harvestParam)
					}

					// --------------- ADD BUSD, XRP-B, BTC-B BEP3 parameters to Stability Committee Permissions
					busdAllowedAssetParam := v0_11committee.AllowedAssetParam{
						Active:        true,
						CoinID:        true, // allow busd coinID to be updated in case it gets its own slip-44
						Denom:         "busd",
						Limit:         true,
						MaxSwapAmount: true,
						MinBlockLock:  true,
					}
					xrpbAllowedAssetParam := v0_11committee.AllowedAssetParam{
						Active:        true,
						CoinID:        false,
						Denom:         "xrpb",
						Limit:         true,
						MaxSwapAmount: true,
						MinBlockLock:  true,
					}
					btcbAllowedAssetParam := v0_11committee.AllowedAssetParam{
						Active:        true,
						CoinID:        false,
						Denom:         "btcb",
						Limit:         true,
						MaxSwapAmount: true,
						MinBlockLock:  true,
					}
					// --------- ADD BTC-B, XRP-B, BUSD(a), BUSD(b) cdp collateral params to stability committee
					busdaAllowedCollateralParam := v0_11committee.NewAllowedCollateralParam(
						"busd-a", false, false, true, true, true, false, false, false, false, false,
					)
					busdbAllowedCollateralParam := v0_11committee.NewAllowedCollateralParam(
						"busd-b", false, false, true, true, true, false, false, false, false, false,
					)
					btcbAllowedCollateralParam := v0_11committee.NewAllowedCollateralParam(
						"btcb-a", false, false, true, true, true, false, false, false, false, false,
					)
					xrpbAllowedCollateralParam := v0_11committee.NewAllowedCollateralParam(
						"xrpb-a", false, false, true, true, true, false, false, false, false, false,
					)

					newStabilitySubParamPermissions.AllowedAssetParams = v0_11committee.AllowedAssetParams{
						newAssetParam, busdAllowedAssetParam, btcbAllowedAssetParam, xrpbAllowedAssetParam}
					newStabilitySubParamPermissions.AllowedCollateralParams = v0_11committee.AllowedCollateralParams{
						newCollateralParam, busdaAllowedCollateralParam, busdbAllowedCollateralParam, btcbAllowedCollateralParam, xrpbAllowedCollateralParam}
					newStabilitySubParamPermissions.AllowedDebtParam = newDebtParam
					newStabilitySubParamPermissions.AllowedMarkets = newMarketParams
					newStabilitySubParamPermissions.AllowedParams = newAllowedParams
					newStabilityPermissions = append(newStabilityPermissions, newStabilitySubParamPermissions)
				}
			}
			newStabilityPermissions = append(newStabilityPermissions, v0_11committee.TextPermission{})
			newStabilityCommittee.Permissions = newStabilityPermissions
			newCommittees = append(newCommittees, newStabilityCommittee)
		} else {
			newSafetyCommittee.ID = committee.ID
			newSafetyCommittee.Description = committee.Description
			newSafetyCommittee.Members = committee.Members
			newSafetyCommittee.Permissions = []v0_11committee.Permission{v0_11committee.SoftwareUpgradePermission{}}
			newSafetyCommittee.VoteThreshold = committee.VoteThreshold
			newSafetyCommittee.ProposalDuration = committee.ProposalDuration
			newCommittees = append(newCommittees, newSafetyCommittee)
		}
	}
	for _, oldProp := range oldGenState.Proposals {
		newPubProposal := v0_11committee.PubProposal(oldProp.PubProposal)
		newProp := v0_11committee.NewProposal(newPubProposal, oldProp.ID, oldProp.CommitteeID, oldProp.Deadline)
		newProposals = append(newProposals, newProp)
	}

	for _, oldVote := range oldGenState.Votes {
		newVote := v0_11committee.NewVote(oldVote.ProposalID, oldVote.Voter)
		newVotes = append(newVotes, newVote)
	}

	return v0_11committee.GenesisState{
		NextProposalID: oldGenState.NextProposalID,
		Committees:     newCommittees,
		Proposals:      newProposals,
		Votes:          newVotes,
	}
}

// MigrateAuth migrates from a v0.38.5 auth genesis state to a v0.39.1 auth genesis state
func MigrateAuth(oldGenState v38_5auth.GenesisState) v39_1auth.GenesisState {
	var newAccounts v39_1authexported.GenesisAccounts
	deputyBnbBalance = sdk.NewCoin("bnb", sdk.ZeroInt())
	deputyAddr, err := sdk.AccAddressFromBech32("kava1r4v2zdhdalfj2ydazallqvrus9fkphmglhn6u6")
	if err != nil {
		panic(err)
	}
	deputyColdAddr, err := sdk.AccAddressFromBech32("kava1qm2u6nyv7kg6awdm46caccgzn5h7mdkde0sue6")
	if err != nil {
		panic(err)
	}
	for _, account := range oldGenState.Accounts {
		switch acc := account.(type) {
		case *v38_5auth.BaseAccount:
			a := v39_1auth.BaseAccount(*acc)
			// Remove deputy bnb
			if a.GetAddress().Equals(deputyAddr) || a.GetAddress().Equals(deputyColdAddr) {
				deputyBnbBalance = deputyBnbBalance.Add(sdk.NewCoin("bnb", a.GetCoins().AmountOf("bnb")))
				err := a.SetCoins(a.GetCoins().Sub(sdk.NewCoins(sdk.NewCoin("bnb", a.GetCoins().AmountOf("bnb")))))
				if err != nil {
					panic(err)
				}
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&a))

		case *v38_5auth.BaseVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.OriginalVesting,
				DelegatedFree:    acc.DelegatedFree,
				DelegatedVesting: acc.DelegatedVesting,
				EndTime:          acc.EndTime,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&bva))

		case *v38_5auth.ContinuousVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			cva := v39_1vesting.ContinuousVestingAccount{
				BaseVestingAccount: &bva,
				StartTime:          acc.StartTime,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&cva))

		case *v38_5auth.DelayedVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			dva := v39_1vesting.DelayedVestingAccount{
				BaseVestingAccount: &bva,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&dva))

		case *v38_5auth.PeriodicVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseVestingAccount.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			var newPeriods v39_1vesting.Periods
			for _, p := range acc.VestingPeriods {
				newPeriods = append(newPeriods, v39_1vesting.Period(p))
			}
			pva := v39_1vesting.PeriodicVestingAccount{
				BaseVestingAccount: &bva,
				StartTime:          acc.StartTime,
				VestingPeriods:     newPeriods,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&pva))

		case *v38_5supply.ModuleAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseAccount))
			ma := v39_1supply.ModuleAccount{
				BaseAccount: &ba,
				Name:        acc.Name,
				Permissions: acc.Permissions,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&ma))

		case *v0_9validator_vesting.ValidatorVestingAccount:
			ba := v39_1auth.BaseAccount(*(acc.BaseAccount))
			bva := v39_1vesting.BaseVestingAccount{
				BaseAccount:      &ba,
				OriginalVesting:  acc.BaseVestingAccount.OriginalVesting,
				DelegatedFree:    acc.BaseVestingAccount.DelegatedFree,
				DelegatedVesting: acc.BaseVestingAccount.DelegatedVesting,
				EndTime:          acc.BaseVestingAccount.EndTime,
			}
			var newPeriods v39_1vesting.Periods
			for _, p := range acc.VestingPeriods {
				newPeriods = append(newPeriods, v39_1vesting.Period(p))
			}
			pva := v39_1vesting.PeriodicVestingAccount{
				BaseVestingAccount: &bva,
				StartTime:          acc.StartTime,
				VestingPeriods:     newPeriods,
			}
			var newVestingProgress []v0_11validator_vesting.VestingProgress
			for _, p := range acc.VestingPeriodProgress {
				newVestingProgress = append(newVestingProgress, v0_11validator_vesting.VestingProgress(p))
			}
			vva := v0_11validator_vesting.ValidatorVestingAccount{
				PeriodicVestingAccount: &pva,
				ValidatorAddress:       acc.ValidatorAddress,
				ReturnAddress:          acc.ReturnAddress,
				SigningThreshold:       acc.SigningThreshold,
				CurrentPeriodProgress:  v0_11validator_vesting.CurrentPeriodProgress(acc.CurrentPeriodProgress),
				VestingPeriodProgress:  newVestingProgress,
				DebtAfterFailedVesting: acc.DebtAfterFailedVesting,
			}
			newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(&vva))

		default:
			panic(fmt.Sprintf("unrecognized account type: %T", acc))
		}
	}

	// ---- add harvest module accounts -------
	lpMacc := v39_1supply.NewEmptyModuleAccount(v0_11harvest.LPAccount, v39_1supply.Minter, v39_1supply.Burner)
	err = lpMacc.SetCoins(sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(80000000000000))))
	if err != nil {
		panic(err)
	}
	newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(lpMacc))
	delegatorMacc := v39_1supply.NewEmptyModuleAccount(v0_11harvest.DelegatorAccount, v39_1supply.Minter, v39_1supply.Burner)
	err = delegatorMacc.SetCoins(sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(40000000000000))))
	if err != nil {
		panic(err)
	}
	newAccounts = append(newAccounts, v39_1authexported.GenesisAccount(delegatorMacc))
	hardBalance = sdk.NewCoin("hard", sdk.NewInt(200000000000000))

	hardTeam := createHardTeamVestingAccount()
	hardTreasury := createHardTreasuryVestingAccount()
	hardIEO := createHardIEOAccount()
	newAccounts = append(newAccounts, hardTeam, hardTreasury, &hardIEO)

	return v39_1auth.NewGenesisState(v39_1auth.Params(oldGenState.Params), newAccounts)

}

// MigrateSupply reconciles supply from kava-3 to kava-4
// deputy balance of bnb coins is removed (deputy now mints coins)
// hard token supply is added
func MigrateSupply(oldGenState v39_1supply.GenesisState, deputyBalance sdk.Coin, hardBalance sdk.Coin) v39_1supply.GenesisState {
	oldGenState.Supply = oldGenState.Supply.Sub(sdk.Coins{deputyBalance}).Add(hardBalance)
	return oldGenState
}

// MigrateGov migrates gov genesis state
func MigrateGov(oldGenState v39_1gov.GenesisState) v39_1gov.GenesisState {
	oldGenState.VotingParams.VotingPeriod = time.Hour * 24 * 7
	return oldGenState
}

// // MigrateHarvest initializes the harvest genesis state for kava-4
// func MigrateHarvest() v0_11harvest.GenesisState {
// 	// total HARD per second for lps (week one): 633761
// 	// HARD per second for delegators (week one): 1267522
// 	incentiveGoLiveDate := time.Date(2020, 10, 16, 14, 0, 0, 0, time.UTC)
// 	incentiveEndDate := time.Date(2024, 10, 16, 14, 0, 0, 0, time.UTC)
// 	claimEndDate := time.Date(2025, 10, 16, 14, 0, 0, 0, time.UTC)
// 	harvestGS := v0_11harvest.NewGenesisState(v0_11harvest.NewParams(
// 		true,
// 		v0_11harvest.DistributionSchedules{
// 			v0_11harvest.NewDistributionSchedule(true, "usdx", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(310543)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
// 			v0_11harvest.NewDistributionSchedule(true, "hard", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(285193)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
// 			v0_11harvest.NewDistributionSchedule(true, "bnb", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(12675)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
// 			v0_11harvest.NewDistributionSchedule(true, "ukava", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(25350)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
// 		},
// 		v0_11harvest.DelegatorDistributionSchedules{v0_11harvest.NewDelegatorDistributionSchedule(
// 			v0_11harvest.NewDistributionSchedule(true, "ukava", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(1267522)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
// 			time.Hour*24,
// 		),
// 		},
// 		v0_11harvest.BlockLimits{
// 			v0_11harvest.NewBlockLimit("usdx", sdk.Dec(0.9)),
// 			v0_11harvest.NewBlockLimit("ukava", sdk.Dec(0.6)),
// 			v0_11harvest.NewBlockLimit("bnb", sdk.Dec(0.9)),
// 		},
// 	), v0_11harvest.DefaultPreviousBlockTime, v0_11harvest.DefaultDistributionTimes)
// 	return harvestGS
// }

// MigrateCDP migrates from a v0.9 (or v0.10) cdp genesis state to a v0.11 cdp genesis state
func MigrateCDP(oldGenState v0_9cdp.GenesisState) v0_11cdp.GenesisState {
	var newCDPs v0_11cdp.CDPs
	var newDeposits v0_11cdp.Deposits
	var newCollateralParams v0_11cdp.CollateralParams
	newStartingID := oldGenState.StartingCdpID

	for _, cdp := range oldGenState.CDPs {
		newCDP := v0_11cdp.NewCDPWithFees(cdp.ID, cdp.Owner, cdp.Collateral, "bnb-a", cdp.Principal, cdp.AccumulatedFees, cdp.FeesUpdated)
		newCDPs = append(newCDPs, newCDP)
	}

	for _, dep := range oldGenState.Deposits {
		newDep := v0_11cdp.NewDeposit(dep.CdpID, dep.Depositor, dep.Amount)
		newDeposits = append(newDeposits, newDep)
	}

	for _, cp := range oldGenState.Params.CollateralParams {
		newCollateralParam := v0_11cdp.NewCollateralParam(cp.Denom, "bnb-a", cp.LiquidationRatio, cp.DebtLimit, cp.StabilityFee, cp.AuctionSize, cp.LiquidationPenalty, 0x01, cp.SpotMarketID, cp.LiquidationMarketID, cp.ConversionFactor)
		newCollateralParams = append(newCollateralParams, newCollateralParam)
	}
	btcbCollateralParam := v0_11cdp.NewCollateralParam("btcb", "btcb-a", sdk.MustNewDecFromStr("1.5"), sdk.NewCoin("usdx", sdk.NewInt(100000000000)), sdk.MustNewDecFromStr("1.000000001547125958"), sdk.NewInt(100000000), sdk.MustNewDecFromStr("0.075000000000000000"), 0x02, "btc:usd", "btc:usd:30", sdk.NewInt(8))
	busdaCollateralParam := v0_11cdp.NewCollateralParam("busd", "busd-a", sdk.MustNewDecFromStr("1.01"), sdk.NewCoin("usdx", sdk.NewInt(3000000000000)), sdk.OneDec(), sdk.NewInt(1000000000000), sdk.MustNewDecFromStr("0.075000000000000000"), 0x03, "busd:usd", "busd:usd:30", sdk.NewInt(8))
	busdbCollateralParam := v0_11cdp.NewCollateralParam("busd", "busd-b", sdk.MustNewDecFromStr("1.1"), sdk.NewCoin("usdx", sdk.NewInt(1000000000000)), sdk.MustNewDecFromStr("1.000000012857214317"), sdk.NewInt(1000000000000), sdk.MustNewDecFromStr("0.075000000000000000"), 0x04, "busd:usd", "busd:usd:30", sdk.NewInt(8))
	xrpbCollateralParam := v0_11cdp.NewCollateralParam("xrpb", "xrpb-a", sdk.MustNewDecFromStr("1.5"), sdk.NewCoin("usdx", sdk.NewInt(100000000000)), sdk.MustNewDecFromStr("1.000000001547125958"), sdk.NewInt(4000000000000), sdk.MustNewDecFromStr("0.075000000000000000"), 0x05, "xrp:usd", "xrp:usd:30", sdk.NewInt(8))
	newCollateralParams = append(newCollateralParams, btcbCollateralParam, busdaCollateralParam, busdbCollateralParam, xrpbCollateralParam)
	oldDebtParam := oldGenState.Params.DebtParam

	newDebtParam := v0_11cdp.NewDebtParam(oldDebtParam.Denom, oldDebtParam.ReferenceAsset, oldDebtParam.ConversionFactor, oldDebtParam.DebtFloor, oldDebtParam.SavingsRate)

	newGlobalDebtLimit := oldGenState.Params.GlobalDebtLimit.Add(btcbCollateralParam.DebtLimit).Add(busdaCollateralParam.DebtLimit).Add(busdbCollateralParam.DebtLimit).Add(xrpbCollateralParam.DebtLimit)

	newParams := v0_11cdp.NewParams(newGlobalDebtLimit, newCollateralParams, newDebtParam, oldGenState.Params.SurplusAuctionThreshold, oldGenState.Params.SurplusAuctionLot, oldGenState.Params.DebtAuctionThreshold, oldGenState.Params.DebtAuctionLot, oldGenState.Params.SavingsDistributionFrequency, false)

	return v0_11cdp.NewGenesisState(
		newParams,
		newCDPs,
		newDeposits,
		newStartingID,
		oldGenState.DebtDenom,
		oldGenState.GovDenom,
		oldGenState.PreviousDistributionTime,
		sdk.ZeroInt(),
	)
}

// MigrateIncentive migrates from a v0.9 (or v0.10) incentive genesis state to a v0.11 incentive genesis state
func MigrateIncentive(oldGenState v0_9incentive.GenesisState) v0_11incentive.GenesisState {
	var newRewards v0_11incentive.Rewards
	var newRewardPeriods v0_11incentive.RewardPeriods
	var newClaimPeriods v0_11incentive.ClaimPeriods
	var newClaims v0_11incentive.Claims
	var newClaimPeriodIds v0_11incentive.GenesisClaimPeriodIDs

	newMultiplier := v0_11incentive.NewMultiplier(v0_11incentive.Large, 12, sdk.OneDec())
	smallMultiplier := v0_11incentive.NewMultiplier(v0_11incentive.Small, 1, sdk.MustNewDecFromStr("0.25"))

	for _, oldReward := range oldGenState.Params.Rewards {
		newReward := v0_11incentive.NewReward(oldReward.Active, oldReward.Denom+"-a", oldReward.AvailableRewards, oldReward.Duration, v0_11incentive.Multipliers{smallMultiplier, newMultiplier}, oldReward.ClaimDuration)
		newRewards = append(newRewards, newReward)
	}
	newParams := v0_11incentive.NewParams(true, newRewards)

	for _, oldRewardPeriod := range oldGenState.RewardPeriods {

		newRewardPeriod := v0_11incentive.NewRewardPeriod(oldRewardPeriod.Denom+"-a", oldRewardPeriod.Start, oldRewardPeriod.End, oldRewardPeriod.Reward, oldRewardPeriod.ClaimEnd, v0_11incentive.Multipliers{smallMultiplier, newMultiplier})
		newRewardPeriods = append(newRewardPeriods, newRewardPeriod)
	}

	for _, oldClaimPeriod := range oldGenState.ClaimPeriods {
		newClaimPeriod := v0_11incentive.NewClaimPeriod(oldClaimPeriod.Denom+"-a", oldClaimPeriod.ID, oldClaimPeriod.End, v0_11incentive.Multipliers{smallMultiplier, newMultiplier})
		newClaimPeriods = append(newClaimPeriods, newClaimPeriod)
	}

	for _, oldClaim := range oldGenState.Claims {
		newClaim := v0_11incentive.NewClaim(oldClaim.Owner, oldClaim.Reward, oldClaim.Denom+"-a", oldClaim.ClaimPeriodID)
		newClaims = append(newClaims, newClaim)
	}

	for _, oldClaimPeriodID := range oldGenState.NextClaimPeriodIDs {
		newClaimPeriodID := v0_11incentive.GenesisClaimPeriodID{
			CollateralType: oldClaimPeriodID.Denom + "-a",
			ID:             oldClaimPeriodID.ID,
		}
		newClaimPeriodIds = append(newClaimPeriodIds, newClaimPeriodID)
	}

	return v0_11incentive.NewGenesisState(newParams, oldGenState.PreviousBlockTime, newRewardPeriods, newClaimPeriods, newClaims, newClaimPeriodIds)
}

// MigratePricefeed migrates from a v0.9 (or v0.10) pricefeed genesis state to a v0.11 pricefeed genesis state
func MigratePricefeed(oldGenState v0_9pricefeed.GenesisState) v0_11pricefeed.GenesisState {
	var newMarkets v0_11pricefeed.Markets
	var newPostedPrices v0_11pricefeed.PostedPrices
	var oracles []sdk.AccAddress

	for _, market := range oldGenState.Params.Markets {
		newMarket := v0_11pricefeed.NewMarket(market.MarketID, market.BaseAsset, market.QuoteAsset, market.Oracles, market.Active)
		newMarkets = append(newMarkets, newMarket)
		oracles = market.Oracles
	}
	// ------- add btc, xrp, busd markets --------
	btcSpotMarket := v0_11pricefeed.NewMarket("btc:usd", "btc", "usd", oracles, true)
	btcLiquidationMarket := v0_11pricefeed.NewMarket("btc:usd:30", "btc", "usd", oracles, true)
	xrpSpotMarket := v0_11pricefeed.NewMarket("xrp:usd", "xrp", "usd", oracles, true)
	xrpLiquidationMarket := v0_11pricefeed.NewMarket("xrp:usd:30", "xrp", "usd", oracles, true)
	busdSpotMarket := v0_11pricefeed.NewMarket("busd:usd", "busd", "usd", oracles, true)
	busdLiquidationMarket := v0_11pricefeed.NewMarket("busd:usd:30", "busd", "usd", oracles, true)
	newMarkets = append(newMarkets, btcSpotMarket, btcLiquidationMarket, xrpSpotMarket, xrpLiquidationMarket, busdSpotMarket, busdLiquidationMarket)

	for _, price := range oldGenState.PostedPrices {
		newPrice := v0_11pricefeed.NewPostedPrice(price.MarketID, price.OracleAddress, price.Price, price.Expiry)
		newPostedPrices = append(newPostedPrices, newPrice)
	}
	newParams := v0_11pricefeed.NewParams(newMarkets)

	return v0_11pricefeed.NewGenesisState(newParams, newPostedPrices)
}

func mustAccAddressFromBech32(bech32Addr string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(bech32Addr)
	if err != nil {
		panic(err)
	}
	return addr
}

func createHardTeamVestingAccount() *v39_1vesting.PeriodicVestingAccount {
	bacc := v39_1auth.NewBaseAccountWithAddress(mustAccAddressFromBech32("kava17a9m9zxs3r5zhxnultt5k5kyer0afd7kc8dq80"))
	coins := sdk.NewCoin("hard", sdk.NewInt(20000000000000))
	tokenSchedule := []sdk.Coin{
		sdk.NewCoin("hard", sdk.NewInt(6666666720000)),
		sdk.NewCoin("hard", sdk.NewInt(1666666660000)),
		sdk.NewCoin("hard", sdk.NewInt(1666666660000)),
		sdk.NewCoin("hard", sdk.NewInt(1666666660000)),
		sdk.NewCoin("hard", sdk.NewInt(1666666660000)),
		sdk.NewCoin("hard", sdk.NewInt(1666666660000)),
		sdk.NewCoin("hard", sdk.NewInt(1666666660000)),
		sdk.NewCoin("hard", sdk.NewInt(1666666660000)),
		sdk.NewCoin("hard", sdk.NewInt(1666666660000)),
	}
	err := bacc.SetCoins(sdk.NewCoins(coins))
	if err != nil {
		panic(err)
	}
	bva, err := v39_1vesting.NewBaseVestingAccount(&bacc, sdk.NewCoins(coins), 1697378400)
	if err != nil {
		panic(err)
	}
	vestingPeriodLengths := []int64{31536000, 7948800, 7776000, 7862400, 7948800, 7948800, 7776000, 7862400, 7948800}

	periods := v39_1vesting.Periods{}
	for i, vestingCoin := range tokenSchedule {
		period := v39_1vesting.Period{Length: vestingPeriodLengths[i], Amount: sdk.NewCoins(vestingCoin)}
		periods = append(periods, period)
	}
	return vesting.NewPeriodicVestingAccountRaw(bva, 1602770400, periods)
}

func createHardTreasuryVestingAccount() *v39_1vesting.PeriodicVestingAccount {
	bacc := v39_1auth.NewBaseAccountWithAddress(mustAccAddressFromBech32("kava1yqt02z2e4gpt4w0jnw9n0hnqyfu45afns669r2"))
	coins := sdk.NewCoin("hard", sdk.NewInt(50000000000000))
	originalVestingCoins := sdk.NewCoin("hard", sdk.NewInt(35000000000000))
	tokenSchedule := []sdk.Coin{
		sdk.NewCoin("hard", sdk.NewInt(4375000000000)),
		sdk.NewCoin("hard", sdk.NewInt(4375000000000)),
		sdk.NewCoin("hard", sdk.NewInt(4375000000000)),
		sdk.NewCoin("hard", sdk.NewInt(4375000000000)),
		sdk.NewCoin("hard", sdk.NewInt(4375000000000)),
		sdk.NewCoin("hard", sdk.NewInt(4375000000000)),
		sdk.NewCoin("hard", sdk.NewInt(4375000000000)),
		sdk.NewCoin("hard", sdk.NewInt(4375000000000)),
	}
	err := bacc.SetCoins(sdk.NewCoins(coins))
	if err != nil {
		panic(err)
	}
	bva, err := v39_1vesting.NewBaseVestingAccount(&bacc, sdk.NewCoins(originalVestingCoins), 1665842400)
	if err != nil {
		panic(err)
	}
	vestingPeriodLengths := []int64{7948800, 7776000, 7862400, 7948800, 7948800, 7776000, 7862400, 7948800}

	periods := v39_1vesting.Periods{}
	for i, vestingCoin := range tokenSchedule {
		period := v39_1vesting.Period{Length: vestingPeriodLengths[i], Amount: sdk.NewCoins(vestingCoin)}
		periods = append(periods, period)
	}
	return vesting.NewPeriodicVestingAccountRaw(bva, 1602770400, periods)
}

func createHardIEOAccount() v39_1auth.BaseAccount {
	bacc := v39_1auth.NewBaseAccountWithAddress(mustAccAddressFromBech32("kava16yapwtdxm5hkjfpeatr39vhu5c336fgf4utlyf"))
	coins := sdk.NewCoin("hard", sdk.NewInt(10000000000000))
	err := bacc.SetCoins(sdk.NewCoins(coins))
	if err != nil {
		panic(err)
	}
	return bacc
}
