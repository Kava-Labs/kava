package v0_8

import (
	"encoding/json"
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)
	config.Seal()

	os.Exit(m.Run())
}

func TestMigrate_Auth_BaseAccount(t *testing.T) {
	oldGenesisState := genutil.AppMap{
		"auth": []byte(`{
	"params": {
		"max_memo_characters": "256",
		"sig_verify_cost_ed25519": "590",
		"sig_verify_cost_secp256k1": "1000",
		"tx_sig_limit": "7",
		"tx_size_cost_per_byte": "10"
	},
	"accounts": [
		{
			"type": "cosmos-sdk/Account",
			"value": {
			  "account_number": "4589",
			  "address": "kava1qqfzmtucfc2ky6qm2yysypvehay0jytjp87czf",
			  "coins": [
				{
				  "amount": "2769",
				  "denom": "ukava"
				}
			  ],
			  "public_key": null,
			  "sequence": "0"
			}
		}
	]
}`),
	}

	expectedAuthGenState := genutil.AppMap{
		"auth": json.RawMessage(`{
			"params": {
				"max_memo_characters": "256",
				"sig_verify_cost_ed25519": "590",
				"sig_verify_cost_secp256k1": "1000",
				"tx_sig_limit": "7",
				"tx_size_cost_per_byte": "10"
			},
			"accounts": [
				{
					"type": "cosmos-sdk/Account",
					"value": {
						"account_number": 4589,
						"address": "kava1qqfzmtucfc2ky6qm2yysypvehay0jytjp87czf",
						"coins": [
							{
								"amount": "2769",
								"denom": "ukava"
							}
						],
						"public_key": "",
						"sequence": 0
					}
				}
			]
		}`),
	}

	newGenesisState := Migrate(oldGenesisState)
	require.JSONEq(t, string(expectedAuthGenState["auth"]), string(newGenesisState["auth"]))
}

