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
	ProposalTypeShutdown = "Shutdown"
)

// Assert ShutdownProposal implements govtypes.Content at compile-time
var _ govtypes.Content = ShutdownProposal{}

type ShutdownProposal struct {
	Title       string
	Description string
	MsgRoutes   []MsgRoute
}

// GetTitle returns the title of a community pool spend proposal.
func (sp ShutdownProposal) GetTitle() string { return sp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (sp ShutdownProposal) GetDescription() string { return sp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (sp ShutdownProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (sp ShutdownProposal) ProposalType() string { return ProposalTypeShutdown }

// ValidateBasic runs basic stateless validity checks
func (sp ShutdownProposal) ValidateBasic() sdk.Error {
	err := govtypes.ValidateAbstract(DefaultCodespace, sp)
	if err != nil {
		return err
	}
	// TODO
	return nil
}

// String implements the Stringer interface.
func (sp ShutdownProposal) String() string {
	// TODO
	return ""
}

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	// ModuleName is the module name constant used in many places
	ModuleName = "shutdown"

	// RouterKey is the message route for distribution
	RouterKey = ModuleName
)
