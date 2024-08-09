package iavlviewer

import (
	"encoding/json"
	"fmt"
	"os"

	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/spf13/cobra"
)

func newCommitInfoCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit-info [version number]",
		Short: "Print the commit info at a given height.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := parseVersion(args[0])
			if err != nil {
				return fmt.Errorf("invalid version: %w", err)
			}

			db, err := openDB(opts, cmd)
			if err != nil {
				return err
			}

			infos, err := GetCommitInfo(db, int64(version))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting commit info: %s\n", err)
				os.Exit(1)
			}

			jsonBytes, err := json.MarshalIndent(infos, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshalling commit infos: %s\n", err)
				os.Exit(1)
			}

			fmt.Println(string(jsonBytes))

			return nil
		},
	}

	return cmd
}
