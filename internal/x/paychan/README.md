Payment channel implementation sketch

Simplifications:

 - unidirectional paychans
 - no top ups or partial withdrawals (only opening and closing)


 TODO
 - chnge module name to "channel"?
 - Find a better name for Queue - clarify distinction between int slice and abstract queue concept
 - refactor queue into one object
 - Do all the small functions need to be methods on the keeper or can they just be floating around?
 - Tidy up - standardise var names, method descriptions, heading comments
 - any problem in signing your own address?
 - Gas
 - find nicer name for payout
 - tags - return channel id
 - create custom errors instead of using sdk.ErrInternal
 - maybe split off signatures from update as with txs/msgs - testing easier, code easier to use, doesn't store sigs unecessarily on chain
 - consider removing pubKey from UpdateSignature - instead let channel module access accountMapper
 - remove printout from tests when app initialised
