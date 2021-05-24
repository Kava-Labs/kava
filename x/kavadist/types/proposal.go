package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeCommunityPoolMultiSpend defines the type for a CommunityPoolSpendProposal
	ProposalTypeCommunityPoolMultiSpend = "CommunityPoolMultiSpend"
)

// Assert CommunityPoolMultiSpendProposal implements govtypes.Content at compile-time
var _ govtypes.Content = CommunityPoolMultiSpendProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolMultiSpend)
	govtypes.RegisterProposalTypeCodec(CommunityPoolMultiSpendProposal{}, "kava/CommunityPoolMultiSpendProposal")
}

// CommunityPoolMultiSpendProposal spends from the community pool
type CommunityPoolMultiSpendProposal struct {
	Title         string               `json:"title" yaml:"title"`
	Description   string               `json:"description" yaml:"description"`
	RecipientList MultiSpendRecipients `json:"recipient_list" yaml:"recipient_list"`
}

// NewCommunityPoolMultiSpendProposal creates a new community pool spned proposal.
func NewCommunityPoolMultiSpendProposal(title, description string, recipientList MultiSpendRecipients) CommunityPoolMultiSpendProposal {
	return CommunityPoolMultiSpendProposal{
		Title:         title,
		Description:   description,
		RecipientList: recipientList}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp CommunityPoolMultiSpendProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp CommunityPoolMultiSpendProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp CommunityPoolMultiSpendProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp CommunityPoolMultiSpendProposal) ProposalType() string {
	return ProposalTypeCommunityPoolMultiSpend
}

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

func (csp CommunityPoolMultiSpendProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Multi Spend Proposal:
  Title:       %s
  Description: %s
  Recipient List:   %s
`, csp.Title, csp.Description, csp.RecipientList))
	return b.String()
}

type MultiSpendRecipient struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}

func (msr MultiSpendRecipient) Validate() error {
	if !msr.Amount.IsValid() {
		return ErrInvalidProposalAmount
	}
	if msr.Address.Empty() {
		return ErrEmptyProposalRecipient
	}
	return nil
}

func (msr MultiSpendRecipient) String() string {
	return fmt.Sprintf(`Receiver: %s
	Amount: %s
	`, msr.Address, msr.Amount)
}

type MultiSpendRecipients []MultiSpendRecipient

func (msrs MultiSpendRecipients) Validate() error {
	for _, msr := range msrs {
		if err := msr.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (msrs MultiSpendRecipients) String() string {
	out := ""
	for _, msr := range msrs {
		out += msr.String()
	}
	return out
}
