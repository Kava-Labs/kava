package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeCommunityPoolMultiSpend defines the type for a CommunityPoolMultiSpendProposal
	ProposalTypeCommunityPoolMultiSpend = "CommunityPoolMultiSpend"
)

// Assert CommunityPoolMultiSpendProposal implements govtypes.Content at compile-time
var _ govtypes.Content = CommunityPoolMultiSpendProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolMultiSpend)
	govtypes.RegisterProposalTypeCodec(CommunityPoolMultiSpendProposal{}, "kava/CommunityPoolMultiSpendProposal")
}

// NewCommunityPoolMultiSpendProposal creates a new community pool multi-spend proposal.
func NewCommunityPoolMultiSpendProposal(title, description string, recipientList []MultiSpendRecipient) *CommunityPoolMultiSpendProposal {
	return &CommunityPoolMultiSpendProposal{
		Title:         title,
		Description:   description,
		RecipientList: recipientList,
	}
}

// GetTitle returns the title of a community pool multi-spend proposal.
func (csp CommunityPoolMultiSpendProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool multi-spend proposal.
func (csp CommunityPoolMultiSpendProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool multi-spend proposal.
func (csp CommunityPoolMultiSpendProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool multi-spend proposal.
func (csp CommunityPoolMultiSpendProposal) ProposalType() string {
	return ProposalTypeCommunityPoolMultiSpend
}

// ValidateBasic stateless validation of a community pool multi-spend proposal.
func (csp CommunityPoolMultiSpendProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}
	for _, msr := range csp.RecipientList {
		if err := msr.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// String implements fmt.Stringer
func (csp CommunityPoolMultiSpendProposal) String() string {
	receiptList := ""
	for _, msr := range csp.RecipientList {
		receiptList += msr.String()
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Multi Spend Proposal:
  Title:       %s
  Description: %s
  Recipient List:   %s
`, csp.Title, csp.Description, receiptList))
	return b.String()
}

// Validate stateless validation of MultiSpendRecipient
func (msr MultiSpendRecipient) Validate() error {
	if !msr.Amount.IsValid() {
		return ErrInvalidProposalAmount
	}
	if msr.Address == "" {
		return ErrEmptyProposalRecipient
	}
	if _, err := sdk.AccAddressFromBech32(msr.Address); err != nil {
		return err
	}
	return nil
}

// String implements fmt.Stringer
func (msr MultiSpendRecipient) String() string {
	return fmt.Sprintf(`Receiver: %s
	Amount: %s
	`, msr.Address, msr.Amount)
}

// Gets recipient address in sdk.AccAddress
func (msr MultiSpendRecipient) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msr.Address)
	if err != nil {
		panic(fmt.Errorf("couldn't convert %q to account address: %v", msr.Address, err))
	}

	return addr
}
