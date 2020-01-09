<!--
order: 3
-->

# Messages

## Bidding

Users can bid on auctions using the `MsgPlaceBid` message type. All auction types can be bid on using the same message type.

```go
// MsgPlaceBid is the message type used to place a bid on any type of auction.
type MsgPlaceBid struct {
	AuctionID uint64
	Bidder    sdk.AccAddress
	Bid       sdk.Coin
	Lot       sdk.Coin
}
```

**State Modifications:**

* Update bidder if different than previous bidder
* For Surplus auctions:
  * Update Bid Amount
  * Return bid coins to previous bidder
  * Burn coins equal to the increment in the bid (CurrentBid - PreviousBid)
* For Debt auctions:
  * Update lot amount
  * Return bid coins to previous bidder
* For Collateral auctions:
  * Return bid coins to previous bidder
  * If in forward phase:
    * Update bid amount
  * If in reverse phase:
    * Update lot amount
* Extend auction by `BidDuration`, or `MaxEndTime`
