package modules

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/stretchr/testify/assert"

	"github.com/kava-labs/kava/app"
)

func TestCosmosMigrate_Gov(t *testing.T) {
	// Test the gov state with some proposals so we can test the proto type migration in v40
	originalAppState := `
	{
		"gov": {
			"deposit_params": {
				"max_deposit_period": "600000000000",
				"min_deposit": [{ "amount": "200000000", "denom": "ukava" }]
			},
			"deposits": null,
			"proposals": [
				{
					"content": {
						"type": "cosmos-sdk/TextProposal",
						"value": {
							"description": "test",
							"title": "Nominate Kava Stability Committee"
						}
					},
					"deposit_end_time": "2020-05-27T00:03:48.610420744Z",
					"final_tally_result": {
						"abstain": "388367903902",
						"no": "0",
						"no_with_veto": "0",
						"yes": "45414514436439"
					},
					"id": "5",
					"proposal_status": "Passed",
					"submit_time": "2020-05-13T00:03:48.610420744Z",
					"total_deposit": [
						{
							"amount": "1100000000",
							"denom": "ukava"
						}
					],
					"voting_end_time": "2020-05-30T02:13:16.534185463Z",
					"voting_start_time": "2020-05-16T02:13:16.534185463Z"
				},
				{
					"content": {
						"type": "cosmos-sdk/ParameterChangeProposal",
						"value": {
							"changes": [
								{
									"key": "sendenabled",
									"subspace": "bank",
									"value": "true"
								}
							],
							"description": "Enable transactions",
							"title": "Enable Transactions Param Change"
						}
					},
					"deposit_end_time": "2019-11-30T18:31:15.707527932Z",
					"final_tally_result": {
						"abstain": "0",
						"no": "0",
						"no_with_veto": "0",
						"yes": "56424920427790"
					},
					"id": "1",
					"proposal_status": "Passed",
					"submit_time": "2019-11-16T18:31:15.707527932Z",
					"total_deposit": [
						{
							"amount": "1024601000",
							"denom": "ukava"
						}
					],
					"voting_end_time": "2019-12-03T02:48:16.507225189Z",
					"voting_start_time": "2019-11-19T02:48:16.507225189Z"
				}
			],
			"starting_proposal_id": "1",
			"tally_params": {
				"quorum": "0.334000000000000000",
				"threshold": "0.500000000000000000",
				"veto": "0.334000000000000000"
			},
			"votes": null,
			"voting_params": { "voting_period": "600000000000" }
		}
	}
`
	expected := `
	{
		"gov": {
			"starting_proposal_id": "1",
			"deposits": [],
			"votes": [],
			"proposals": [
				{
					"proposal_id": "5",
					"content": {
						"@type": "/cosmos.gov.v1beta1.TextProposal",
						"title": "Nominate Kava Stability Committee",
						"description": "test"
					},
					"status": "PROPOSAL_STATUS_PASSED",
					"final_tally_result": {
						"yes": "45414514436439",
						"abstain": "388367903902",
						"no": "0",
						"no_with_veto": "0"
					},
					"submit_time": "2020-05-13T00:03:48.610420744Z",
					"deposit_end_time": "2020-05-27T00:03:48.610420744Z",
					"total_deposit": [{ "denom": "ukava", "amount": "1100000000" }],
					"voting_start_time": "2020-05-16T02:13:16.534185463Z",
					"voting_end_time": "2020-05-30T02:13:16.534185463Z"
				},
				{
					"proposal_id": "1",
					"content": {
						"@type": "/cosmos.params.v1beta1.ParameterChangeProposal",
						"title": "Enable Transactions Param Change",
						"description": "Enable transactions",
						"changes": [
							{ "subspace": "bank", "key": "sendenabled", "value": "true" }
						]
					},
					"status": "PROPOSAL_STATUS_PASSED",
					"final_tally_result": {
						"yes": "56424920427790",
						"abstain": "0",
						"no": "0",
						"no_with_veto": "0"
					},
					"submit_time": "2019-11-16T18:31:15.707527932Z",
					"deposit_end_time": "2019-11-30T18:31:15.707527932Z",
					"total_deposit": [{ "denom": "ukava", "amount": "1024601000" }],
					"voting_start_time": "2019-11-19T02:48:16.507225189Z",
					"voting_end_time": "2019-12-03T02:48:16.507225189Z"
				}
			],
			"deposit_params": {
				"min_deposit": [{ "denom": "ukava", "amount": "200000000" }],
				"max_deposit_period": "600s"
			},
			"voting_params": { "voting_period": "600s" },
			"tally_params": {
				"quorum": "0.334000000000000000",
				"threshold": "0.500000000000000000",
				"veto_threshold": "0.334000000000000000"
			}
		}
	}
`
	actual := MustMigrateAppStateJSON(originalAppState)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Auth_Bank(t *testing.T) {
	original := GetTestDataJSON("appstate-v15-auth")
	// expected := GetTestDataJSON("appstate-v16-auth")
	MustMigrateAppStateJSON(original)
	// assert.JSONEq(t, expected, actual)
}

// MustMigrateAppStateJSON migrate v15 app state json to v16
func MustMigrateAppStateJSON(appStateJson string) string {
	var appState genutiltypes.AppMap
	if err := json.Unmarshal([]byte(appStateJson), &appState); err != nil {
		panic(err)
	}
	newGenState := MigrateCosmosAppState(appState, NewClientContext())
	actual, err := json.Marshal(newGenState)
	if err != nil {
		panic(err)
	}
	return string(actual)
}

func GetTestDataJSON(filename string) string {
	file := filepath.Join("testdata", filename+".json")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func NewClientContext() client.Context {
	config := app.MakeEncodingConfig()
	return client.Context{}.
		WithCodec(config.Marshaler).
		WithLegacyAmino(config.Amino)
}
