package bep3

import (
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

const (
	ModuleName             = types.ModuleName
	RouterKey              = types.RouterKey
	StoreKey               = types.StoreKey
	DefaultParamspace      = types.DefaultParamspace
	DefaultCodespace       = types.DefaultCodespace
	QueryGetParams         = types.QueryGetParams
	QueryGetAtomicSwap     = types.QueryGetAtomicSwap
	QueryGetAtomicSwaps    = types.QueryGetAtomicSwaps
	QuerierRoute           = types.QuerierRoute
	AttributeValueCategory = types.AttributeValueCategory
)

var (
	// functions aliases
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	RegisterCodec       = types.RegisterCodec
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	Keeper               = keeper.Keeper
	CodeType             = types.CodeType
	GenesisState         = types.GenesisState
	Params               = types.Params
	MsgCreateAtomicSwap  = types.MsgCreateAtomicSwap
	MsgDepositAtomicSwap = types.MsgDepositAtomicSwap
	MsgRefundAtomicSwap  = types.MsgRefundAtomicSwap
	MsgClaimAtomicSwap   = types.MsgClaimAtomicSwap
)
