package v0_15

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "kavadist"

	// RouterKey Top level router key
	RouterKey = ModuleName

	// ProposalTypeCommunityPoolMultiSpend defines the type for a CommunityPoolMultiSpendProposal
	ProposalTypeCommunityPoolMultiSpend = "CommunityPoolMultiSpend"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params            Params    `json:"params" yaml:"params"`
	PreviousBlockTime time.Time `json:"previous_block_time" yaml:"previous_block_time"`
}

// Params governance parameters for kavadist module
type Params struct {
	Active  bool    `json:"active" yaml:"active"`
	Periods Periods `json:"periods" yaml:"periods"`
}

// Periods array of Period
type Periods []Period

// Period stores the specified start and end dates, and the inflation, expressed as a decimal representing the yearly APR of KAVA tokens that will be minted during that period
type Period struct {
	Start     time.Time `json:"start" yaml:"start"`         // example "2020-03-01T15:20:00Z"
	End       time.Time `json:"end" yaml:"end"`             // example "2020-06-01T15:20:00Z"
	Inflation sdk.Dec   `json:"inflation" yaml:"inflation"` // example "1.000000003022265980"  - 10% inflation
}

var _ v036gov.Content = CommunityPoolMultiSpendProposal{}

// CommunityPoolMultiSpendProposal spends from the community pool by sending to one or more addresses
type CommunityPoolMultiSpendProposal struct {
	Title         string               `json:"title" yaml:"title"`
	Description   string               `json:"description" yaml:"description"`
	RecipientList MultiSpendRecipients `json:"recipient_list" yaml:"recipient_list"`
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
	err := v036gov.ValidateAbstract(csp)
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

// MultiSpendRecipient defines a recipient and the amount of coins they are receiving
type MultiSpendRecipient struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Amount  sdk.Coins      `json:"amount" yaml:"amount"`
}

// Validate stateless validation of MultiSpendRecipient
func (msr MultiSpendRecipient) Validate() error {
	if !msr.Amount.IsValid() {
		return fmt.Errorf("invalid community pool multi-spend proposal amount")
	}
	if msr.Address.Empty() {
		return fmt.Errorf("invalid community pool multi-spend proposal recipient")
	}
	return nil
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(CommunityPoolMultiSpendProposal{}, "kava/CommunityPoolMultiSpendProposal", nil)
}
