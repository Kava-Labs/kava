package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeCommunityPoolDeposit defines the type for a CommunityPoolDepositProposal
	ProposalTypeCommunityPoolDeposit = "CommunityPoolDeposit"
	// ProposalTypeCommunityPoolWithdraw defines the type for a CommunityPoolDepositProposal
	ProposalTypeCommunityPoolWithdraw = "CommunityPoolWithdraw"
)

// Assert CommunityPoolDepositProposal implements govtypes.Content at compile-time
var (
	_ govtypes.Content = &CommunityPoolDepositProposal{}
	_ govtypes.Content = &CommunityPoolWithdrawProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolDeposit)
	govtypes.RegisterProposalTypeCodec(&CommunityPoolDepositProposal{}, "kava/CommunityPoolDepositProposal")
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolWithdraw)
	govtypes.RegisterProposalTypeCodec(&CommunityPoolWithdrawProposal{}, "kava/CommunityPoolWithdrawProposal")
}

// NewCommunityPoolDepositProposal creates a new community pool deposit proposal.
func NewCommunityPoolDepositProposal(title, description string, amount sdk.Coin) *CommunityPoolDepositProposal {
	return &CommunityPoolDepositProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
	}
}

// GetTitle returns the title of a community pool deposit proposal.
func (cdp *CommunityPoolDepositProposal) GetTitle() string { return cdp.Title }

// GetDescription returns the description of a community pool deposit proposal.
func (cdp *CommunityPoolDepositProposal) GetDescription() string { return cdp.Description }

// GetDescription returns the routing key of a community pool deposit proposal.
func (cdp *CommunityPoolDepositProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool deposit proposal.
func (cdp *CommunityPoolDepositProposal) ProposalType() string {
	return ProposalTypeCommunityPoolDeposit
}

// String implements fmt.Stringer
func (cdp *CommunityPoolDepositProposal) String() string {

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Deposit Proposal:
  Title:       %s
  Description: %s
  Amount:   %s
`, cdp.Title, cdp.Description, cdp.Amount))
	return b.String()
}

// ValidateBasic stateless validation of a community pool multi-spend proposal.
func (cdp *CommunityPoolDepositProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(cdp)
	if err != nil {
		return err
	}
	return cdp.Amount.Validate()
}

// NewCommunityPoolWithdrawProposal creates a new community pool deposit proposal.
func NewCommunityPoolWithdrawProposal(title, description string, amount sdk.Coin) *CommunityPoolWithdrawProposal {
	return &CommunityPoolWithdrawProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
	}
}

// GetTitle returns the title of a community pool withdraw proposal.
func (cdp *CommunityPoolWithdrawProposal) GetTitle() string { return cdp.Title }

// GetDescription returns the description of a community pool withdraw proposal.
func (cdp *CommunityPoolWithdrawProposal) GetDescription() string { return cdp.Description }

// GetDescription returns the routing key of a community pool withdraw proposal.
func (cdp *CommunityPoolWithdrawProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool withdraw proposal.
func (cdp *CommunityPoolWithdrawProposal) ProposalType() string {
	return ProposalTypeCommunityPoolWithdraw
}

// String implements fmt.Stringer
func (cdp *CommunityPoolWithdrawProposal) String() string {

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Withdraw Proposal:
  Title:       %s
  Description: %s
  Amount:   %s
`, cdp.Title, cdp.Description, cdp.Amount))
	return b.String()
}

// ValidateBasic stateless validation of a community pool multi-spend proposal.
func (cdp *CommunityPoolWithdrawProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(cdp)
	if err != nil {
		return err
	}
	return cdp.Amount.Validate()
}
