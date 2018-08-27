package paychan

import ()

func EndBlocker(ctx sdk.Context, k Keeper) sdk.Tags {

	// Iterate through submittedUpdatesQueue
	// TODO optimise so it doesn't pull every update from DB every block
	var sUpdate SubmittedUpdate
	q := k.getSubmittedUpdatesQueue(ctx)
	for _, id := range q {
		// close the channel if the update has reached its execution time.
		// Using >= in case some are somehow missed.
		sUpdate = k.getSubmittedUpdate(ctx, id)
		if ctx.BlockHeight() >= sUpdate.ExecutionTime {
			k.closeChannel(ctx, sUpdate.Update)
		}
	}

	tags := sdk.NewTags()
	return tags
}
