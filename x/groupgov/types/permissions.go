package types

import (
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	cbtypes "github.com/kava-labs/kava/x/circuit-breaker/types"
)

// EXAMPLE PERMISSIONS ------------------------------

// Allow only changes to inflation_rate
type InflationRateChangePermission struct{}

func (InflationRateChangePermission) Allows(p gov.Proposal) bool {
	pcp, _ := p.Content.(params.ParameterChangeProposal)
	for _, pc := range pcp.Changes {
		if pc.Key == "inflation_rate" {
			return true
		}
	}
	return false
}

// Allow only circuit breaking of the CDP Deposit msg
type CircuitBreakCDPDepsitPermission struct{}

func (CircuitBreakCDPDepsitPermission) Allows(p gov.Proposal) bool {
	cbp, _ := p.Content.(cbtypes.CircuitBreakProposal)
	for _, r := range cbp.MsgRoutes {
		if r.Route == "cdp" && r.Msg == "MsgCDPDeposit" {
			return true
		}
	}
	return false
}

// Same as above but the route the permssion allows can be set
type CircuitBreakPermission struct {
	MsgRoute cbtypes.MsgRoute
}

func (perm CircuitBreakPermission) Allows(p gov.Proposal) bool {
	cbp, _ := p.Content.(cbtypes.CircuitBreakProposal)
	for _, r := range cbp.MsgRoutes {
		if r == perm.MsgRoute {
			return true
		}
	}
	return false
}
