<!--
order: 1
-->

# Concepts

Auctions are broken down into three distinct types, which correspond to three specific functionalities within the CDP system.

* **Surplus Auction:** An auction in which a fixed lot of coins (c1) is sold for increasing amounts of other coins (c2). Bidders increment the amount of c2 they are willing to pay for the lot of c1. After the completion of a surplus auction, the winning bid of c2 is burned, and the bidder receives the lot of c1. As a concrete example, surplus auction are used to sell a fixed amount of USDX stable coins in exchange for increasing bids of KAVA governance tokens. The governance tokens are then burned and the winner receives USDX.
* **Debt Auction:** An auction in which a fixed amount of coins (c1) is bid for a decreasing lot of other coins (c2). Bidders decrement the lot of c2 they are willing to receive for the fixed amount of c1. As a concrete example, debt auctions are used to raise a certain amount of USDX stable coins in exchange for decreasing lots of KAVA governance tokens. The USDX tokens are used to recapitalize the cdp system and the winner receives KAVA.
* **Surplus Reverse Auction:** Are two phase auction is which a fixed lot of coins (c1) is sold for increasing amounts of other coins (c2). Bidders increment the amount of c2 until a specific `maxBid` is reached. Once `maxBid` is reached, a fixed amount of c2 is bid for a decreasing lot of c1. In the second phase, bidders decrement the lot of c1 they are willing to receive for a fixed amount of c2. As a concrete example, collateral auctions are used to sell collateral (ATOM, for example) for up to a `maxBid` amount of USDX. The USDX tokens are used to recapitalize the cdp system and the winner receives the specified lot of ATOM. In the event that the winning lot is smaller than the total lot, the excess ATOM is ratably returned to the original owners of the liquidated CDPs that were collateralized with that ATOM.

Auctions are always initiated by another module, and not directly by users. Auctions start with an expiry, the time at which the auction is guaranteed to end, even if there have been no bidders. After each bid, the auction is extended by a specific amount of time, `BidDuration`. In the case that increasing the auction time by `BidDuration` would cause the auction to go past its expiry, the expiry is chosen as the ending time.
