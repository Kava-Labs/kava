package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/spf13/cobra"

	"github.com/kava-labs/kava/cmd/kava/cmd"
)

func main() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		cobra.CheckErr(fmt.Errorf("Failed to get home dir: %s", err))
	}
	defaultNodeHome := filepath.Join(userHomeDir, ".kava")

	rootCmd := cmd.NewRootCmd(defaultNodeHome)

	if err := svrcmd.Execute(rootCmd, defaultNodeHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}
