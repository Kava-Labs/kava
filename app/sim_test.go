package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	auctionsimops "github.com/kava-labs/kava/x/auction/simulation/operations"
	bep3simops "github.com/kava-labs/kava/x/bep3/simulation/operations"
	cdpsimops "github.com/kava-labs/kava/x/cdp/simulation/operations"
	pricefeedsimops "github.com/kava-labs/kava/x/pricefeed/simulation/operations"
)

// Simulation parameter constants
const (
	StakePerAccount                                    = "stake_per_account"
	InitiallyBondedValidators                          = "initially_bonded_validators"
	OpWeightDeductFee                                  = "op_weight_deduct_fee"
	OpWeightMsgSend                                    = "op_weight_msg_send"
	OpWeightSingleInputMsgMultiSend                    = "op_weight_single_input_msg_multisend"
	OpWeightMsgSetWithdrawAddress                      = "op_weight_msg_set_withdraw_address"
	OpWeightMsgWithdrawDelegationReward                = "op_weight_msg_withdraw_delegation_reward"
	OpWeightMsgWithdrawValidatorCommission             = "op_weight_msg_withdraw_validator_commission"
	OpWeightSubmitVotingSlashingTextProposal           = "op_weight_submit_voting_slashing_text_proposal"
	OpWeightSubmitVotingSlashingCommunitySpendProposal = "op_weight_submit_voting_slashing_community_spend_proposal"
	OpWeightSubmitVotingSlashingParamChangeProposal    = "op_weight_submit_voting_slashing_param_change_proposal"
	OpWeightMsgDeposit                                 = "op_weight_msg_deposit"
	OpWeightMsgCreateValidator                         = "op_weight_msg_create_validator"
	OpWeightMsgEditValidator                           = "op_weight_msg_edit_validator"
	OpWeightMsgDelegate                                = "op_weight_msg_delegate"
	OpWeightMsgUndelegate                              = "op_weight_msg_undelegate"
	OpWeightMsgBeginRedelegate                         = "op_weight_msg_begin_redelegate"
	OpWeightMsgUnjail                                  = "op_weight_msg_unjail"
	OpWeightMsgPlaceBid                                = "op_weight_msg_place_bid"
	OpWeightMsgPricefeed                               = "op_weight_msg_pricefeed"
	OpWeightMsgCreateAtomicSwap                        = "op_weight_msg_create_atomic_Swap"
	OpWeightMsgCdp                                     = "op_weight_msg_cdp"
)

// TestMain runs setup and teardown code before all tests.
func TestMain(m *testing.M) {
	// set prefixes
	config := sdk.GetConfig()
	SetBech32AddressPrefixes(config)
	config.Seal()
	// load the values from simulation specific flags
	simapp.GetSimulatorFlags()
}

