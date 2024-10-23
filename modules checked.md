
Modules test checked:

-- write table with 3 columns

Table:

| Module                                      | Status | Notes                                                                                                                                                                                                                |
|---------------------------------------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| auction                      | ✅     |                                                                                                                                                                                                                      |
| bep3                            | ✅     |                                                                                                                                                                                                                      |
| cdp                              |       | 1 test with begin blocker (some problem with previousAccrualTime 0001-01-01 00:00:00 +0000 UTC true, should be false)                                                                                                |
| committee                  | ✅     |                                                                                                                                                                                                                      |
| community                  | ✅     |                                                                                                                                                                                                                      |
| earn                            | ✅     |                                                                                                                                                                                                                      |
| evmutil                      |       |                                                                                                                                                                                                                      |
| hard                            | ✅     |                                                                                                                                                                                                                      |
| incentive                  |       | Many test related to delegator sync, some problematic part with extra total bond                                                                                                                                     |
| issuance                    | ✅     |                                                                                                                                                                                                                      |
| kavadist                    |       | 2 test at least failed with calculation, problem with ButnCoins GetAll banances (it is different from expected, not enought) inside func (tApp TestApp) DeleteGenesisValidatorCoins(t *testing.T, ctx sdk.Context) { |
| liquid                        | ✅     |                                                                                                                                                                                                                      |
| metrics                      | ✅     |                                                                                                                                                                                                                      |
| precisebank             | ✅      |                                                                                                                                                                                                                      |
| pricefeed                  | ✅     |                                                                                                                                                                                                                      |
| router                        | ✅     |                                                                                                                                                                                                                      |
| savings                      | ✅      |                                                                                                                                                                                                                      |
| swap](x%2Fswap)                            |       | many errors with 1000 multiplication difference                                                                                                                                                                      |
| validator-vesting](x%2Fvalidator-vesting)  | ✅     |                                                                                                                                                                                                                      |

