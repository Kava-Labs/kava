package iavlviewer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/iavl"
	ethermintserver "github.com/evmos/ethermint/server"
	"github.com/spf13/cobra"
)

func newVersionsCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions <prefix>",
		Short: "Print all versions of iavl tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := args[0]
			tree, err := openPrefixTree(opts, cmd, prefix, 15)
			if err != nil {
				return err
			}

			printVersions(tree)

			return nil
		},
	}

	return cmd
}

func printVersions(tree *iavl.MutableTree) {
	versions := tree.AvailableVersions()
	fmt.Println("Available versions:")
	for _, v := range versions {
		fmt.Printf("  %d\n", v)
	}
}

// parseWeaveKey assumes a separating : where all in front should be ascii,
// and all afterwards may be ascii or binary
func parseWeaveKey(key []byte) string {
	cut := bytes.IndexRune(key, ':')
	if cut == -1 {
		return encodeID(key)
	}
	prefix := key[:cut]
	id := key[cut+1:]
	return fmt.Sprintf("%s:%s", encodeID(prefix), encodeID(id))
}

// casts to a string if it is printable ascii, hex-encodes otherwise
func encodeID(id []byte) string {
	for _, b := range id {
		if b < 0x20 || b >= 0x80 {
			return strings.ToUpper(hex.EncodeToString(id))
		}
	}
	return string(id)
}

func nodeEncoder(id []byte, depth int, isLeaf bool) string {
	prefix := fmt.Sprintf("-%d ", depth)
	if isLeaf {
		prefix = fmt.Sprintf("*%d ", depth)
	}
	if len(id) == 0 {
		return fmt.Sprintf("%s<nil>", prefix)
	}
	return fmt.Sprintf("%s%s", prefix, parseWeaveKey(id))
}
