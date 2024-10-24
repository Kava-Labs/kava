
Modules test checked:

-- write table with 3 columns

Table:

| Module                                      | Status | Notes                                                                                                                                                                                                                |
|---------------------------------------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| auction                      | ✅     |                                                                                                                                                                                                                      |
| bep3                            | ✅     |                                                                                                                                                                                                                      |
| cdp                              |       | 2 test for some reason keeps the data after init setup (old version didn't contain that data). Because of that it 1) doesn't calculate correctly 2) causes seize                                                     |
| committee                  | ✅     |                                                                                                                                                                                                                      |
| community                  | ✅     |                                                                                                                                                                                                                      |
| earn                            | ✅     |                                                                                                                                                                                                                      |
| evmutil                      |       | Problem with seting up environment, as setup account causes index error (looks like wrong context passed)                                                                                                            |
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
| swap](x%2Fswap)                            | ✅      |                                                                                                                                                                                                                      |
| validator-vesting](x%2Fvalidator-vesting)  | ✅     |                                                                                                                                                                                                                      |

