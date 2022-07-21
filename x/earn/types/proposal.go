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
)

// Assert CommunityPoolMultiSpendProposal implements govtypes.Content at compile-time
var _ govtypes.Content = CommunityPoolDepositProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolDeposit)
	govtypes.RegisterProposalTypeCodec(CommunityPoolDepositProposal{}, "kava/CommunityPoolDepositProposal")
}

// NewCommunityPoolMultiSpendProposal creates a new community pool deposit proposal.
func NewCommunityPoolDepositProposal(title, description string, amount sdk.Coins) *CommunityPoolDepositProposal {
	return &CommunityPoolDepositProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
	}
}

// GetTitle returns the title of a community pool deposit proposal.
func (cdp CommunityPoolDepositProposal) GetTitle() string { return cdp.Title }

// GetDescription returns the description of a community pool deposit proposal.
func (cdp CommunityPoolDepositProposal) GetDescription() string { return cdp.Description }

// GetDescription returns the routing key of a community pool deposit proposal.
func (cdp CommunityPoolDepositProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool deposit proposal.
func (cdp CommunityPoolDepositProposal) ProposalType() string {
	return ProposalTypeCommunityPoolDeposit
}

// String implements fmt.Stringer
func (cdp CommunityPoolDepositProposal) String() string {

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Multi Spend Proposal:
  Title:       %s
  Description: %s
  Amount:   %s
`, cdp.Title, cdp.Description, cdp.Amount))
	return b.String()
}

// ValidateBasic stateless validation of a community pool multi-spend proposal.
func (cdp CommunityPoolDepositProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(cdp)
	if err != nil {
		return err
	}
	return cdp.Amount.Validate()
}
