
Modules test checked:

-- write table with 3 columns

Table:

| Module                                     | Status | Notes                                                                                                                 |
|--------------------------------------------|-------|-----------------------------------------------------------------------------------------------------------------------|
| [auction](x%2Fauction)                     | ✅     |                                                                                                                       |
| [bep3](x%2Fbep3)                           | ✅     |                                                                                                                       |
| [cdp](x%2Fcdp)                             |       | 1 test with begin blocker (some problem with previousAccrualTime 0001-01-01 00:00:00 +0000 UTC true, should be false) |
| [committee](x%2Fcommittee)                 | ✅     |                                                                                                                       |
| [community](x%2Fcommunity)                 | ✅     |                                                                                                                       |
| [earn](x%2Fearn)                           | ✅     |                                                                                                                       |
| [evmutil](x%2Fevmutil)                     |       |                                                                                                                       |
| [hard](x%2Fhard)                           | ✅     |                                                                                                                       |
| [incentive](x%2Fincentive)                 |       | Many test related to delegator sync, some problematic part with extra total bond                                      |
| [issuance](x%2Fissuance)                   | ✅     |                                                                                                                       |
| [kavadist](x%2Fkavadist)                   |       | 2 test at least failed with calculation                                                                               |
| [liquid](x%2Fliquid)                       | ✅     |                                                                                                                       |
| [metrics](x%2Fmetrics)                     | ✅     |                                                                                                                       |
| [precisebank](x%2Fprecisebank)             | ✅      |                                                                                                                       |
| [pricefeed](x%2Fpricefeed)                 | ✅     |                                                                                                                       |
| [router](x%2Frouter)                       | ✅     |                                                                                                                       |
| [savings](x%2Fsavings)                     | ✅      |                                                                                                                       |
| [swap](x%2Fswap)                           |       | many errors with 1000 multiplication difference                                                                       |
| [validator-vesting](x%2Fvalidator-vesting) | ✅     |                                                                                                                       |

