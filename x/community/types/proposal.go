package types

import (
	fmt "fmt"
	"strings"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeCommunityPoolLendDeposit defines the type for a CommunityPoolLendDepositProposal
	ProposalTypeCommunityPoolLendDeposit = "CommunityPoolLendDeposit"
	// ProposalTypeCommunityPoolLendWithdraw defines the type for a CommunityPoolLendDepositProposal
	ProposalTypeCommunityPoolLendWithdraw = "CommunityPoolLendWithdraw"
	// ProposalTypeCommunityPool defines the type for a CommunityPoolProposal
	ProposalTypeCommunityPool = "ProposalTypeCommunityPool"
)

// Assert proposals implements govtypes.Content at compile-time
var (
	_ govtypes.Content                 = &CommunityPoolLendDepositProposal{}
	_ govtypes.Content                 = &CommunityPoolLendWithdrawProposal{}
	_ govtypes.Content                 = CommunityPoolProposal{}
	_ cdctypes.UnpackInterfacesMessage = CommunityPoolProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolLendDeposit)
	govtypes.RegisterProposalTypeCodec(&CommunityPoolLendDepositProposal{}, "kava/CommunityPoolLendDepositProposal")
	govtypes.RegisterProposalType(ProposalTypeCommunityPoolLendWithdraw)
	govtypes.RegisterProposalTypeCodec(&CommunityPoolLendWithdrawProposal{}, "kava/CommunityPoolLendWithdrawProposal")
	govtypes.RegisterProposalType(ProposalTypeCommunityPool)
	govtypes.RegisterProposalTypeCodec(CommunityPoolProposal{}, "kava/CommunityPoolProposal")
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

// NewCommunityPoolProposal creates a new community pool proposal.
func NewCommunityPoolProposal(title, description string, messages []sdk.Msg) (*CommunityPoolProposal, error) {
	msgs, err := packMsgs(messages)
	if err != nil {
		return &CommunityPoolProposal{}, err
	}

	return &CommunityPoolProposal{
		Title:       title,
		Description: description,
		Messages:    msgs,
	}, nil
}

// GetTitle returns the title of a community pool proposal.
func (p CommunityPoolProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of a community pool proposal.
func (p CommunityPoolProposal) GetDescription() string { return p.Description }

// GetDescription returns the routing key of a community pool proposal.
func (p CommunityPoolProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool proposal.
func (p CommunityPoolProposal) ProposalType() string {
	return ProposalTypeCommunityPool
}

// ValidateBasic stateless validation of a community pool proposal.
func (p CommunityPoolProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	msgs, err := p.GetMsgs()
	if err != nil {
		return sdkerrors.Wrap(ErrInvalidProposal, err.Error())
	}
	for _, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(ErrInvalidProposal, err.Error())
		}
	}
	return nil
}

// String implements fmt.Stringer
func (p CommunityPoolProposal) String() string {
	msgsList := ""
	msgs, err := p.GetMsgs()
	if err != nil {
		panic(err)
	}
	for _, msg := range msgs {
		msgsList += msg.String()
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Proposal:
  Title:       %s
  Description: %s
  messages:   %s
`, p.Title, p.Description, msgsList))
	return b.String()
}

func (p CommunityPoolProposal) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	for _, any := range p.Messages {
		var msg sdk.Msg
		err := unpacker.UnpackAny(any, &msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p CommunityPoolProposal) GetMsgs() ([]sdk.Msg, error) {
	return unpackMsgs(p.Messages)
}

func unpackMsgs(anys []*cdctypes.Any) ([]sdk.Msg, error) {
	msgs := make([]sdk.Msg, len(anys))
	for i, any := range anys {
		cached := any.GetCachedValue()
		if cached == nil {
			return nil, fmt.Errorf("any cached value is nil, sdk.Msgs messages must be correctly packed any values")
		}
		msgs[i] = cached.(sdk.Msg)
	}
	return msgs, nil
}

func packMsgs(msgs []sdk.Msg) ([]*cdctypes.Any, error) {
	anys := make([]*cdctypes.Any, len(msgs))
	for i, msg := range msgs {
		var err error
		anys[i], err = cdctypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
	}
	return anys, nil
}
