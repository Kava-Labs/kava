package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgClaimReward = "op_weight_msg_claim_reward"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak auth.AccountKeeper, sk types.SupplyKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgClaimReward int

	appParams.GetOrGenerate(cdc, OpWeightMsgClaimReward, &weightMsgClaimReward, nil,
		func(_ *rand.Rand) {
			weightMsgClaimReward = appparams.DefaultWeightMsgClaimReward
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgClaimReward,
			SimulateMsgClaimReward(ak, sk, k),
		),
	}
}

// SimulateMsgClaimReward generates a MsgClaimReward
func SimulateMsgClaimReward(ak auth.AccountKeeper, sk types.SupplyKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		return simulation.NewOperationMsgBasic(types.ModuleName,
			"no-operation (no accounts currently have fulfillable claims)", "", false, nil), nil, nil
	}
}
