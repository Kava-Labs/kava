package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeCommunityPoolLendDeposit defines the type for a CommunityPoolLendDepositProposal
	ProposalTypeCommunityPoolLendDeposit = "CommunityPoolLendDeposit"
	// ProposalTypeCommunityPoolLendWithdraw defines the type for a CommunityPoolLendDepositProposal
	ProposalTypeCommunityPoolLendWithdraw = "CommunityPoolLendWithdraw"
)

// Assert CommunityPoolLendDepositProposal implements govtypes.Content at compile-time
var (
	_ govtypes.Content = &CommunityPoolLendDepositProposal{}
	_ govtypes.Content = &CommunityPoolLendWithdrawProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolLendDeposit)
	govtypes.RegisterProposalTypeCodec(&CommunityPoolLendDepositProposal{}, "kava/CommunityPoolLendDepositProposal")
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolLendWithdraw)
	govtypes.RegisterProposalTypeCodec(&CommunityPoolLendWithdrawProposal{}, "kava/CommunityPoolLendWithdrawProposal")
}

// NewCommunityPoolLendDepositProposal creates a new community pool deposit proposal.
func NewCommunityPoolLendDepositProposal(title, description string, amount sdk.Coins) *CommunityPoolLendDepositProposal {
	return &CommunityPoolLendDepositProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
	}
}

// GetTitle returns the title of a community pool lend deposit proposal.
func (p *CommunityPoolLendDepositProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of a community pool lend deposit proposal.
func (p *CommunityPoolLendDepositProposal) GetDescription() string { return p.Description }

// GetDescription returns the routing key of a community pool lend deposit proposal.
func (p *CommunityPoolLendDepositProposal) ProposalRoute() string { return ModuleName }

// ProposalType returns the type of a community pool lend deposit proposal.
func (p *CommunityPoolLendDepositProposal) ProposalType() string {
	return ProposalTypeCommunityPoolLendDeposit
}

// String implements fmt.Stringer
func (p *CommunityPoolLendDepositProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Lend Deposit Proposal:
  Title:       %s
  Description: %s
  Amount:      %s
`, p.Title, p.Description, p.Amount))
	return b.String()
}

// ValidateBasic stateless validation of a community pool lend deposit proposal.
func (p *CommunityPoolLendDepositProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(p); err != nil {
		return err
	}
	// ensure the proposal has valid amount
	if !p.Amount.IsValid() || p.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "deposit amount %s", p.Amount)
	}
	return p.Amount.Validate()
}

// NewCommunityPoolLendWithdrawProposal creates a new community pool lend withdraw proposal.
func NewCommunityPoolLendWithdrawProposal(title, description string, amount sdk.Coins) *CommunityPoolLendWithdrawProposal {
	return &CommunityPoolLendWithdrawProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
	}
}

// GetTitle returns the title of a community pool withdraw proposal.
func (p *CommunityPoolLendWithdrawProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of a community pool withdraw proposal.
func (p *CommunityPoolLendWithdrawProposal) GetDescription() string { return p.Description }

// GetDescription returns the routing key of a community pool withdraw proposal.
func (p *CommunityPoolLendWithdrawProposal) ProposalRoute() string { return ModuleName }

// ProposalType returns the type of a community pool withdraw proposal.
func (p *CommunityPoolLendWithdrawProposal) ProposalType() string {
	return ProposalTypeCommunityPoolLendWithdraw
}

// String implements fmt.Stringer
func (p *CommunityPoolLendWithdrawProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Lend Withdraw Proposal:
  Title:       %s
  Description: %s
  Amount:      %s
`, p.Title, p.Description, p.Amount))
	return b.String()
}

// ValidateBasic stateless validation of a community pool withdraw proposal.
func (p *CommunityPoolLendWithdrawProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(p); err != nil {
		return err
	}
	// ensure the proposal has valid amount
	if !p.Amount.IsValid() || p.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "withdraw amount %s", p.Amount)
	}
	return p.Amount.Validate()
}
