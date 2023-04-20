package types

import (
	"errors"
	fmt "fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeCommunityPoolLendDeposit defines the type for a CommunityPoolLendDepositProposal
	ProposalTypeCommunityPoolLendDeposit = "CommunityPoolLendDeposit"
	// ProposalTypeCommunityPoolLendWithdraw defines the type for a CommunityPoolLendDepositProposal
	ProposalTypeCommunityPoolLendWithdraw = "CommunityPoolLendWithdraw"
	// ProposalTypeCommunityCDPRepayDebt defines the type for a CommunityCDPRepayDebtProposal
	ProposalTypeCommunityCDPRepayDebt = "CommunityCDPRepayDebt"
	// ProposalTypeCommunityCDPWithdrawCollateral defines the type for a CommunityCDPWithdrawCollateralProposal
	ProposalTypeCommunityCDPWithdrawCollateral = "CommunityCDPWithdrawCollateral"
)

// Assert CommunityPoolLendDepositProposal implements govtypes.Content at compile-time
var (
	_ govv1beta1.Content = &CommunityPoolLendDepositProposal{}
	_ govv1beta1.Content = &CommunityPoolLendWithdrawProposal{}
	_ govv1beta1.Content = &CommunityCDPRepayDebtProposal{}
	_ govv1beta1.Content = &CommunityCDPWithdrawCollateralProposal{}
)

func init() {
	govv1beta1.RegisterProposalType(ProposalTypeCommunityPoolLendDeposit)
	govv1beta1.ModuleCdc.Amino.RegisterConcrete(&CommunityPoolLendDepositProposal{}, "kava/CommunityPoolLendDepositProposal", nil)
	govv1beta1.RegisterProposalType(ProposalTypeCommunityPoolLendWithdraw)
	govv1beta1.ModuleCdc.Amino.RegisterConcrete(&CommunityPoolLendWithdrawProposal{}, "kava/CommunityPoolLendWithdrawProposal", nil)
	govv1beta1.RegisterProposalType(ProposalTypeCommunityCDPRepayDebt)
	govv1beta1.ModuleCdc.Amino.RegisterConcrete(&CommunityCDPRepayDebtProposal{}, "kava/CommunityCDPRepayDebtProposal", nil)
	govv1beta1.RegisterProposalType(ProposalTypeCommunityCDPWithdrawCollateral)
	govv1beta1.ModuleCdc.Amino.RegisterConcrete(&CommunityCDPWithdrawCollateralProposal{}, "kava/CommunityCDPWithdrawCollateralProposal", nil)
}

//////////////////
// Lend Proposals
//////////////////

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
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	// ensure the proposal has valid amount
	if !p.Amount.IsValid() || p.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "deposit amount %s", p.Amount)
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
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	// ensure the proposal has valid amount
	if !p.Amount.IsValid() || p.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "withdraw amount %s", p.Amount)
	}
	return p.Amount.Validate()
}

/////////////////
// CDP Proposals
/////////////////

// NewCommunityCDPRepayDebtProposal creates a new community pool cdp debt repay proposal.
func NewCommunityCDPRepayDebtProposal(
	title string,
	description string,
	collateralType string,
	payment sdk.Coin,
) *CommunityCDPRepayDebtProposal {
	return &CommunityCDPRepayDebtProposal{
		Title:          title,
		Description:    description,
		CollateralType: collateralType,
		Payment:        payment,
	}
}

// GetTitle returns the title of the proposal.
func (p *CommunityCDPRepayDebtProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of the proposal.
func (p *CommunityCDPRepayDebtProposal) GetDescription() string { return p.Description }

// GetDescription returns the routing key of the proposal.
func (p *CommunityCDPRepayDebtProposal) ProposalRoute() string { return ModuleName }

// ProposalType returns the type of the proposal.
func (p *CommunityCDPRepayDebtProposal) ProposalType() string {
	return ProposalTypeCommunityCDPRepayDebt
}

// String implements fmt.Stringer
func (p *CommunityCDPRepayDebtProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community CDP Repay Debt Proposal:
  Title:           %s
  Description:     %s
  Collateral Type: %s
  Payment:         %s
`, p.Title, p.Description, p.CollateralType, p.Payment))
	return b.String()
}

// ValidateBasic stateless validation of the proposal.
func (p *CommunityCDPRepayDebtProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	// ensure collateral type is set
	if strings.TrimSpace(p.CollateralType) == "" {
		return errors.New("cdp collateral type cannot be blank")
	}
	// ensure the proposal has payment amount
	if !p.Payment.IsValid() || p.Payment.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "payment amount %s", p.Payment)
	}
	return nil
}

// NewCommunityCDPRepayDebtProposal creates a new community pool cdp debt repay proposal.
func NewCommunityCDPWithdrawCollateralProposal(
	title string,
	description string,
	collateralType string,
	collateral sdk.Coin,
) *CommunityCDPWithdrawCollateralProposal {
	return &CommunityCDPWithdrawCollateralProposal{
		Title:          title,
		Description:    description,
		CollateralType: collateralType,
		Collateral:     collateral,
	}
}

// GetTitle returns the title of the proposal.
func (p *CommunityCDPWithdrawCollateralProposal) GetTitle() string { return p.Title }

// GetDescription returns the description of the proposal.
func (p *CommunityCDPWithdrawCollateralProposal) GetDescription() string { return p.Description }

// GetDescription returns the routing key of the proposal.
func (p *CommunityCDPWithdrawCollateralProposal) ProposalRoute() string { return ModuleName }

// ProposalType returns the type of the proposal.
func (p *CommunityCDPWithdrawCollateralProposal) ProposalType() string {
	return ProposalTypeCommunityCDPWithdrawCollateral
}

// String implements fmt.Stringer
func (p *CommunityCDPWithdrawCollateralProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community CDP Withdraw Collateral Proposal:
  Title:           %s
  Description:     %s
  Collateral Type: %s
  Collateral:      %s
`, p.Title, p.Description, p.CollateralType, p.Collateral))
	return b.String()
}

// ValidateBasic stateless validation of the proposal.
func (p *CommunityCDPWithdrawCollateralProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}

	// ensure collateral type is set
	if strings.TrimSpace(p.CollateralType) == "" {
		return errors.New("cdp collateral type cannot be blank")
	}

	// ensure the proposal has collateral amount
	if !p.Collateral.IsValid() || p.Collateral.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "collateral amount %s", p.Collateral)
	}
	return nil
}
