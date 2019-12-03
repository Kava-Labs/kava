package keeper

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/kava-labs/kava/x/cdp/types"
)

// Keeper cdp Keeper
type Keeper struct {
	key             sdk.StoreKey
	cdc             *codec.Codec
	paramSubspace   subspace.Subspace
	pricefeedKeeper types.PricefeedKeeper
	bankKeeper      types.BankKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace, pfk types.PricefeedKeeper, bk types.BankKeeper) Keeper {
	return Keeper{
		key:             key,
		cdc:             cdc,
		paramSubspace:   paramstore.WithKeyTable(types.ParamKeyTable()),
		pricefeedKeeper: pfk,
		bankKeeper:      bk,
	}
}

// ModifyCDP creates, changes, or deletes a CDP
// TODO can/should this function be split up?
func (k Keeper) ModifyCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string, changeInCollateral sdk.Int, changeInDebt sdk.Int) sdk.Error {

	// Phase 1: Get state, make changes in memory and check if they're ok.

	// Check collateral type ok
	p := k.GetParams(ctx)
	if !p.IsCollateralPresent(collateralDenom) { // maybe abstract this logic into GetCDP
		return sdk.ErrInternal("collateral type not enabled to create CDPs")
	}

	// Check the owner has enough collateral and stable coins
	if changeInCollateral.IsPositive() { // adding collateral to CDP
		ok := k.bankKeeper.HasCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateralDenom, changeInCollateral)))
		if !ok {
			return sdk.ErrInsufficientCoins("not enough collateral in sender's account")
		}
	}
	if changeInDebt.IsNegative() { // reducing debt, by adding stable coin to CDP
		ok := k.bankKeeper.HasCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin("usdx", changeInDebt.Neg())))
		if !ok {
			return sdk.ErrInsufficientCoins("not enough stable coin in sender's account")
		}
	}

	// Change collateral and debt recorded in CDP
	// Get CDP (or create if not exists)
	cdp, found := k.GetCDP(ctx, owner, collateralDenom)
	if !found {
		cdp = types.CDP{Owner: owner, CollateralDenom: collateralDenom, CollateralAmount: sdk.ZeroInt(), Debt: sdk.ZeroInt()}
	}
	// Add/Subtract collateral and debt
	cdp.CollateralAmount = cdp.CollateralAmount.Add(changeInCollateral)
	if cdp.CollateralAmount.IsNegative() {
		return sdk.ErrInternal(" can't withdraw more collateral than exists in CDP")
	}
	cdp.Debt = cdp.Debt.Add(changeInDebt)
	if cdp.Debt.IsNegative() {
		return sdk.ErrInternal("can't pay back more debt than exists in CDP")
	}
	isUnderCollateralized := cdp.IsUnderCollateralized(
		k.pricefeedKeeper.GetCurrentPrice(ctx, cdp.CollateralDenom).Price,
		p.GetCollateralParams(cdp.CollateralDenom).LiquidationRatio,
	)
	if isUnderCollateralized {
		return sdk.ErrInternal("Change to CDP would put it below liquidation ratio")
	}
	// TODO check for dust

	// Add/Subtract from global debt limit
	gDebt := k.GetGlobalDebt(ctx)
	gDebt = gDebt.Add(changeInDebt)
	if gDebt.IsNegative() {
		return sdk.ErrInternal("global debt can't be negative") // This should never happen if debt per CDP can't be negative
	}
	if gDebt.GT(p.GlobalDebtLimit) {
		return sdk.ErrInternal("change to CDP would put the system over the global debt limit")
	}

	// Add/Subtract from collateral debt limit
	collateralState, found := k.GetCollateralState(ctx, cdp.CollateralDenom)
	if !found {
		collateralState = types.CollateralState{Denom: cdp.CollateralDenom, TotalDebt: sdk.ZeroInt()} // Already checked that this denom is authorized, so ok to create new CollateralState
	}
	collateralState.TotalDebt = collateralState.TotalDebt.Add(changeInDebt)
	if collateralState.TotalDebt.IsNegative() {
		return sdk.ErrInternal("total debt for this collateral type can't be negative") // This should never happen if debt per CDP can't be negative
	}
	if collateralState.TotalDebt.GT(p.GetCollateralParams(cdp.CollateralDenom).DebtLimit) {
		return sdk.ErrInternal("change to CDP would put the system over the debt limit for this collateral type")
	}

	// Phase 2: Update all the state

	// change owner's coins (increase or decrease)
	var err sdk.Error
	if changeInCollateral.IsNegative() {
		_, err = k.bankKeeper.AddCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateralDenom, changeInCollateral.Neg())))
	} else {
		_, err = k.bankKeeper.SubtractCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateralDenom, changeInCollateral)))
	}
	if err != nil {
		panic(err) // this shouldn't happen because coin balance was checked earlier
	}
	if changeInDebt.IsNegative() {
		_, err = k.bankKeeper.SubtractCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin("usdx", changeInDebt.Neg())))
	} else {
		_, err = k.bankKeeper.AddCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin("usdx", changeInDebt)))
	}
	if err != nil {
		panic(err) // this shouldn't happen because coin balance was checked earlier
	}
	// Set CDP
	if cdp.CollateralAmount.IsZero() && cdp.Debt.IsZero() { // TODO maybe abstract this logic into SetCDP
		k.DeleteCDP(ctx, cdp)
	} else {
		k.SetCDP(ctx, cdp)
	}
	// set total debts
	k.SetGlobalDebt(ctx, gDebt)
	k.SetCollateralState(ctx, collateralState)

	return nil
}

