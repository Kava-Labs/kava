package v18de63

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/slashing"
)

const ModuleName = "slashing"

// The field types (except Params) are the same between v18de63 and v0.38, so the types can be imported rather than copied in

type GenesisState struct {
	Params       Params                                   `json:"params" yaml:"params"`
	SigningInfos map[string]slashing.ValidatorSigningInfo `json:"signing_infos" yaml:"signing_infos"`
	MissedBlocks map[string][]slashing.MissedBlock        `json:"missed_blocks" yaml:"missed_blocks"`
}

// These sdk types are the same between v18de63 and v0.38

type Params struct {
	MaxEvidenceAge          time.Duration `json:"max_evidence_age" yaml:"max_evidence_age"`
	SignedBlocksWindow      int64         `json:"signed_blocks_window" yaml:"signed_blocks_window"`
	MinSignedPerWindow      sdk.Dec       `json:"min_signed_per_window" yaml:"min_signed_per_window"`
	DowntimeJailDuration    time.Duration `json:"downtime_jail_duration" yaml:"downtime_jail_duration"`
	SlashFractionDoubleSign sdk.Dec       `json:"slash_fraction_double_sign" yaml:"slash_fraction_double_sign"`
	SlashFractionDowntime   sdk.Dec       `json:"slash_fraction_downtime" yaml:"slash_fraction_downtime"`
}
