package iavlviewer

import (
	"fmt"
	"strings"

	"github.com/cosmos/iavl"
	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/spf13/cobra"
)

func newShapeCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shape <prefix> [version number]",
		Short: "View shape of iavl tree.",
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

			printShape(tree)

			return nil
		},
	}

	return cmd
}

func printShape(tree *iavl.MutableTree) {
	// shape := tree.RenderShape("  ", nil)
	// TODO: handle this error
	shape, _ := tree.RenderShape("  ", nodeEncoder)
	fmt.Println(strings.Join(shape, "\n"))
}
