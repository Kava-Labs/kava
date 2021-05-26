package types

import (
	"bytes"

	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeCommitteeChange = "CommitteeChange"
	ProposalTypeCommitteeDelete = "CommitteeDelete"
)

// ProposalOutcome indicates the status of a proposal when it's closed and deleted from the store
type ProposalOutcome uint64

const (
	// Passed indicates that the proposal passed and was succesfully enacted
	Passed ProposalOutcome = iota
	// Failed indicates that the proposal failed and was not enacted
	Failed
	// Invalid indicates that proposal passed but an error occured when attempting to enact it
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

func (p ProposalOutcome) Marshal(cdc *codec.Codec) ([]byte, error) {
	x, err := cdc.MarshalJSON(p.String())
	if err != nil {
		return []byte{}, err
	}
	return x[1 : len(x)-1], nil
}

func MatchMarshaledOutcome(value []byte, cdc *codec.Codec) (ProposalOutcome, error) {
	passed, err := Passed.Marshal(cdc)
	if err != nil {
		return 0, err
	}
	if bytes.Compare(passed, value) == 0 {
		return Passed, nil
	}
	failed, err := Failed.Marshal(cdc)
	if err != nil {
		return 0, err
	}
	if bytes.Compare(failed, value) == 0 {
		return Failed, nil
	}
	invalid, err := Invalid.Marshal(cdc)
	if err != nil {
		return 0, err
	}
	if bytes.Compare(invalid, value) == 0 {
		return Invalid, nil
	}
	return 0, nil
}

// ensure proposal types fulfill the PubProposal interface and the gov Content interface.
var _, _ govtypes.Content = CommitteeChangeProposal{}, CommitteeDeleteProposal{}
var _, _ PubProposal = CommitteeChangeProposal{}, CommitteeDeleteProposal{}

func init() {
	// Gov proposals need to be registered on gov's ModuleCdc so MsgSubmitProposal can be encoded.
	govtypes.RegisterProposalType(ProposalTypeCommitteeChange)
	govtypes.RegisterProposalTypeCodec(CommitteeChangeProposal{}, "kava/CommitteeChangeProposal")

	govtypes.RegisterProposalType(ProposalTypeCommitteeDelete)
	govtypes.RegisterProposalTypeCodec(CommitteeDeleteProposal{}, "kava/CommitteeDeleteProposal")
}

// CommitteeChangeProposal is a gov proposal for creating a new committee or modifying an existing one.
type CommitteeChangeProposal struct {
	Title        string    `json:"title" yaml:"title"`
	Description  string    `json:"description" yaml:"description"`
	NewCommittee Committee `json:"new_committee" yaml:"new_committee"`
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

// ProposalRoute returns the routing key of the proposal.
func (ccp CommitteeChangeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal.
func (ccp CommitteeChangeProposal) ProposalType() string { return ProposalTypeCommitteeChange }

// ValidateBasic runs basic stateless validity checks
func (ccp CommitteeChangeProposal) ValidateBasic() error {
	if err := govtypes.ValidateAbstract(ccp); err != nil {
		return err
	}
	if err := ccp.NewCommittee.Validate(); err != nil {
		return sdkerrors.Wrap(ErrInvalidCommittee, err.Error())
	}
	return nil
}

// String implements the Stringer interface.
func (ccp CommitteeChangeProposal) String() string {
	bz, _ := yaml.Marshal(ccp)
	return string(bz)
}

// CommitteeDeleteProposal is a gov proposal for removing a committee.
type CommitteeDeleteProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	CommitteeID uint64 `json:"committee_id" yaml:"committee_id"`
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
	return govtypes.ValidateAbstract(cdp)
}

// String implements the Stringer interface.
func (cdp CommitteeDeleteProposal) String() string {
	bz, _ := yaml.Marshal(cdp)
	return string(bz)
}
