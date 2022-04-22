package v0_17

import (
	v016auction "github.com/kava-labs/kava/x/auction/legacy/v0_16"
	v017auction "github.com/kava-labs/kava/x/auction/types"
)

func Migrate(oldState v016auction.GenesisState) *v017auction.GenesisState {
	return &v017auction.GenesisState{
		NextAuctionId: oldState.NextAuctionId,
		Params:        migrateParams(oldState.Params),
		Auctions:      oldState.Auctions,
	}
}

func migrateParams(params v016auction.Params) v017auction.Params {
	return v017auction.Params{
		MaxAuctionDuration:  params.MaxAuctionDuration,
		ForwardBidDuration:  v017auction.DefaultForwardBidDuration,
		ReverseBidDuration:  v017auction.DefaultReverseBidDuration,
		IncrementSurplus:    params.IncrementSurplus,
		IncrementDebt:       params.IncrementDebt,
		IncrementCollateral: params.IncrementCollateral,
	}
}
