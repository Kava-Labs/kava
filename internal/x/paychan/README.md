Payment channel implementation sketch

Simplifications:

 - unidirectional paychans
 - no top ups or partial withdrawals (only opening and closing)


 TODO
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
 - remove printout from tests when app initialised
 - refactor queue into one object
