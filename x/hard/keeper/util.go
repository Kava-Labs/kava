package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Set setDifference: A - B
func setDifference(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

func getDenoms(coins sdk.Coins) []string {
	denoms := []string{}
	for _, coin := range coins {
		denoms = append(denoms, coin.Denom)
	}
	return denoms
}

func removeDuplicates(one []string, two []string) []string {
	check := make(map[string]int)
	fullList := append(one, two...)

	res := []string{}
	for _, val := range fullList {
		check[val] = 1
	}

	for key := range check {
		res = append(res, key)
	}
	sort.Strings(res)
	return res
}