func TestMigrate_Auth_ValidatorVestingAccount(t *testing.T) {
	oldGenesisState := genutil.AppMap{
		"auth": []byte(`{
	"params": {
		"max_memo_characters": "256",
		"sig_verify_cost_ed25519": "590",
		"sig_verify_cost_secp256k1": "1000",
		"tx_sig_limit": "7",
		"tx_size_cost_per_byte": "10"
	},
	"accounts": [
		  {
			"type": "cosmos-sdk/ValidatorVestingAccount",
			"value": {
			  "PeriodicVestingAccount": {
				"BaseVestingAccount": {
				  "BaseAccount": {
					"account_number": "104",
					"address": "kava1pjm84k90qnmcexv6704cxe243j52vww572j78u",
					"coins": [
					  {
						"amount": "410694803",
						"denom": "ukava"
					  }
					],
					"public_key": {
					  "type": "tendermint/PubKeySecp256k1",
					  "value": "A1dLIMH2gbFq6WhsnOc0aXicwjXva/8QZLQxeLcUxGTk"
					},
					"sequence": "10"
				  },
				  "delegated_free": [],
				  "delegated_vesting": [
					{
					  "amount": "699980000000",
					  "denom": "ukava"
					}
				  ],
				  "end_time": "1636120800",
				  "original_vesting": [
					{
					  "amount": "699990000000",
					  "denom": "ukava"
					}
				  ]
				},
				"start_time": "1572962400",
				"vesting_periods": [
				  {
					"amount": [
					  {
						"amount": "349995000000",
						"denom": "ukava"
					  }
					],
					"length": "15724800"
				  },
				  {
					"amount": [
					  {
						"amount": "58332500000",
						"denom": "ukava"
					  }
					],
					"length": "7948800"
				  },
				  {
					"amount": [
					  {
						"amount": "58332500000",
						"denom": "ukava"
					  }
					],
					"length": "7948800"
				  },
				  {
					"amount": [
					  {
						"amount": "58332500000",
						"denom": "ukava"
					  }
					],
					"length": "7948800"
				  },
				  {
					"amount": [
					  {
						"amount": "58332500000",
						"denom": "ukava"
					  }
					],
					"length": "7689600"
				  },
				  {
					"amount": [
					  {
						"amount": "58332500000",
						"denom": "ukava"
					  }
					],
					"length": "7948800"
				  },
				  {
					"amount": [
					  {
						"amount": "58332500000",
						"denom": "ukava"
					  }
					],
					"length": "7948800"
				  }
				]
			  },
			  "current_period_progress": {
				"missed_blocks": "9",
				"total_blocks": "190565"
			  },
			  "debt_after_failed_vesting": [],
			  "return_address": "kava1qvsus5qg8yhre7k2c78xkkw4nvqqgev7ezrja8",
			  "signing_threshold": "90",
			  "validator_address": "kavavalcons1rcgcrswwvunnfrx73ksc5ks8t9jtcnpaehf726",
			  "vesting_period_progress": [
				{
				  "period_complete": true,
				  "vesting_successful": true
				},
				{
				  "period_complete": false,
				  "vesting_successful": false
				},
				{
				  "period_complete": false,
				  "vesting_successful": false
				},
				{
				  "period_complete": false,
				  "vesting_successful": false
				},
				{
				  "period_complete": false,
				  "vesting_successful": false
				},
				{
				  "period_complete": false,
				  "vesting_successful": false
				},
				{
				  "period_complete": false,
				  "vesting_successful": false
				}
			  ]
			}
		  }
	]
}`),
	}

	expectedAuthGenState := genutil.AppMap{
		"auth": json.RawMessage(`{
			"params": {
				"max_memo_characters": "256",
				"sig_verify_cost_ed25519": "590",
				"sig_verify_cost_secp256k1": "1000",
				"tx_sig_limit": "7",
				"tx_size_cost_per_byte": "10"
			},
			"accounts": [
				{
					"type": "cosmos-sdk/ValidatorVestingAccount",
					"value": {
						"address": "kava1pjm84k90qnmcexv6704cxe243j52vww572j78u",
						"coins": [
							{
								"amount": "410694803",
								"denom": "ukava"
							}
						],
						"public_key": "kavapub1addwnpepqdt5kgxp76qmz6hfdpkfeee5d9ufes34aa4l7yryksch3dc5c3jwgdh2lju",
						"account_number": 104,
						"sequence": 10,
						"delegated_free": [],
						"delegated_vesting": [
							{
								"amount": "699980000000",
								"denom": "ukava"
							}
						],
						"end_time": 1636120800,
						"original_vesting": [
							{
								"amount": "699990000000",
								"denom": "ukava"
							}
						],
						"start_time": 1572962400,
						"vesting_periods": [
							{
								"amount": [
									{
										"amount": "349995000000",
										"denom": "ukava"
									}
								],
								"length": 15724800
							},
							{
								"amount": [
									{
										"amount": "58332500000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "58332500000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "58332500000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "58332500000",
										"denom": "ukava"
									}
								],
								"length": 7689600
							},
							{
								"amount": [
									{
										"amount": "58332500000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "58332500000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							}
						],
						"current_period_progress": {
							"missed_blocks": 9,
							"total_blocks": 190565
						},
						"debt_after_failed_vesting": [],
						"return_address": "kava1qvsus5qg8yhre7k2c78xkkw4nvqqgev7ezrja8",
						"signing_threshold": "90",
						"validator_address": "kavavalcons1rcgcrswwvunnfrx73ksc5ks8t9jtcnpaehf726",
						"vesting_period_progress": [
							{
								"period_complete": true,
								"vesting_successful": true
							},
							{
								"period_complete": false,
								"vesting_successful": false
							},
							{
								"period_complete": false,
								"vesting_successful": false
							},
							{
								"period_complete": false,
								"vesting_successful": false
							},
							{
								"period_complete": false,
								"vesting_successful": false
							},
							{
								"period_complete": false,
								"vesting_successful": false
							},
							{
								"period_complete": false,
								"vesting_successful": false
							}
						]
				}
			]
		}`),
	}

	newGenesisState := Migrate(oldGenesisState)

	t.Log(string(newGenesisState["auth"]))

	require.JSONEq(t, string(expectedAuthGenState["auth"]), string(newGenesisState["auth"]))
}

func TestMigrate_Auth_ModuleAccount(t *testing.T) {
	oldGenesisState := genutil.AppMap{
		"auth": []byte(`{
	"params": {
		"max_memo_characters": "256",
		"sig_verify_cost_ed25519": "590",
		"sig_verify_cost_secp256k1": "1000",
		"tx_sig_limit": "7",
		"tx_size_cost_per_byte": "10"
	},
	"accounts": [
		{
			"type": "cosmos-sdk/ModuleAccount",
			"value": {
			  "BaseAccount": {
				"account_number": "168",
				"address": "kava1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3fwaj0s",
				"coins": [
				  {
					"amount": "87921781313382",
					"denom": "ukava"
				  }
				],
				"public_key": null,
				"sequence": "0"
			  },
			  "name": "bonded_tokens_pool",
			  "permissions": [
				"burner",
				"staking"
			  ]
			}
		}
	]
}`),
	}

	expectedAuthGenState := genutil.AppMap{
		"auth": json.RawMessage(`{
			"params": {
				"max_memo_characters": "256",
				"sig_verify_cost_ed25519": "590",
				"sig_verify_cost_secp256k1": "1000",
				"tx_sig_limit": "7",
				"tx_size_cost_per_byte": "10"
			},
			"accounts": [
				{
					"type": "cosmos-sdk/ModuleAccount",
					"value": {
						"account_number": 168,
						"address": "kava1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3fwaj0s",
						"coins": [
						  {
							"amount": "87921781313382",
							"denom": "ukava"
						  }
						],
						"public_key": "",
						"sequence": 0,
						"name": "bonded_tokens_pool",
						"permissions": [
							"burner",
							"staking"
						]
					}
				}
			]
		}`),
	}

	newGenesisState := Migrate(oldGenesisState)
	require.JSONEq(t, string(expectedAuthGenState["auth"]), string(newGenesisState["auth"]))
}

