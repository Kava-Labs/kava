package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto"
	"testing"
)

func TestEndBlocker(t *testing.T) {
	// SETUP
	ctx, _, channelKeeper, addrs, _ := createMockApp()
	sender := addrs[0]
	receiver := addrs[1]
	coins := sdk.Coins{sdk.NewCoin("KVA", 10)}

	// create new channel
	channelID := ChannelID(0) // should be 0 as first channel
	channel := Channel{
		ID:           channelID,
		Participants: [2]sdk.AccAddress{sender, receiver},
		Coins:        coins,
	}
	channelKeeper.setChannel(ctx, channel)

	// create closing update and submittedUpdate
	payouts := Payouts{
		{sender, sdk.Coins{sdk.NewCoin("KVA", 3)}},
		{receiver, sdk.Coins{sdk.NewCoin("KVA", 7)}},
	}
	update := Update{
		ChannelID: channelID,
		Payouts:   payouts,
		Sequence:  0,
		Sigs:      [1]crypto.Signature{},
	}
	sUpdate := SubmittedUpdate{
		Update:        update,
		ExecutionTime: 0, // current blocktime
	}
	// Set empty submittedUpdatesQueue TODO work out proper genesis initialisation
	channelKeeper.setSubmittedUpdatesQueue(ctx, SubmittedUpdatesQueue{})
	// flag channel for closure
	channelKeeper.addToSubmittedUpdatesQueue(ctx, sUpdate)

	// ACTION
	EndBlocker(ctx, channelKeeper)

	// CHECK RESULTS
	// ideally just check if keeper.channelClose was called, but can't
	// writing endBlocker to accept an interface of which keeper is implementation would make this possible
	// check channel is gone
	_, found := channelKeeper.getChannel(ctx, channelID)
	assert.False(t, found)
	// check queue is empty, NOTE: due to encoding, an empty queue (underneath just an int slice) will be decoded as nil slice rather than an empty slice
	suq, _ := channelKeeper.getSubmittedUpdatesQueue(ctx)
	assert.Equal(t, SubmittedUpdatesQueue(nil), suq)
	// check submittedUpdate is gone
	_, found = channelKeeper.getSubmittedUpdate(ctx, channelID)
	assert.False(t, found)
}
