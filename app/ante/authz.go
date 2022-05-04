package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// AuthzLimiterDecorator blocks certain msg types from being granted or executed within authz.
type AuthzLimiterDecorator struct {
	// disabledMsgTypes is the type urls of the msgs to block.
	disabledMsgTypes []string
}

// NewAuthzLimiterDecorator creates a decorator to block certain msg types from being granted or executed within authz.
func NewAuthzLimiterDecorator(disabledMsgTypes ...string) AuthzLimiterDecorator {
	return AuthzLimiterDecorator{
		disabledMsgTypes: disabledMsgTypes,
	}
}

func (ald AuthzLimiterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msgI := range tx.GetMsgs() {
		switch msg := msgI.(type) {
		case *authz.MsgGrant:
			authorization := msg.GetAuthorization()
			for _, typeUrl := range ald.disabledMsgTypes {
				if authorization.MsgTypeURL() == typeUrl {
					return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot authz grant msg type %s", typeUrl)
				}
			}
		case *authz.MsgExec:
			innerMsgs, err := msg.GetMessages()
			if err != nil {
				return ctx, err
			}
			for _, innerMsg := range innerMsgs {
				for _, typeUrl := range ald.disabledMsgTypes {
					if sdk.MsgTypeURL(innerMsg) == typeUrl {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot authz exec msg type %s", typeUrl)
					}
				}
			}
		}
	}
	return next(ctx, tx, simulate)
}
