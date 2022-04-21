package v0_17

import (
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	v016auction "github.com/kava-labs/kava/x/auction/legacy/v0_16"
	v017auction "github.com/kava-labs/kava/x/auction/types"
)

func migrateAuction(auction *codectypes.Any) *codectypes.Any {
	auction.TypeUrl = strings.Replace(auction.TypeUrl, "v1beta1", "v1beta2",-1)
	return auction
}

func migrateAuctions(auctions []*codectypes.Any) []*codectypes.Any {
	anyAuctions := make([]*codectypes.Any, len(auctions))
	for i, auction := range auctions {
		anyAuctions[i] = migrateAuction(auction)
	}
	return anyAuctions
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

func Migrate(oldState v016auction.GenesisState) *v017auction.GenesisState {
	return &v017auction.GenesisState{
		NextAuctionId: oldState.NextAuctionId,
		Params:        migrateParams(oldState.Params),
		Auctions:      migrateAuctions(oldState.Auctions),
	}
}
