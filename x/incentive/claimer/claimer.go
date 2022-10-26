package claimer

// Similar to Distributor, we could create a Claimer service.
// It would interact with the distributor to withdraw a user's rewards and send them to the user.
// Having a separate service should make it easier to change the claiming logic, eg could have different claimers for claiming vested amounts with multipliers, or simple claiming.

// type Claimer interface {
// 	Claim(ctx sdk.Context, sourceID string, opts types.Selections)
// }

// type SimplerMultiplierClaimer struct {
// 	rewardsWithdrawer rewardsWithdrawer // pass in distributor
// 	// needs access to params
// 	// needs service to send locked funds
// }

// type rewardsWithdrawer interface {
// 	GetUserBalance(sourceID string, owner sdk.AccAddress) sdk.Coins
// 	WithdrawUserBalance(sourceID string, owner sdk.AccAddress, amt sdk.Coins)
// }

// func (smc SimplerMultiplierClaimer) Claim(ctx sdk.Context, sourceID string, opts types.Selections) {
// 	// do normal claim stuff:
// 	// get user balance
// 	// verify selection
// 	// withdraw rewards from distributor

// }
