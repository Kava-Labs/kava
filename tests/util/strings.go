package util

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func PrettyPrintCoins(coins sdk.Coins) string {
	if len(coins) == 0 {
		return ""
	}

	out := make([]string, 0, len(coins))
	for _, coin := range coins {
		out = append(out, coin.String())
	}
	return fmt.Sprintf("- %s", strings.Join(out, "\n- "))
}
