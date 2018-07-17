Payment channel implementation sketch

Simplifications:

 - unidirectional paychans
 - no top ups or partial withdrawals (only opening and closing)
 - no protection against fund lock up from dissapearing receiver


 TODO
  - fix issue with multisig accounts and sequence numbers
  - create a nicer paychan store key for querying (and implement query)
  - expand client code
  - tidy up - add return tags
  - start removing simplifications, refactor
