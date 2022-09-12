package earn

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/types"
)

// InitGenesis initializes genesis state
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	ak types.AccountKeeper,
	gs types.GenesisState,
) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	// Total of all vault share records, vault record total supply should equal this
	vaultTotalShares := types.NewVaultShares()

	for _, vaultShareRecord := range gs.VaultShareRecords {
		if err := vaultShareRecord.Validate(); err != nil {
			panic(fmt.Sprintf("invalid vault share: %s", err))
		}

		vaultTotalShares = vaultTotalShares.Add(vaultShareRecord.Shares...)

		k.SetVaultShareRecord(ctx, vaultShareRecord)
	}

	for _, vaultRecord := range gs.VaultRecords {
		if err := vaultRecord.Validate(); err != nil {
			panic(fmt.Sprintf("invalid vault record: %s", err))
		}

		if !vaultRecord.TotalShares.Amount.Equal(vaultTotalShares.AmountOf(vaultRecord.TotalShares.Denom)) {
			panic(fmt.Sprintf(
				"invalid vault record total supply for %s, got %s but sum of vault shares is %s",
				vaultRecord.TotalShares.Denom,
				vaultRecord.TotalShares.Amount,
				vaultTotalShares.AmountOf(vaultRecord.TotalShares.Denom),
			))
		}

		k.SetVaultRecord(ctx, vaultRecord)
	}

	k.SetParams(ctx, gs.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)
	vaultRecords := k.GetAllVaultRecords(ctx)
	vaultShareRecords := k.GetAllVaultShareRecords(ctx)

	return types.NewGenesisState(params, vaultRecords, vaultShareRecords)
}
