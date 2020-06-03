<!--
order: 1
-->

# Concepts

Prices can be posted by any account which is added as an oracle. Oracles are specific to each market and can be updated via param change proposals. When an oracle posts a price, they submit a message to the blockchain that contains the current price for that market and a time when that price should be considered expired. If an oracle posts a new price, that price becomes the current price for that oracle, regardless of the previous price's expiry. A group of prices posted by a set of oracles for a particular market are referred to as 'raw prices' and the current median price of all valid oracle prices is referred to as the 'current price'. Each block, the current price for each market is determined by calculating the median of the raw prices.
