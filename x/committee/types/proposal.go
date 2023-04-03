package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeCommitteeChange = "CommitteeChange"
	ProposalTypeCommitteeDelete = "CommitteeDelete"
)

// ProposalOutcome indicates the status of a proposal when it's closed and deleted from the store
type ProposalOutcome uint64

const (
	// Passed indicates that the proposal passed and was successfully enacted
	Passed ProposalOutcome = iota
	// Failed indicates that the proposal failed and was not enacted
	Failed
	// Invalid indicates that proposal passed but an error occurred when attempting to enact it
	Invalid
)

var toString = map[ProposalOutcome]string{
	Passed:  "Passed",
	Failed:  "Failed",
	Invalid: "Invalid",
}

func (p ProposalOutcome) String() string {
	return toString[p]
}

// ensure proposal types fulfill the PubProposal interface and the gov Content interface.
var _, _ govv1beta1.Content = &CommitteeChangeProposal{}, &CommitteeDeleteProposal{}
var _, _ PubProposal = &CommitteeChangeProposal{}, &CommitteeDeleteProposal{}

// ensure CommitteeChangeProposal fulfill the codectypes.UnpackInterfacesMessage interface
var _ codectypes.UnpackInterfacesMessage = &CommitteeChangeProposal{}

func init() {
	// Gov proposals need to be registered on gov's ModuleCdc so MsgSubmitProposal can be encoded.
	govv1beta1.RegisterProposalType(ProposalTypeCommitteeChange)
	govv1beta1.RegisterProposalType(ProposalTypeCommitteeDelete)
}

func NewCommitteeChangeProposal(title string, description string, newCommittee Committee) (CommitteeChangeProposal, error) {
	committeeAny, err := PackCommittee(newCommittee)
	if err != nil {
		return CommitteeChangeProposal{}, err
	}
	return CommitteeChangeProposal{
		Title:        title,
		Description:  description,
		NewCommittee: committeeAny,
	}, nil
}

func MustNewCommitteeChangeProposal(title string, description string, newCommittee Committee) CommitteeChangeProposal {
	proposal, err := NewCommitteeChangeProposal(title, description, newCommittee)
	if err != nil {
		panic(err)
	}
	return proposal
}

// GetTitle returns the title of the proposal.
func (ccp CommitteeChangeProposal) GetTitle() string { return ccp.Title }

// GetDescription returns the description of the proposal.
func (ccp CommitteeChangeProposal) GetDescription() string { return ccp.Description }

// ProposalRoute returns the routing key of the proposal.
func (ccp CommitteeChangeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal.
func (ccp CommitteeChangeProposal) ProposalType() string { return ProposalTypeCommitteeChange }

// GetNewCommittee returns the new committee of the proposal.
func (ccp CommitteeChangeProposal) GetNewCommittee() Committee {
	committee, err := UnpackCommittee(ccp.NewCommittee)
	if err != nil {
		panic(err)
	}
	return committee
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (ccp CommitteeChangeProposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var committee Committee
	return unpacker.UnpackAny(ccp.NewCommittee, &committee)
}

// ValidateBasic runs basic stateless validity checks
func (ccp CommitteeChangeProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(&ccp); err != nil {
		return err
	}
	committee, err := UnpackCommittee(ccp.NewCommittee)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidCommittee, err.Error())
	}
	if err := committee.Validate(); err != nil {
		return errorsmod.Wrap(ErrInvalidCommittee, err.Error())
	}
	return nil
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

// ProposalRoute returns the routing key of the proposal.
func (cdp CommitteeDeleteProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal.
func (cdp CommitteeDeleteProposal) ProposalType() string { return ProposalTypeCommitteeDelete }

// ValidateBasic runs basic stateless validity checks
func (cdp CommitteeDeleteProposal) ValidateBasic() error {
	return govv1beta1.ValidateAbstract(&cdp)
}
