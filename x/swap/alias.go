package swap

import (
	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/types"
)

const (
	ModuleName        = types.ModuleName
	QuerierRoute      = types.QuerierRoute
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	DefaultParamspace = types.DefaultParamspace
)

type (
	GenesisState = types.GenesisState
	Keeper       = keeper.Keeper
)

var (
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	ModuleCdc           = types.ModuleCdc
	ParamKeyTable       = types.ParamKeyTable
	RegisterCodec       = types.RegisterCodec
	DefaultGenesisState = types.DefaultGenesisState
)
