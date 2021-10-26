package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
)

func init() {
	kavaConfig := sdk.GetConfig()
	app.SetPrefixes(kavaConfig)
	kavaConfig.Seal()
}
