package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params"
	// "github.com/kava-labs/kava/x/bep3/types"
)

// Parameter keys
var (
	KeyMinimumLockTime = []byte("MinimumLockTime")
	KeyChainParams     = []byte("Chains")
	// TODO: validate this time as reasonable
	DefaultHTLTStartingID                = uint64(1)
	DefaultRelayer                       = sdk.AccAddress{}
	DefaultMinimumLockTime time.Duration = 24 * time.Hour
	DefaultChainParams                   = ChainParams{}
)

// Params governance parameters for bep3 module
type Params struct {
	Relayer         sdk.AccAddress `json:"relayer" yaml:"relayer"`
	MinimumLockTime time.Duration  `json:"minimum_lock_time" yaml:"minimum_lock_time"`
	ChainParams     ChainParams    `json:"chain_params" yaml:"chain_params"`
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Relayer: %s
	Minimum Lock Time: %s
	Chains: %s`,
		p.Relayer, p.MinimumLockTime, p.ChainParams)
}

// NewParams returns a new params object
func NewParams(relayer sdk.AccAddress, minimumLockTime time.Duration, chainParams ChainParams) Params {
	return Params{
		Relayer:         relayer,
		MinimumLockTime: minimumLockTime,
		ChainParams:     chainParams,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	return NewParams(DefaultRelayer, DefaultMinimumLockTime, DefaultChainParams)
}

// ChainParam governance parameters for each chain within the bep3 module
type ChainParam struct {
	Name   string      `json:"name" yaml:"name"`       // name of blockchain
	RPCURL string      `json:"rpc_url" yaml:"rpc_url"` // rpc url that relayer uses to subscribe to txs
	Active bool        `json:"active" yaml:"active"`
	Assets AssetParams `json:"chain_assets" yaml:"chain_assets"` // list of supported assets
}

// String implements fmt.Stringer
func (cp ChainParam) String() string {
	return fmt.Sprintf(`Chain:
	Name: %s
	RPC-URL: %s
	Active: %s
	Assets: %s`,
		cp.Name, cp.RPCURL, cp.Active, cp.Assets)
}

// ChainParams array of ChainParam
type ChainParams []ChainParam

// String implements fmt.Stringer
func (cps ChainParams) String() string {
	out := "Chain Params\n"
	for _, cp := range cps {
		out += fmt.Sprintf("%s\n", cp)
	}
	return out
}

// AssetParam governance parameters for each asset within a supported chain
type AssetParam struct {
	Name       string  `json:"name" yaml:"name"`             // name of asset
	CoinID     string  `json:"coin_id" yaml:"coin_id"`       // internationally recognized coin ID
	LimitUSDX  sdk.Int `json:"limit_usdx" yaml:"limit_usdx"` // asset limit in usdx
	Active     bool    `json:"active" yaml:"active"`
	AboveLimit bool    `json:"above_limit" yaml:"above_limit"`
}

// String implements fmt.Stringer
func (ap AssetParam) String() string {
	return fmt.Sprintf(`Asset:
	Name: %s
	Coin ID: %s
	Limit USDX: %s
	Active: %s
	Above Limit: %s`,
		ap.Name, ap.CoinID, ap.LimitUSDX, ap.Active, ap.AboveLimit)
}

// AssetParams array of AssetParam
type AssetParams []AssetParam

// String implements fmt.Stringer
func (aps AssetParams) String() string {
	out := "Asset Params\n"
	for _, ap := range aps {
		out += fmt.Sprintf("%s\n", ap)
	}
	return out
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of bep3 module's parameters.
// nolint
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyMinimumLockTime, Value: &p.MinimumLockTime},
		{Key: KeyChainParams, Value: &p.ChainParams},
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	if p.Relayer.Empty() {
		return sdk.ErrInternal("relayer address cannot be empty")
	}
	if p.MinimumLockTime <= 0 {
		return sdk.ErrInternal("minimum lock time must be greater than 0")
	}
	// TODO: validate ChainParams
	return nil
}