func TestMigrate_Auth_PeriodicVestingAccount(t *testing.T) {
	oldGenesisState := genutil.AppMap{
		"auth": []byte(`{
	"params": {
		"max_memo_characters": "256",
		"sig_verify_cost_ed25519": "590",
		"sig_verify_cost_secp256k1": "1000",
		"tx_sig_limit": "7",
		"tx_size_cost_per_byte": "10"
	},
	"accounts": [
		{
			"type": "cosmos-sdk/PeriodicVestingAccount",
			"value": {
			  "BaseVestingAccount": {
				"BaseAccount": {
				  "account_number": "118",
				  "address": "kava13vt44t6uwht8mnsy0x0nx8873r5tfux7tkh4ah",
				  "coins": [
					{
					  "amount": "62500000000",
					  "denom": "ukava"
					}
				  ],
				  "public_key": null,
				  "sequence": "0"
				},
				"delegated_free": [],
				"delegated_vesting": [],
				"end_time": "1667656800",
				"original_vesting": [
				  {
					"amount": "62490000000",
					"denom": "ukava"
				  }
				]
			  },
			  "start_time": "1572962400",
			  "vesting_periods": [
				{
				  "amount": [
					{
					  "amount": "15615000000",
					  "denom": "ukava"
					}
				  ],
				  "length": "31622400"
				},
				{
				  "amount": [
					{
					  "amount": "5859375000",
					  "denom": "ukava"
					}
				  ],
				  "length": "7948800"
				},
				{
				  "amount": [
					{
					  "amount": "5859375000",
					  "denom": "ukava"
					}
				  ],
				  "length": "7689600"
				},
				{
				  "amount": [
					{
					  "amount": "5859375000",
					  "denom": "ukava"
					}
				  ],
				  "length": "7948800"
				},
				{
				  "amount": [
					{
					  "amount": "5859375000",
					  "denom": "ukava"
					}
				  ],
				  "length": "7948800"
				},
				{
				  "amount": [
					{
					  "amount": "5859375000",
					  "denom": "ukava"
					}
				  ],
				  "length": "7948800"
				},
				{
				  "amount": [
					{
					  "amount": "5859375000",
					  "denom": "ukava"
					}
				  ],
				  "length": "7689600"
				},
				{
				  "amount": [
					{
					  "amount": "5859375000",
					  "denom": "ukava"
					}
				  ],
				  "length": "7948800"
				},
				{
				  "amount": [
					{
					  "amount": "5859375000",
					  "denom": "ukava"
					}
				  ],
				  "length": "7948800"
				}
			  ]
			}
		  }
	]
}`),
	}

	expectedAuthGenState := genutil.AppMap{
		"auth": json.RawMessage(`{
			"params": {
				"max_memo_characters": "256",
				"sig_verify_cost_ed25519": "590",
				"sig_verify_cost_secp256k1": "1000",
				"tx_sig_limit": "7",
				"tx_size_cost_per_byte": "10"
			},
			"accounts": [
				{
					"type": "cosmos-sdk/PeriodicVestingAccount",
					"value": {
						"account_number": 118,
						"address": "kava13vt44t6uwht8mnsy0x0nx8873r5tfux7tkh4ah",
						"coins": [
							{
								"amount": "62500000000",
								"denom": "ukava"
							}
						],
						"public_key": "",
						"sequence": 0,
						"delegated_free": [],
						"delegated_vesting": [],
						"end_time": 1667656800,
						"original_vesting": [
							{
								"amount": "62490000000",
								"denom": "ukava"
							}
						],
						"start_time": 1572962400,
						"vesting_periods": [
							{
								"amount": [
									{
										"amount": "15615000000",
										"denom": "ukava"
									}
								],
								"length": 31622400
							},
							{
								"amount": [
									{
										"amount": "5859375000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "5859375000",
										"denom": "ukava"
									}
								],
								"length": 7689600
							},
							{
								"amount": [
									{
										"amount": "5859375000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "5859375000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "5859375000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "5859375000",
										"denom": "ukava"
									}
								],
								"length": 7689600
							},
							{
								"amount": [
									{
										"amount": "5859375000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							},
							{
								"amount": [
									{
										"amount": "5859375000",
										"denom": "ukava"
									}
								],
								"length": 7948800
							}
						]
					}
				}
			]
		}`),
	}

	newGenesisState := Migrate(oldGenesisState)
	require.JSONEq(t, string(expectedAuthGenState["auth"]), string(newGenesisState["auth"]))
}
