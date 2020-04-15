package operations

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

func TestPriceGen(t *testing.T) {
	now := time.Now()
	r := rand.New(rand.NewSource(now.Unix()))

	n := int64(10000)

	price := sdk.ZeroDec()
	marketID := "btc:usd"
	prices := make([]sdk.Dec, n)
	for i := int64(0); i < n; i++ {
		prices[i] = price
		fmt.Print(i)
		fmt.Print(",")
		fmt.Println(price)
		price = pickNewRandomPrice(r, marketID)
	}

	bz, err := types.ModuleCdc.MarshalJSONIndent(prices, "", "  ")
	if err != nil {
		t.FailNow()
	}
	err = ioutil.WriteFile(fmt.Sprintf("./pricefeed_test(%s).json", now), bz, 0644)
	if err != nil {
		t.FailNow()
	}

}
