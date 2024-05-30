package iavlviewer

import (
	"fmt"

	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/spf13/cobra"
)

func newHashCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash <prefix> [version number]",
		Short: "Print the root hash of the iavl tree.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := args[0]
			version := 0
			if len(args) == 2 {
				var err error
				version, err = parseVersion(args[1])
				if err != nil {
					return err
				}
			}

			tree, err := openPrefixTree(opts, cmd, prefix, version)
			if err != nil {
				return err
			}

			hash, err := tree.Hash()
			if err != nil {
				return err
			}
			fmt.Printf("Hash: %X\n", hash)

			return nil
		},
	}

	return cmd
}
