Payment channel implementation sketch

Simplifications:

 - unidirectional paychans
 - no top ups or partial withdrawals (only opening and closing)


 TODO
 - chnge module name to "channel"?
 - Find a better name for Queue - clarify distinction between int slice and abstract queue concept
 - Do all the small functions need to be methods on the keeper or can they just be floating around?
 - Tidy up - standardise var names, comments and method descriptions
 - is having all the get functions return a bool if not found reasonable?
 - any problem in signing your own address?
 - Gas