func testAndRunTxs(app *App, config simulation.Config) []simulation.WeightedOperation {
	ap := make(simulation.AppParams)

	paramChanges := app.sm.GenerateParamChanges(config.Seed)

	if config.ParamsFile != "" {
		bz, err := ioutil.ReadFile(config.ParamsFile)
		if err != nil {
			panic(err)
		}

		app.cdc.MustUnmarshalJSON(bz, &ap)
	}

	// nolint: govet
	return []simulation.WeightedOperation{
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightDeductFee, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			authsimops.SimulateDeductFee(app.accountKeeper, app.supplyKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgSend, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			banksimops.SimulateMsgSend(app.accountKeeper, app.bankKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightSingleInputMsgMultiSend, &v, nil,
					func(_ *rand.Rand) {
						v = 10
					})
				return v
			}(nil),
			banksimops.SimulateSingleInputMsgMultiSend(app.accountKeeper, app.bankKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgSetWithdrawAddress, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			distrsimops.SimulateMsgSetWithdrawAddress(app.distrKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgWithdrawDelegationReward, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			distrsimops.SimulateMsgWithdrawDelegatorReward(app.distrKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgWithdrawValidatorCommission, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			distrsimops.SimulateMsgWithdrawValidatorCommission(app.distrKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightSubmitVotingSlashingTextProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			govsimops.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, govsimops.SimulateTextProposalContent),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightSubmitVotingSlashingCommunitySpendProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			govsimops.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, distrsimops.SimulateCommunityPoolSpendProposalContent(app.distrKeeper)),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightSubmitVotingSlashingParamChangeProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			govsimops.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, paramsimops.SimulateParamChangeProposalContent(paramChanges)),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgDeposit, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			govsimops.SimulateMsgDeposit(app.govKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgCreateValidator, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			stakingsimops.SimulateMsgCreateValidator(app.accountKeeper, app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgEditValidator, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			stakingsimops.SimulateMsgEditValidator(app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgDelegate, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			stakingsimops.SimulateMsgDelegate(app.accountKeeper, app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgUndelegate, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			stakingsimops.SimulateMsgUndelegate(app.accountKeeper, app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgBeginRedelegate, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			stakingsimops.SimulateMsgBeginRedelegate(app.accountKeeper, app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgUnjail, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			slashingsimops.SimulateMsgUnjail(app.slashingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgPlaceBid, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			auctionsimops.SimulateMsgPlaceBid(app.accountKeeper, app.auctionKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgCreateAtomicSwap, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			bep3simops.SimulateMsgCreateAtomicSwap(app.accountKeeper, app.bep3Keeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgPricefeed, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			pricefeedsimops.SimulateMsgUpdatePrices(app.pricefeedKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(app.cdc, OpWeightMsgCdp, &v, nil,
					func(_ *rand.Rand) {
						v = 100 // TODO
					})
				return v
			}(nil),
			cdpsimops.SimulateMsgCdp(app.accountKeeper, app.cdpKeeper, app.pricefeedKeeper),
		},
	}
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestFullAppSimulation(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewApp(logger, db, nil, true, simapp.FlagPeriodValue, fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}
}

func TestAppImportExport(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewApp(logger, db, nil, true, simapp.FlagPeriodValue, fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	appState, _, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := simapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewApp(log.NewNopLogger(), newDB, nil, true, simapp.FlagPeriodValue, fauxMerkleModeOpt)
	require.Equal(t, appName, newApp.Name())

	var genesisState GenesisState
	err = app.Codec().UnmarshalJSON(appState, &genesisState)
	require.NoError(t, err)

	ctxA := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContext(true, abci.Header{Height: app.LastBlockHeight()})
	newApp.mm.InitGenesis(ctxB, genesisState)

	fmt.Printf("comparing stores...\n")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keys[baseapp.MainStoreKey], newApp.keys[baseapp.MainStoreKey], [][]byte{}},
		{app.keys[auth.StoreKey], newApp.keys[auth.StoreKey], [][]byte{}},
		{app.keys[staking.StoreKey], newApp.keys[staking.StoreKey],
			[][]byte{
				staking.UnbondingQueueKey, staking.RedelegationQueueKey, staking.ValidatorQueueKey,
			}}, // ordering may change but it doesn't matter
		{app.keys[slashing.StoreKey], newApp.keys[slashing.StoreKey], [][]byte{}},
		{app.keys[mint.StoreKey], newApp.keys[mint.StoreKey], [][]byte{}},
		{app.keys[distr.StoreKey], newApp.keys[distr.StoreKey], [][]byte{}},
		{app.keys[supply.StoreKey], newApp.keys[supply.StoreKey], [][]byte{}},
		{app.keys[params.StoreKey], newApp.keys[params.StoreKey], [][]byte{}},
		{app.keys[gov.StoreKey], newApp.keys[gov.StoreKey], [][]byte{}},
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, len(failedKVAs), 0, simapp.GetSimulationLog(skp.A.Name(), app.SimulationManager().StoreDecoders, app.Codec(), failedKVAs, failedKVBs))
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation after import")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewApp(logger, db, nil, true, simapp.FlagPeriodValue, fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	appState, _, err := app.ExportAppStateAndValidators(true, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := simapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewApp(log.NewNopLogger(), newDB, nil, true, simapp.FlagPeriodValue, fauxMerkleModeOpt)
	require.Equal(t, appName, newApp.Name())

	newApp.InitChain(abci.RequestInitChain{
		AppStateBytes: appState,
	})

	_, _, err = simulation.SimulateFromSeed(
		t, os.Stdout, newApp.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(newApp, newApp.Codec(), config),
		newApp.ModuleAccountAddrs(), config,
	)
	require.NoError(t, err)
}

func TestAppStateDeterminism(t *testing.T) {
	if !simapp.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simapp.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = helpers.SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simapp.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()

			app := NewApp(logger, db, nil, true, simapp.FlagPeriodValue, interBlockCacheOpt())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
				simapp.SimulationOperations(app, app.Codec(), config),
				app.ModuleAccountAddrs(), config,
			)
			require.NoError(t, err)

			if config.Commit {
				simapp.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, appHashList[0], appHashList[j],
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
