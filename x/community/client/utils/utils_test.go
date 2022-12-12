package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/client/utils"
)

func TestParseDepositProposal(t *testing.T) {
	cdc := codec.NewAminoCodec(codec.NewLegacyAmino())
	okJSON := testutil.WriteToNewTempFile(t, `
{
  "title": "Community Pool Lend Deposit",
  "description": "Deposit some KAVA from community pool to Lend!",
  "amount": [
    {
      "denom": "ukava",
      "amount": "100000000000"
    }
  ]
}
`)
	proposal, err := utils.ParseCommunityPoolLendDepositProposal(cdc, okJSON.Name())
	require.NoError(t, err)

	expectedAmount, err := sdk.ParseCoinsNormalized("100000000000ukava")
	require.NoError(t, err)

	require.Equal(t, "Community Pool Lend Deposit", proposal.Title)
	require.Equal(t, "Deposit some KAVA from community pool to Lend!", proposal.Description)
	require.Equal(t, expectedAmount, proposal.Amount)
}

func TestParseWithdrawProposal(t *testing.T) {
	cdc := codec.NewAminoCodec(codec.NewLegacyAmino())
	okJSON := testutil.WriteToNewTempFile(t, `
{
  "title": "Community Pool Lend Withdraw",
  "description": "Withdraw some KAVA from community pool to Lend!",
  "amount": [
    {
      "denom": "ukava",
      "amount": "100000000000"
    }
  ]
}
`)
	proposal, err := utils.ParseCommunityPoolLendWithdrawProposal(cdc, okJSON.Name())
	require.NoError(t, err)

	expectedAmount, err := sdk.ParseCoinsNormalized("100000000000ukava")
	require.NoError(t, err)

	require.Equal(t, "Community Pool Lend Withdraw", proposal.Title)
	require.Equal(t, "Withdraw some KAVA from community pool to Lend!", proposal.Description)
	require.Equal(t, expectedAmount, proposal.Amount)
}