// TODO
// // TransferCDP allows people to transfer ownership of their CDPs to others
// func (k Keeper) TransferCDP(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, collateralDenom string) sdk.Error {
// 	return nil
// }

// PartialSeizeCDP removes collateral and debt from a CDP and decrements global debt counters. It does not move collateral to another account so is unsafe.
// TODO should this be made safer by moving collateral to liquidatorModuleAccount ? If so how should debt be moved?
func (k Keeper) PartialSeizeCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string, collateralToSeize sdk.Int, debtToSeize sdk.Int) sdk.Error {
	// get CDP
	cdp, found := k.GetCDP(ctx, owner, collateralDenom)
	if !found {
		return sdk.ErrInternal("could not find CDP")
	}

	// Check if CDP is undercollateralized
	p := k.GetParams(ctx)
	isUnderCollateralized := cdp.IsUnderCollateralized(
		k.pricefeedKeeper.GetCurrentPrice(ctx, cdp.CollateralDenom).Price,
		p.GetCollateralParams(cdp.CollateralDenom).LiquidationRatio,
	)
	if !isUnderCollateralized {
		return sdk.ErrInternal("CDP is not currently under the liquidation ratio")
	}

	// Remove Collateral
	if collateralToSeize.IsNegative() {
		return sdk.ErrInternal("cannot seize negative collateral")
	}
	cdp.CollateralAmount = cdp.CollateralAmount.Sub(collateralToSeize)
	if cdp.CollateralAmount.IsNegative() {
		return sdk.ErrInternal("can't seize more collateral than exists in CDP")
	}

	// Remove Debt
	if debtToSeize.IsNegative() {
		return sdk.ErrInternal("cannot seize negative debt")
	}
	cdp.Debt = cdp.Debt.Sub(debtToSeize)
	if cdp.Debt.IsNegative() {
		return sdk.ErrInternal("can't seize more debt than exists in CDP")
	}

	// Update debt per collateral type
	collateralState, found := k.GetCollateralState(ctx, cdp.CollateralDenom)
	if !found {
		return sdk.ErrInternal("could not find collateral state")
	}
	collateralState.TotalDebt = collateralState.TotalDebt.Sub(debtToSeize)
	if collateralState.TotalDebt.IsNegative() {
		return sdk.ErrInternal("Total debt per collateral type is negative.") // This should not happen given the checks on the CDP.
	}

	// Note: Global debt is not decremented here. It's only decremented when debt and stable coin are annihilated (aka heal)
	// TODO update global seized debt? this is what maker does (named vice in Vat.grab) but it's not used anywhere

	// Store updated state
	if cdp.CollateralAmount.IsZero() && cdp.Debt.IsZero() { // TODO maybe abstract this logic into SetCDP
		k.DeleteCDP(ctx, cdp)
	} else {
		k.SetCDP(ctx, cdp)
	}
	k.SetCollateralState(ctx, collateralState)
	return nil
}

