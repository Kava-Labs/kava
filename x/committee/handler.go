package committee

// committee, subcommittee, council, caucus, commission, synod, board
/*
import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/committee/keeper"

	"github.com/kava-labs/kava/x/committee/types"
)

// NewHandler creates an sdk.Handler for committee messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgSubmitProposal:
			handleMsgSubmitProposal(ctx, k, msg)
		case types.MsgVote:
			handleMsgVote(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s msg type: %T", types.ModuleName, msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, k keeper.Keeper, msg types.MsgSubmitProposal) sdk.Result {
	err := keeper.SubmitProposal(ctx, msg)

	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}

func handleMsgVote(ctx sdk.Context, k keeper.Keeper, msg types.MsgVote) sdk.Result {
	err := keeper.AddVote(ctx, msg)

	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
*/
