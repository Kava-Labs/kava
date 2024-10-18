package app

//const UpgradeName = "v045-to-v050"
//
//func setupLegacyKeyTables(k *paramskeeper.Keeper) {
//	for _, subspace := range k.GetSubspaces() {
//		subspace := subspace
//
//		var keyTable paramstypes.KeyTable
//		switch subspace.Name() {
//		case authtypes.ModuleName:
//			keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
//		case banktypes.ModuleName:
//			keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
//		case stakingtypes.ModuleName:
//			keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
//		case minttypes.ModuleName:
//			keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
//		case distrtypes.ModuleName:
//			keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
//		case slashingtypes.ModuleName:
//			keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
//		case govtypes.ModuleName:
//			keyTable = govv1.ParamKeyTable() //nolint:staticcheck
//		case crisistypes.ModuleName:
//			keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
//			// wasm
//		case ibctransfertypes.ModuleName:
//			keyTable = ibctransfertypes.ParamKeyTable() //nolint:staticcheck
//		default:
//			continue
//		}
//
//		if !subspace.HasKeyTable() {
//			subspace.WithKeyTable(keyTable)
//		}
//	}
//
//	// sdk 47
//	k.Subspace(baseapp.Paramspace).
//		WithKeyTable(paramstypes.ConsensusParamsKeyTable())
//}
//
//func (app App) RegisterUpgradeHandlers() {
//	setupLegacyKeyTables(&app.paramsKeeper)
//
//	app.upgradeKeeper.SetUpgradeHandler(
//		UpgradeName,
//		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
//			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
//		},
//	)
//
//	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
//	if err != nil {
//		panic(err)
//	}
//
//	if upgradeInfo.Name == UpgradeName && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
//		storeUpgrades := storetypes.StoreUpgrades{
//			Added: []string{
//				consensustypes.ModuleName,
//				crisistypes.ModuleName,
//				capabilitytypes.MemStoreKey,
//				nft.ModuleName,
//			},
//		}
//
//		// configure store loader that checks if version == upgradeHeight and applies store upgrades
//		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
//	}
//}
//
//func (app App) RegisterUpgradeHandlersOld() {
//	baseAppLegacySS := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
//	app.UpgradeKeeper.SetUpgradeHandler(
//		UpgradeName,
//		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
//			baseapp.MigrateParams(ctx.(sdk.Context), baseAppLegacySS, app.ConsensusParamsKeeper.ParamsStore)
//			consensusParams := baseapp.GetConsensusParams(ctx.(sdk.Context), baseAppLegacySS)
//			// make sure the consensus params are set
//			if consensusParams.Block == nil || consensusParams.Evidence == nil || consensusParams.Validator == nil {
//				defaultParams := tmtypes.DefaultConsensusParams().ToProto()
//				app.ConsensusParamsKeeper.ParamsStore.Set(ctx.(sdk.Context), defaultParams)
//			}
//
//			storesvc := runtime.NewKVStoreService(app.GetKey("upgrade"))
//			consensuskeeper := consensuskeeper.NewKeeper(
//				app.appCodec,
//				storesvc,
//				app.AccountKeeper.GetAuthority(),
//				runtime.EventService{},
//			)
//
//			params, err := consensuskeeper.ParamsStore.Get(ctx)
//			if err != nil {
//				return nil, err
//			}
//
//			err = app.ConsensusParamsKeeper.ParamsStore.Set(ctx, params)
//			if err != nil {
//				return nil, err
//			}
//
//			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
//		},
//	)
//
//	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
//	if err != nil {
//		panic(err)
//	}
//
//	if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
//		storeUpgrades := storetypes.StoreUpgrades{
//			Added: []string{
//				consensustypes.ModuleName,
//				crisistypes.ModuleName,
//				circuittypes.ModuleName,
//				ibcfee.ModuleName,
//			},
//		}
//
//		// configure store loader that checks if version == upgradeHeight and applies store upgrades
//		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
//	}
//}

func (app App) RegisterUpgradeHandlers() {}
