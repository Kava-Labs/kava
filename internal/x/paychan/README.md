# Unidrectional Payment Channels

This module implements simple but feature complete unidirectional payment channels. Channels can be opened by a sender and closed immediately by the receiver, or by the sender subject to a dispute period. There are no top-ups or partial withdrawals (yet). Channels support multiple currencies.

>Note: This is a work in progress. More feature planned. More test cases needed.

# Usage

## Create a channel

	kvcli paychan create --from <your account name> --to <receivers address> --amount 100KVA --chain-id <your chain ID>

## Send an off-chain payment
Send a payment for 10 KVA.

	kvcli paychan pay --from <your account name> --sen-amt 90KVA --rec-amt 10KVA --chan-id <ID of channel> --filename payment.json --chain-id <your chain ID>

Send the file payment.json to your receiver. Then they run the following to verify.

	kvcli paychan verify --filename payment.json

## Close a channel
The receiver can close immediately at any time.

	kvcli paychan submit --from <receiver's account name> --payment payment.json --chain-id <your chain ID>

The sender can close subject to a dispute period during which the receiver can overrule them.

	kvcli paychan submit --from <receiver's account name> --payment payment.json --chain-id <your chain ID>

## Get info on a channel

	kvcli get --chan-id <ID of channel>


# TODOs

 - in code TODOs
 - Tidy up - method descriptions, heading comments, remove uneccessary comments, README/docs
 - Find a better name for Queue - clarify distinction between int slice and abstract queue concept
 - write some sort of integration test
 	- possible bug in submitting same update repeatedly
 - find nicer name for payout
 - add Gas usage
 - add tags (return channel id on creation)
 - refactor cmds to be able to test them, then test them
 	- verify doesn’t throw json parsing error on invalid json
 	- can’t submit an update from an unitialised account
 	- pay without a --from returns confusing error
 - use custom errors instead of using sdk.ErrInternal
 - split off signatures from update as with txs/msgs - testing easier, code easier to use, doesn't store sigs unecessarily on chain
 - consider removing pubKey from UpdateSignature - instead let channel module access accountMapper
 - refactor queue into one object
 - remove printout during tests caused by mock app initialisation
