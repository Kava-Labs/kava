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
	QueryGetHTLT           = types.QueryGetHTLT
	QueryGetHTLTs          = types.QueryGetHTLTs
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
	Keeper         = keeper.Keeper
	CodeType       = types.CodeType
	GenesisState   = types.GenesisState
	Params         = types.Params
	MsgCreateHTLT  = types.MsgCreateHTLT
	MsgDepositHTLT = types.MsgDepositHTLT
	MsgRefundHTLT  = types.MsgRefundHTLT
	MsgClaimHTLT   = types.MsgClaimHTLT
)
