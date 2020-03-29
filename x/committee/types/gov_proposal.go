package types

import (
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeCommitteeChange = "CommitteeChange"
	ProposalTypeCommitteeDelete = "CommitteeDelete"
)

// CommitteeChangeProposal is a gov proposal for creating a new committee or modifying an existing one.
type CommitteeChangeProposal struct {
	Title        string    `json:"title" yaml:"title"`
	Description  string    `json:"description" yaml:"description"`
	NewCommittee Committee `json:"new_committee" yaml:"new_committee"`
}

var _ govtypes.Content = CommitteeChangeProposal{}

func init() {
	// Gov proposals need to be registered on gov's ModuleCdc so MsgSubmitProposal can be encoded.
	govtypes.RegisterProposalType(ProposalTypeCommitteeChange)
	govtypes.RegisterProposalTypeCodec(CommitteeChangeProposal{}, "kava/CommitteeChangeProposal")
}

func NewCommitteeChangeProposal(title string, description string, newCommittee Committee) CommitteeChangeProposal {
	return CommitteeChangeProposal{
		Title:        title,
		Description:  description,
		NewCommittee: newCommittee,
	}
}

// GetTitle returns the title of the proposal.
func (ccp CommitteeChangeProposal) GetTitle() string { return ccp.Title }

// GetDescription returns the description of the proposal.
func (ccp CommitteeChangeProposal) GetDescription() string { return ccp.Description }

// GetDescription returns the routing key of the proposal.
func (ccp CommitteeChangeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal.
func (ccp CommitteeChangeProposal) ProposalType() string { return ProposalTypeCommitteeChange }

// ValidateBasic runs basic stateless validity checks
func (ccp CommitteeChangeProposal) ValidateBasic() sdk.Error {
	if err := govtypes.ValidateAbstract(DefaultCodespace, ccp); err != nil {
		return err
	}
	if err := ccp.NewCommittee.Validate(); err != nil {
		return ErrInvalidCommittee(DefaultCodespace, err.Error())
	}
	return nil
}

// String implements the Stringer interface.
func (ccp CommitteeChangeProposal) String() string {
	bz, _ := yaml.Marshal(ccp) // TODO test
	return string(bz)
}

// CommitteeDeleteProposal is a gov proposal for removing a committee.
type CommitteeDeleteProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	CommitteeID uint64 `json:"committee_id" yaml:"committee_id"`
}

var _ govtypes.Content = CommitteeDeleteProposal{}

func init() {
	// Gov proposals need to be registered on gov's ModuleCdc so MsgSubmitProposal can be encoded.
	govtypes.RegisterProposalType(ProposalTypeCommitteeDelete)
	govtypes.RegisterProposalTypeCodec(CommitteeDeleteProposal{}, "kava/CommitteeDeleteProposal")
}

func NewCommitteeDeleteProposal(title string, description string, committeeID uint64) CommitteeDeleteProposal {
	return CommitteeDeleteProposal{
		Title:       title,
		Description: description,
		CommitteeID: committeeID,
	}
}

// GetTitle returns the title of the proposal.
func (cdp CommitteeDeleteProposal) GetTitle() string { return cdp.Title }

// GetDescription returns the description of the proposal.
func (cdp CommitteeDeleteProposal) GetDescription() string { return cdp.Description }

// GetDescription returns the routing key of the proposal.
func (cdp CommitteeDeleteProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal.
func (cdp CommitteeDeleteProposal) ProposalType() string { return ProposalTypeCommitteeDelete }

// ValidateBasic runs basic stateless validity checks
func (cdp CommitteeDeleteProposal) ValidateBasic() sdk.Error {
	if err := govtypes.ValidateAbstract(DefaultCodespace, cdp); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (cdp CommitteeDeleteProposal) String() string {
	bz, _ := yaml.Marshal(cdp) // TODO test
	return string(bz)
}
