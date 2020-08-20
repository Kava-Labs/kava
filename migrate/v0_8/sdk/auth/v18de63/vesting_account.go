package v18de63

import (
	"encoding/json"
	"strconv"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BaseVestingAccount implements the VestingAccount interface. It contains all
// the necessary fields needed for any vesting account implementation.
type BaseVestingAccount struct {
	*BaseAccount

	OriginalVesting  sdk.Coins `json:"original_vesting" yaml:"original_vesting"`   // coins in account upon initialization
	DelegatedFree    sdk.Coins `json:"delegated_free" yaml:"delegated_free"`       // coins that are vested and delegated
	DelegatedVesting sdk.Coins `json:"delegated_vesting" yaml:"delegated_vesting"` // coins that vesting and delegated
	EndTime          int64     `json:"end_time" yaml:"end_time"`                   // when the coins become unlocked
}

// ContinuousVestingAccount implements the VestingAccount interface. It
// continuously vests by unlocking coins linearly with respect to time.
type ContinuousVestingAccount struct {
	*BaseVestingAccount

	StartTime int64 `json:"start_time" yaml:"start_time"` // when the coins start to vest
}

// PeriodicVestingAccount implements the VestingAccount interface. It
// periodically vests by unlocking coins during each specified period
type PeriodicVestingAccount struct {
	*BaseVestingAccount
	StartTime      int64   `json:"start_time" yaml:"start_time"`           // when the coins start to vest
	VestingPeriods Periods `json:"vesting_periods" yaml:"vesting_periods"` // the vesting schedule
}

type prettyPublicKey struct {
	PubKey crypto.PubKey
}

func (p prettyPublicKey) MarshalJSON() ([]byte, error) {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	return cdc.MarshalJSON(p.PubKey)
}

type vestingAccountPretty struct {
	Address          sdk.AccAddress  `json:"address" yaml:"address"`
	Coins            sdk.Coins       `json:"coins" yaml:"coins"`
	PubKey           prettyPublicKey `json:"public_key" yaml:"public_key"`
	AccountNumber    string          `json:"account_number" yaml:"account_number"`
	Sequence         string          `json:"sequence" yaml:"sequence"`
	OriginalVesting  sdk.Coins       `json:"original_vesting" yaml:"original_vesting"`
	DelegatedFree    sdk.Coins       `json:"delegated_free" yaml:"delegated_free"`
	DelegatedVesting sdk.Coins       `json:"delegated_vesting" yaml:"delegated_vesting"`
	EndTime          int64           `json:"end_time" yaml:"end_time"`

	// custom fields based on concrete vesting type which can be omitted
	StartTime      int64   `json:"start_time,omitempty" yaml:"start_time,omitempty"`
	VestingPeriods Periods `json:"vesting_periods,omitempty" yaml:"vesting_periods,omitempty"`
}

func (pva PeriodicVestingAccount) MarshalJSON() ([]byte, error) {
	//pubKey := prettyPublicKey{keyType: "test", value: crypto.AddressHash(pva.PubKey.Bytes())}
	pubKey := prettyPublicKey{PubKey: pva.PubKey}

	alias := vestingAccountPretty{
		Address:          pva.Address,
		Coins:            pva.Coins,
		PubKey:           pubKey,
		AccountNumber:    strconv.FormatUint(pva.AccountNumber, 10),
		Sequence:         strconv.FormatUint(pva.Sequence, 10),
		OriginalVesting:  pva.OriginalVesting,
		DelegatedFree:    pva.DelegatedFree,
		DelegatedVesting: pva.DelegatedVesting,
		EndTime:          pva.EndTime,
		StartTime:        pva.StartTime,
		VestingPeriods:   pva.VestingPeriods,
	}

	return json.Marshal(alias)
}

// DelayedVestingAccount implements the VestingAccount interface. It vests all
// coins after a specific time, but non prior. In other words, it keeps them
// locked until a specified time.
type DelayedVestingAccount struct {
	*BaseVestingAccount
}
