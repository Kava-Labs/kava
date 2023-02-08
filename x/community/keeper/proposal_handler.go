package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/community/types"
)

// HandleCommunityPoolLendDepositProposal is a handler for executing a passed community pool lend deposit proposal.
func HandleCommunityPoolLendDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendDepositProposal) error {
	return k.hardKeeper.Deposit(ctx, k.moduleAddress, p.Amount)
}

// HandleCommunityPoolLendWithdrawProposal is a handler for executing a passed community pool lend withdraw proposal.
func HandleCommunityPoolLendWithdrawProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendWithdrawProposal) error {
	return k.hardKeeper.Withdraw(ctx, k.moduleAddress, p.Amount)
}

// HandleCommunityPoolProposal is a handler for executing a passed community pool proposal
func HandleCommunityPoolProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolProposal) error {
	logger := k.Logger(ctx)

	// attempt to execute all messages within the passed proposal
	// Messages may mutate state thus we use a cached context. If one of
	// the handlers fails, no state mutation is written and the error
	// message is logged.
	cacheCtx, writeCache := ctx.CacheContext()
	msgs, err := p.GetMsgs()
	if err != nil {
		return err
	}

	enabledUrls := k.GetParams(ctx).EnabledProposalMsgUrls
	for idx, msg := range msgs {
		if !isMsgEnabled(msg, enabledUrls) {
			return logAndReturnProposalMsgError(logger, idx, msg, "msg not enabled via params", types.ErrProposalMsgNotEnabledErr)
		}

		// assert that the community module account is the only signer of the messages
		signers := msg.GetSigners()
		if len(signers) != 1 || !signers[0].Equals(k.moduleAddress) {
			return logAndReturnProposalMsgError(logger, idx, msg, "invalid signer", types.ErrProposalSigningErr)
		}

		handler := k.Router().Handler(msg)
		_, err := handler(cacheCtx, msg)

		// fail proposal if any message failed to execute
		if err != nil {
			return logAndReturnProposalMsgError(logger, idx, msg, err.Error(), types.ErrProposalExecutionErr)
		}
	}

	// write state to the underlying multi-store
	writeCache()

	return nil
}

func isMsgEnabled(msg sdk.Msg, enabledUrls []string) bool {
	for _, enabledMsgUrl := range enabledUrls {
		if sdk.MsgTypeURL(msg) == enabledMsgUrl {
			return true
		}
	}
	return false
}

func logAndReturnProposalMsgError(logger log.Logger, msgIndex int, msg sdk.Msg, errMsg string, err *sdkerrors.Error) error {
	errStr := fmt.Sprintf(
		"CommunityPoolProposal msg %d (%s) failed on execution: %s",
		msgIndex, sdk.MsgTypeURL(msg), errMsg,
	)
	logger.Info(errStr)
	return sdkerrors.Wrap(err, errStr)
}
