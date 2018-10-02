// Copyright 2016 All in Bits, inc
// Modifications copyright 2018 Kava Labs

package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/kava-labs/kava/internal/app"
)

func main() {
	// Create an app codec
	cdc := app.CreateKavaAppCodec()

	// Create a server context (a struct of a tendermint config and a logger)
	ctx := server.NewDefaultContext()

	// Create the root kvd command
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "kvd",
		Short:             "Kava Daemon",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	// Add server commands to kvd, passing in the app
	appInit := app.KavaAppInit()
	appCreator := server.ConstructAppCreator(newApp, "kava") // init db before calling newApp
	appExporter := server.ConstructAppExporter(exportAppStateAndTMValidators, "kava")

	server.AddCommands(ctx, cdc, rootCmd, appInit, appCreator, appExporter)

	// Add custom init command
	rootCmd.AddCommand(initTestnetCmd())

	// handle envs and add some flags and stuff
	executor := cli.PrepareBaseCmd(rootCmd, "KV", app.DefaultNodeHome)

	// Run kvd
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewKavaApp(logger, db, traceStore, baseapp.SetPruning(viper.GetString("pruning")))
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB, traceStore io.Writer) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	tempApp := app.NewKavaApp(logger, db, traceStore)
	return tempApp.ExportAppStateAndValidators()
}

func initTestnetCmd() *cobra.Command {
	flagChainID := "chain-id"
	cmd := &cobra.Command{
		Use:   "init-testnet-helper",
		Short: "Setup genesis and config to join testnet.",
		Long:  "Copy the genesis.json and config.toml files from the testnets folder into the default config directories.",
		RunE: func(cmd *cobra.Command, args []string) error {
			testnetVersion := viper.GetString(flagChainID)
			genesisFileName := "genesis.json"
			configFileName := "config.toml"
			configPath := "config"
			testnetsPath := os.ExpandEnv("$GOPATH/src/github.com/kava-labs/kava/testnets/")

			// Copy genesis file from testnet folder to config directories
			// Copied to .kvcli to enable automatic reading of chain-id
			genesis := filepath.Join(testnetsPath, testnetVersion, genesisFileName)
			err := copyFile(genesis, filepath.Join(app.DefaultNodeHome, configPath, genesisFileName))
			if err != nil {
				return err
			}
			err = copyFile(genesis, filepath.Join(app.DefaultCLIHome, configPath, genesisFileName))
			if err != nil {
				return err
			}
			// Copy config file from testnet folder to config directories
			// Custom config file specifies seeds and altered ports
			config := filepath.Join(testnetsPath, testnetVersion, configFileName)
			err = copyFile(config, filepath.Join(app.DefaultNodeHome, configPath, configFileName))
			if err != nil {
				return err
			}
			err = copyFile(config, filepath.Join(app.DefaultCLIHome, configPath, configFileName))
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().String(flagChainID, "", "testnet chain-id, cannot be left blank")
	return cmd
}

func copyFile(src string, dst string) error {
	// read in source file
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// create destination file (and any necessary directories)(overwriting if it exists already)
	path := filepath.Dir(dst)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	// copy file contents
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	// write to disk
	err = out.Sync()
	return err
}
