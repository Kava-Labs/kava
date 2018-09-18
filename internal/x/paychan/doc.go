/*
Package paychan provides unidirectional payment channels.

This module implements simple but feature complete unidirectional payment channels. Channels can be opened by a sender and closed immediately by the receiver, or by the sender subject to a dispute period. There are no top-ups or partial withdrawals (yet). Channels support multiple currencies.

>Note: This module is still a bit rough around the edges. More feature planned. More test cases needed.


TODO Explain how the payment channels are implemented.

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

*/
package paychan
