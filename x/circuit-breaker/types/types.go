package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type MsgRoute struct {
	Route string
	Msg   string // how best to store a Msg type?
}

const (
	ProposalTypeCircuitBreak = "CircuitBreak"
)

// Assert CircuitBreakProposal implements govtypes.Content at compile-time
var _ govtypes.Content = CircuitBreakProposal{}

type CircuitBreakProposal struct {
	Title       string
	Description string
	MsgRoutes   []MsgRoute
}

// GetTitle returns the title of a community pool spend proposal.
func (cbp CircuitBreakProposal) GetTitle() string { return cbp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (cbp CircuitBreakProposal) GetDescription() string { return cbp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (cbp CircuitBreakProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (cbp CircuitBreakProposal) ProposalType() string { return ProposalTypeCircuitBreak }

// ValidateBasic runs basic stateless validity checks
func (cbp CircuitBreakProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, cbp)
	if err != nil {
		return err
	}
	// TODO
	return nil
}

// String implements the Stringer interface.
func (cbp CircuitBreakProposal) String() string {
	// TODO
}

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	// ModuleName is the module name constant used in many places
	ModuleName = "circuit-breaker"

	// RouterKey is the message route for distribution
	RouterKey = ModuleName
)