// ReduceGlobalDebt decreases the stored global debt counter. It is used by the liquidator when it annihilates debt and stable coin.
// TODO Can the interface between cdp and liquidator modules be improved so that this function doesn't exist?
func (k Keeper) ReduceGlobalDebt(ctx sdk.Context, amount sdk.Int) sdk.Error {
	if amount.IsNegative() {
		return sdk.ErrInternal("reduction in global debt must be a positive amount")
	}
	newGDebt := k.GetGlobalDebt(ctx).Sub(amount)
	if newGDebt.IsNegative() {
		return sdk.ErrInternal("cannot reduce global debt by amount specified")
	}
	k.SetGlobalDebt(ctx, newGDebt)
	return nil
}

func (k Keeper) GetStableDenom() string {
	return "usdx"
}
func (k Keeper) GetGovDenom() string {
	return "kava"
}

// ---------- Store Wrappers ----------

func (k Keeper) getCDPKeyPrefix(collateralDenom string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("cdp"),
			[]byte(collateralDenom),
		},
		nil, // no separator
	)
}
func (k Keeper) getCDPKey(owner sdk.AccAddress, collateralDenom string) []byte {
	return bytes.Join(
		[][]byte{
			k.getCDPKeyPrefix(collateralDenom),
			[]byte(owner.String()),
		},
		nil, // no separator
	)
}
func (k Keeper) GetCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string) (types.CDP, bool) {
	// get store
	store := ctx.KVStore(k.key)
	// get CDP
	bz := store.Get(k.getCDPKey(owner, collateralDenom))
	// unmarshal
	if bz == nil {
		return types.CDP{}, false
	}
	var cdp types.CDP
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cdp)
	return cdp, true
}
func (k Keeper) SetCDP(ctx sdk.Context, cdp types.CDP) {
	// get store
	store := ctx.KVStore(k.key)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cdp)
	store.Set(k.getCDPKey(cdp.Owner, cdp.CollateralDenom), bz)
}
func (k Keeper) DeleteCDP(ctx sdk.Context, cdp types.CDP) { // TODO should this id the cdp by passing in owner,collateralDenom pair?
	// get store
	store := ctx.KVStore(k.key)
	// delete key
	store.Delete(k.getCDPKey(cdp.Owner, cdp.CollateralDenom))
}

