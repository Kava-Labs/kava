Payment channel implementation sketch

Simplifications:

 - unidirectional paychans
 - no top ups or partial withdrawals (only opening and closing)


 TODO
 - in code TODOs
 - write basic cmds
 - Tidy up - method descriptions, heading comments, remove uneccessary comments, README/docs
 - chnge module name to "channel"?
 - Find a better name for Queue - clarify distinction between int slice and abstract queue concept
 - write some sort of integration test
 - find nicer name for payout
 - add Gas usage
 - add tags (return channel id on creation)
 - use custom errors instead of using sdk.ErrInternal
 - split off signatures from update as with txs/msgs - testing easier, code easier to use, doesn't store sigs unecessarily on chain
 - consider removing pubKey from UpdateSignature - instead let channel module access accountMapper
 - remove printout from tests when app initialised
 - refactor queue into one object
