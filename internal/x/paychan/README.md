Payment channel implementation sketch

Simplifications:

 - unidirectional paychans
 - no top ups or partial withdrawals (only opening and closing)


 TODO
 - error handling (getter setter return values? and what happens in failures)
 - chnge module name to "channel"?
 - Find a better name for Queue - clarify distinction between int slice and abstract queue concept
 - Do all the small functions need to be methods on the keeper or can they just be floating around?