// GetCDPs returns all CDPs, optionally filtered by collateral type and liquidation price.
// `price` filters for CDPs that will be below the liquidation ratio when the collateral is at that specified price.
func (k Keeper) GetCDPs(ctx sdk.Context, collateralDenom string, price sdk.Dec) (types.CDPs, sdk.Error) {
	// Validate inputs
	params := k.GetParams(ctx)
	if len(collateralDenom) != 0 && !params.IsCollateralPresent(collateralDenom) {
		return nil, sdk.ErrInternal("collateral denom not authorized")
	}
	if len(collateralDenom) == 0 && !(price.IsNil() || price.IsNegative()) {
		return nil, sdk.ErrInternal("cannot specify price without collateral denom")
	}

	// Get an iterator over CDPs
	store := ctx.KVStore(k.key)
	iter := sdk.KVStorePrefixIterator(store, k.getCDPKeyPrefix(collateralDenom)) // could be all CDPs is collateralDenom is ""

	// Decode CDPs into slice
	var cdps types.CDPs
	for ; iter.Valid(); iter.Next() {
		var cdp types.CDP
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &cdp)
		cdps = append(cdps, cdp)
	}

	// Sort by collateral ratio (collateral/debt)
	sort.Sort(types.ByCollateralRatio(cdps)) // TODO this doesn't make much sense across different collateral types

	// Filter for CDPs that would be under-collateralized at the specified price
	// If price is nil or -ve, skip the filtering as it would return all CDPs anyway
	if !price.IsNil() && !price.IsNegative() {
		var filteredCDPs types.CDPs
		for _, cdp := range cdps {
			if cdp.IsUnderCollateralized(price, params.GetCollateralParams(collateralDenom).LiquidationRatio) {
				filteredCDPs = append(filteredCDPs, cdp)
			} else {
				break // break early because list is sorted
			}
		}
		cdps = filteredCDPs
	}

	return cdps, nil
}

var globalDebtKey = []byte("globalDebt")

func (k Keeper) GetGlobalDebt(ctx sdk.Context) sdk.Int {
	// get store
	store := ctx.KVStore(k.key)
	// get bytes
	bz := store.Get(globalDebtKey)
	// unmarshal
	if bz == nil {
		panic("global debt not found")
	}
	var globalDebt sdk.Int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &globalDebt)
	return globalDebt
}
func (k Keeper) SetGlobalDebt(ctx sdk.Context, globalDebt sdk.Int) {
	// get store
	store := ctx.KVStore(k.key)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(globalDebt)
	store.Set(globalDebtKey, bz)
}

func (k Keeper) getCollateralStateKey(collateralDenom string) []byte {
	return []byte(collateralDenom)
}
func (k Keeper) GetCollateralState(ctx sdk.Context, collateralDenom string) (types.CollateralState, bool) {
	// get store
	store := ctx.KVStore(k.key)
	// get bytes
	bz := store.Get(k.getCollateralStateKey(collateralDenom))
	// unmarshal
	if bz == nil {
		return types.CollateralState{}, false
	}
	var collateralState types.CollateralState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &collateralState)
	return collateralState, true
}
func (k Keeper) SetCollateralState(ctx sdk.Context, collateralstate types.CollateralState) {
	// get store
	store := ctx.KVStore(k.key)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(collateralstate)
	store.Set(k.getCollateralStateKey(collateralstate.Denom), bz)
}

// ---------- Weird Bank Stuff ----------
// This only exists because module accounts aren't really a thing yet.
// Also because we need module accounts that allow for burning/minting.

// These functions make the CDP module act as a bank keeper, ie it fulfills the bank.Keeper interface.
// It intercepts calls to send coins to/from the liquidator module account, otherwise passing the calls onto the normal bank keeper.

// Not sure if module accounts are good, but they make the auction module more general:
// - startAuction would just "mints" coins, relying on calling function to decrement them somewhere
// - closeAuction would have to call something specific for the receiver module to accept coins (like liquidationKeeper.AddStableCoins)

// The auction and liquidator modules can probably just use SendCoins to keep things safe (instead of AddCoins and SubtractCoins).
// So they should define their own interfaces which this module should fulfill, rather than this fulfilling the entire bank.Keeper interface.

// bank.Keeper interfaces:
// type SendKeeper interface {
// 	type ViewKeeper interface {
// 		GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
// 		HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool
// 		Codespace() sdk.CodespaceType
// 	}
// 	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
// 	GetSendEnabled(ctx sdk.Context) bool
// 	SetSendEnabled(ctx sdk.Context, enabled bool)
// }
// type Keeper interface {
// 	SendKeeper
// 	SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error
// 	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
// 	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
// 	InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error)
// 	DelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
// 	UndelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)

