package paychan

import ()

func EndBlocker(ctx sdk.Context k Keeper) sdk.Tags {

	// Iterate through submittedUpdates and for each
	//		if current block height >= executionDate
	//			k.CloseChannel(...)

	tags := sdk.NewTags()
	return tags
}