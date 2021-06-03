package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestParams_UnmarshalJSON(t *testing.T) {
	pairs := types.NewPairs(
		types.NewPair("ukava", "hard", sdk.ZeroDec()),
		types.NewPair("usdx", "hard", sdk.ZeroDec()),
	)
	pairData, err := json.Marshal(pairs)
	require.NoError(t, err)

	fee, err := sdk.NewDecFromStr("0.5")
	require.NoError(t, err)
	feeData, err := json.Marshal(fee)
	require.NoError(t, err)

	data := fmt.Sprintf(`{
	"pairs": %s,
	"swap_fee": %s
}`, string(pairData), string(feeData))

	var params types.Params
	err = json.Unmarshal([]byte(data), &params)
	require.NoError(t, err)

	assert.Equal(t, pairs, params.Pairs)
	assert.Equal(t, fee, params.SwapFee)
}

func TestParams_MarshalYAML(t *testing.T) {
	pairs := types.NewPairs(
		types.NewPair("ukava", "hard", sdk.ZeroDec()),
		types.NewPair("usdx", "hard", sdk.ZeroDec()),
	)
	fee, err := sdk.NewDecFromStr("0.5")
	require.NoError(t, err)

	p := types.Params{
		Pairs:   pairs,
		SwapFee: fee,
	}

	data, err := yaml.Marshal(p)
	require.NoError(t, err)

	fmt.Println(string(data))

	var params map[string]interface{}
	err = yaml.Unmarshal(data, &params)
	require.NoError(t, err)

	_, ok := params["pairs"]
	require.True(t, ok)
	_, ok = params["swap_fee"]
	require.True(t, ok)
}
