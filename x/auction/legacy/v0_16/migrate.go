package v0_16

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	v015auction "github.com/kava-labs/kava/x/auction/legacy/v0_15"
	v016auction "github.com/kava-labs/kava/x/auction/types"
)

func migrateBaseAuction(auction v015auction.BaseAuction) v016auction.BaseAuction {
	return v016auction.BaseAuction{
		ID:              auction.ID,
		Initiator:       auction.GetInitiator(),
		Lot:             auction.GetLot(),
		Bidder:          auction.GetBidder(),
		Bid:             auction.GetBid(),
		HasReceivedBids: auction.GetHasReceivedBids(),
		EndTime:         auction.GetEndTime(),
		MaxEndTime:      auction.GetMaxEndTime(),
	}
}

func migrateAuction(auction v015auction.Auction) *codectypes.Any {
	var protoAuction v016auction.GenesisAuction

	switch auction := auction.(type) {
	case v015auction.SurplusAuction:
		{
			protoAuction = &v016auction.SurplusAuction{
				BaseAuction: migrateBaseAuction(auction.BaseAuction),
			}
		}
	case v015auction.DebtAuction:
		{
			protoAuction = &v016auction.DebtAuction{
				BaseAuction:       migrateBaseAuction(auction.BaseAuction),
				CorrespondingDebt: auction.CorrespondingDebt,
			}
		}
	case v015auction.CollateralAuction:
		{
			lotReturns := v016auction.WeightedAddresses{
				Addresses: auction.LotReturns.Addresses,
				Weights:   auction.LotReturns.Weights,
			}
			if err := lotReturns.Validate(); err != nil {
				panic(err)
			}
			protoAuction = &v016auction.CollateralAuction{
				BaseAuction:       migrateBaseAuction(auction.BaseAuction),
				CorrespondingDebt: auction.CorrespondingDebt,
				MaxBid:            auction.MaxBid,
				LotReturns:        lotReturns,
			}
		}
	default:
		panic(fmt.Errorf("'%s' is not a valid auction", auction))
	}

	// Convert the content into Any.
	contentAny, err := codectypes.NewAnyWithValue(protoAuction)
	if err != nil {
		panic(err)
	}

	return contentAny
}

func migrateAuctions(auctions v015auction.GenesisAuctions) []*codectypes.Any {
	anyAuctions := make([]*codectypes.Any, len(auctions))
	for i, auction := range auctions {
		anyAuctions[i] = migrateAuction(auction)
	}
	return anyAuctions
}

func migrateParams(params v015auction.Params) v016auction.Params {
	return v016auction.Params{
		MaxAuctionDuration:  params.MaxAuctionDuration,
		BidDuration:         params.BidDuration,
		IncrementSurplus:    params.IncrementSurplus,
		IncrementDebt:       params.IncrementDebt,
		IncrementCollateral: params.IncrementCollateral,
	}
}

func Migrate(oldState v015auction.GenesisState) *v016auction.GenesisState {
	return &v016auction.GenesisState{
		NextAuctionId: oldState.NextAuctionID,
		Params:        migrateParams(oldState.Params),
		Auctions:      migrateAuctions(oldState.Auctions),
	}
}
