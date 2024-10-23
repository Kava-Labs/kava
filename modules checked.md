
Modules test checked:

-- write table with 3 columns

Table:

| Module                                     | Status | Notes                                                                                                                                                                |
|--------------------------------------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [auction](x%2Fauction)                     | ✅     |                                                                                                                                                                      |
| [bep3](x%2Fbep3)                           | ✅     |                                                                                                                                                                      |
| [cdp](x%2Fcdp)                             |       | 1 test with begin blocker (some problem with previousAccrualTime 0001-01-01 00:00:00 +0000 UTC true, should be false)                                                |
| [committee](x%2Fcommittee)                 |      | 1 test failed (upgrade proposal)                                                                                                                                     |
| [community](x%2Fcommunity)                 | ✅     |                                                                                                                                                                      |
| [earn](x%2Fearn)                           | ✅     |                                                                                                                                                                      |
| [evmutil](x%2Fevmutil)                     |      |                                                                                                                                                                      |
| [hard](x%2Fhard)                           | ✅     |                                                                                                                                                                      |
| [incentive](x%2Fincentive)                 |      | Many test related to delegator sync, some problematic part with extra total bond                                                                                     |
| [issuance](x%2Fissuance)                   | ✅     |                                                                                                                                                                      |
| [kavadist](x%2Fkavadist)                   |      | 2 test at least failed with calculation                                                                                                                              |
| [liquid](x%2Fliquid)                       |      | lots of test failed                                                                                                                                                  |
| [metrics](x%2Fmetrics)                     | ✅     |                                                                                                                                                                      |
| [precisebank](x%2Fprecisebank)             |      | ~ 6 test failed                                                                                                                                                      |
| [pricefeed](x%2Fpricefeed)                 | ✅     |                                                                                                                                                                      |
| [router](x%2Frouter)                       |      | 3 test failed                                                                                                                                                        |
| [savings](x%2Fsavings)                     |      | 1 test failed: "conflict: index uniqueness constrain violation: 2"   TestGrpcQueryTestSuite/TestGrpcQueryTotalSupply/aggregates_bkava_denoms,_accounting_for_slashing |
| [swap](x%2Fswap)                           |      | many errors with 1000 multiplication difference                                                                                                                      |
| [validator-vesting](x%2Fvalidator-vesting) | ✅     |                                                                                                                                                                      |