var LiquidatorAccountAddress = sdk.AccAddress([]byte("whatever"))
var liquidatorAccountKey = []byte("liquidatorAccount")

func (k Keeper) GetLiquidatorAccountAddress() sdk.AccAddress {
	return LiquidatorAccountAddress
}

type LiquidatorModuleAccount struct {
	Coins sdk.Coins // keeps track of seized collateral, surplus usdx, and mints/burns gov coins
}

func (k Keeper) AddCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) (sdk.Coins, sdk.Error) {
	// intercept module account
	if address.Equals(LiquidatorAccountAddress) {
		if !amount.IsValid() {
			return nil, sdk.ErrInvalidCoins(amount.String())
		}
		// remove gov token from list
		filteredCoins := stripGovCoin(amount)
		// add coins to module account
		lma := k.getLiquidatorModuleAccount(ctx)
		updatedCoins := lma.Coins.Add(filteredCoins)
		if updatedCoins.IsAnyNegative() {
			return amount, sdk.ErrInsufficientCoins(fmt.Sprintf("insufficient account funds; %s < %s", lma.Coins, amount))
		}
		lma.Coins = updatedCoins
		k.SetLiquidatorModuleAccount(ctx, lma)
		return updatedCoins, nil
	} else {
		return k.bankKeeper.AddCoins(ctx, address, amount)
	}
}

// TODO abstract stuff better
func (k Keeper) SubtractCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) (sdk.Coins, sdk.Error) {
	// intercept module account
	if address.Equals(LiquidatorAccountAddress) {
		if !amount.IsValid() {
			return nil, sdk.ErrInvalidCoins(amount.String())
		}
		// remove gov token from list
		filteredCoins := stripGovCoin(amount)
		// subtract coins from module account
		lma := k.getLiquidatorModuleAccount(ctx)
		updatedCoins, isNegative := lma.Coins.SafeSub(filteredCoins)
		if isNegative {
			return amount, sdk.ErrInsufficientCoins(fmt.Sprintf("insufficient account funds; %s < %s", lma.Coins, amount))
		}
		lma.Coins = updatedCoins
		k.SetLiquidatorModuleAccount(ctx, lma)
		return updatedCoins, nil
	} else {
		return k.bankKeeper.SubtractCoins(ctx, address, amount)
	}
}

// TODO Should this return anything for the gov coin balance? Currently returns nothing.
func (k Keeper) GetCoins(ctx sdk.Context, address sdk.AccAddress) sdk.Coins {
	if address.Equals(LiquidatorAccountAddress) {
		return k.getLiquidatorModuleAccount(ctx).Coins
	} else {
		return k.bankKeeper.GetCoins(ctx, address)
	}
}

// TODO test this with unsorted coins
func (k Keeper) HasCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) bool {
	if address.Equals(LiquidatorAccountAddress) {
		return true
	} else {
		return k.getLiquidatorModuleAccount(ctx).Coins.IsAllGTE(stripGovCoin(amount))
	}
}

func (k Keeper) getLiquidatorModuleAccount(ctx sdk.Context) LiquidatorModuleAccount {
	// get store
	store := ctx.KVStore(k.key)
	// get bytes
	bz := store.Get(liquidatorAccountKey)
	if bz == nil {
		return LiquidatorModuleAccount{} // TODO is it safe to do this, or better to initialize the account explicitly
	}
	// unmarshal
	var lma LiquidatorModuleAccount
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &lma)
	return lma
}
func (k Keeper) SetLiquidatorModuleAccount(ctx sdk.Context, lma LiquidatorModuleAccount) {
	store := ctx.KVStore(k.key)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(lma)
	store.Set(liquidatorAccountKey, bz)
}
func stripGovCoin(coins sdk.Coins) sdk.Coins {
	filteredCoins := sdk.NewCoins()
	for _, c := range coins {
		if c.Denom != "kava" {
			filteredCoins = append(filteredCoins, c)
		}
	}
	return filteredCoins
}
