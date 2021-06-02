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

// CommunityPoolMultiSpendProposal spends from the community pool by sending to one or more addresses
type CommunityPoolMultiSpendProposal struct {
	Title         string               `json:"title" yaml:"title"`
	Description   string               `json:"description" yaml:"description"`
	RecipientList MultiSpendRecipients `json:"recipient_list" yaml:"recipient_list"`
}

// NewCommunityPoolMultiSpendProposal creates a new community pool multi-spend proposal.
func NewCommunityPoolMultiSpendProposal(title, description string, recipientList MultiSpendRecipients) CommunityPoolMultiSpendProposal {
	return CommunityPoolMultiSpendProposal{
		Title:         title,
		Description:   description,
		RecipientList: recipientList}
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
	if err := csp.RecipientList.Validate(); err != nil {
		return err
	}
	return nil
}

// String implements fmt.Stringer
func (csp CommunityPoolMultiSpendProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Multi Spend Proposal:
  Title:       %s
  Description: %s
  Recipient List:   %s
`, csp.Title, csp.Description, csp.RecipientList))
	return b.String()
}

// MultiSpendRecipient defines a recipient and the amount of coins they are receiving
type MultiSpendRecipient struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}

// Validate stateless validation of MultiSpendRecipient
func (msr MultiSpendRecipient) Validate() error {
	if !msr.Amount.IsValid() {
		return ErrInvalidProposalAmount
	}
	if msr.Address.Empty() {
		return ErrEmptyProposalRecipient
	}
	return nil
}

// String implements fmt.Stringer
func (msr MultiSpendRecipient) String() string {
	return fmt.Sprintf(`Receiver: %s
	Amount: %s
	`, msr.Address, msr.Amount)
}

// MultiSpendRecipients slice of MultiSpendRecipient
type MultiSpendRecipients []MultiSpendRecipient

// Validate stateless validation of MultiSpendRecipients
func (msrs MultiSpendRecipients) Validate() error {
	for _, msr := range msrs {
		if err := msr.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// String implements fmt.Stringer
func (msrs MultiSpendRecipients) String() string {
	out := ""
	for _, msr := range msrs {
		out += msr.String()
	}
	return out
}
