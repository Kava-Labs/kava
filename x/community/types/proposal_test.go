package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/community/types"
)

func TestLendProposals_ValidateBasic(t *testing.T) {
	// each proposalData is tested with Deposit and Withdraw proposals
	type proposalData struct {
		Title       string
		Description string
		Amount      sdk.Coins
	}
	testCases := []struct {
		name        string
		proposal    proposalData
		expectedErr string
	}{
		{
			name: "valid proposal",
			proposal: proposalData{
				Title:       "I'm a lend proposal",
				Description: "I interact with lend",
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e10)),
			},
			expectedErr: "",
		},
		{
			name: "invalid - fails gov validation",
			proposal: proposalData{
				Description: "I have no title.",
			},
			expectedErr: "invalid proposal content",
		},
		{
			name: "invalid - nil coins",
			proposal: proposalData{
				Title:       "Error profoundly",
				Description: "My coins are nil",
				Amount:      nil,
			},
			expectedErr: "invalid coins",
		},
		{
			name: "invalid - empty coins",
			proposal: proposalData{
				Title:       "Error profoundly",
				Description: "My coins are empty",
				Amount:      sdk.NewCoins(),
			},
			expectedErr: "invalid coins",
		},
		{
			name: "invalid - zero coins",
			proposal: proposalData{
				Title:       "Error profoundly",
				Description: "My coins are zero",
				Amount:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.ZeroInt())),
			},
			expectedErr: "invalid coins",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("CommunityPoolLendDepositProposal", func(t *testing.T) {
				deposit := types.NewCommunityPoolLendDepositProposal(
					tc.proposal.Title,
					tc.proposal.Description,
					tc.proposal.Amount,
				)
				err := deposit.ValidateBasic()
				if tc.expectedErr != "" {
					require.ErrorContains(t, err, tc.expectedErr)
					return
				}

				require.NoError(t, err)
				require.Equal(t, deposit.Title, deposit.GetTitle())
				require.Equal(t, deposit.Description, deposit.GetDescription())
				require.Equal(t, types.ModuleName, deposit.ProposalRoute())
				require.Equal(t, types.ProposalTypeCommunityPoolLendDeposit, deposit.ProposalType())
			})

			t.Run("CommunityPoolLendWithdrawProposal", func(t *testing.T) {
				withdrawl := types.NewCommunityPoolLendWithdrawProposal(
					tc.proposal.Title,
					tc.proposal.Description,
					tc.proposal.Amount,
				)
				err := withdrawl.ValidateBasic()
				if tc.expectedErr != "" {
					require.ErrorContains(t, err, tc.expectedErr)
					return
				}

				require.NoError(t, err)
				require.Equal(t, withdrawl.Title, withdrawl.GetTitle())
				require.Equal(t, withdrawl.Description, withdrawl.GetDescription())
				require.Equal(t, types.ModuleName, withdrawl.ProposalRoute())
				require.Equal(t, types.ProposalTypeCommunityPoolLendWithdraw, withdrawl.ProposalType())
			})
		})
	}
}

func TestCommunityPoolLendDepositProposal_Stringer(t *testing.T) {
	proposal := types.NewCommunityPoolLendDepositProposal(
		"Title",
		"Description",
		sdk.NewCoins(sdk.NewInt64Coin("ukava", 42)),
	)
	require.Equal(t, `Community Pool Lend Deposit Proposal:
  Title:       Title
  Description: Description
  Amount:      42ukava
`, proposal.String())
}

func TestCommunityPoolLendWithdrawProposal_Stringer(t *testing.T) {
	proposal := types.NewCommunityPoolLendWithdrawProposal(
		"Title",
		"Description",
		sdk.NewCoins(sdk.NewInt64Coin("ukava", 42)),
	)
	require.Equal(t, `Community Pool Lend Withdraw Proposal:
  Title:       Title
  Description: Description
  Amount:      42ukava
`, proposal.String())
}
