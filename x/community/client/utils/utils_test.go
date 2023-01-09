package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
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

func TestParseCommunityPoolProposal(t *testing.T) {
	config := app.MakeEncodingConfig()
	clientCtx := client.Context{}.
		WithCodec(config.Marshaler).
		WithLegacyAmino(config.Amino).
		WithInterfaceRegistry(config.InterfaceRegistry)
	okJSON := testutil.WriteToNewTempFile(t, `
	{
		"title": "A Title",
		"description": "Community pool proposal description.",
		"messages": [
			{
				"@type": "/kava.evmutil.v1beta1.MsgEVMCall",
				"to": "0x25e9171C98Fc1924Fa9415CF50750274F0664764",
				"fn_abi": "{\"inputs\": [],\"name\": \"deposit\",\"type\": \"function\"}",
				"data": "0xd0e30db0",
				"amount": "120000000000000",
				"authority": "kava17d2wax0zhjrrecvaszuyxdf5wcu5a0p4qlx3t5"
			}
		],
		"deposit": "10ukava"
	}
`)
	proposal, depsoit, err := utils.ParseCommunityPoolProposalJSON(clientCtx.Codec, okJSON.Name())
	require.NoError(t, err)

	expectedAmount, err := sdk.ParseCoinsNormalized("10ukava")
	require.NoError(t, err)

	require.Equal(t, "A Title", proposal.Title)
	require.Equal(t, "Community pool proposal description.", proposal.Description)
	require.Equal(t, expectedAmount, depsoit)

	msgs, err := proposal.GetMsgs()
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	expectedMsgStr := `to:"0x25e9171C98Fc1924Fa9415CF50750274F0664764" fn_abi:"{\"inputs\": [],\"name\": \"deposit\",\"type\": \"function\"}" data:"0xd0e30db0" amount:"120000000000000" authority:"kava17d2wax0zhjrrecvaszuyxdf5wcu5a0p4qlx3t5" `
	require.Equal(t, expectedMsgStr, msgs[0].String())
}
