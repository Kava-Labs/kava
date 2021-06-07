package ante

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/kavadist"
)

var v0142UpgradeTime = time.Date(2021, 6, 14, 14, 0, 0, 0, time.UTC)

// UpdateV142MempoolDecorator blocks new tx types from reaching the mempool until after the upgrade time
type UpdateV142MempoolDecorator struct {
}

func NewUpdateV142MempoolDecorator() UpdateV142MempoolDecorator {
	return UpdateV142MempoolDecorator{}
}

func (umd UpdateV142MempoolDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if ctx.BlockTime().After(v0142UpgradeTime) {
		return next(ctx, tx, simulate)
	}
	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		if isUpdateMsgType(msg.Type()) {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "tx not valid until after upgrade")
		}
		proposalMsg, ok := msg.(govtypes.MsgSubmitProposal)
		if ok {
			switch proposalMsg.Content.(type) {
			case kavadist.CommunityPoolMultiSpendProposal:
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "tx not valid until after upgrade")
			default:
				continue
			}
		}
	}
	return next(ctx, tx, simulate)
}

func isUpdateMsgType(msgType string) bool {
	if msgType == "claim_hard_reward_vvesting" || msgType == "claim_usdx_minting_reward_vvesting" {
		return true
	}
	return false
}
