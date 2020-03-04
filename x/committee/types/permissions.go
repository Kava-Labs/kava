package types

import (
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	sdtypes "github.com/kava-labs/kava/x/shutdown/types"
)

// EXAMPLE PERMISSIONS ------------------------------

// Allow only changes to inflation_rate
type InflationRateChangePermission struct{}

var _ types.Permission = InflationRateChangePermission

func (InflationRateChangePermission) Allows(p gov.Proposal) bool {
	pcp, ok := p.Content.(params.ParameterChangeProposal)
	if !ok {
		return false
	}
	for _, pc := range pcp.Changes {
		if pc.Key == "inflation_rate" {
			return true
		}
	}
	return false
}

// Allow only shutdown of the CDP Deposit msg
type ShutdownCDPDepsitPermission struct{}

var _ types.Permission = ShutdownCDPDepsitPermission

func (ShutdownCDPDepsitPermission) Allows(p gov.Content) bool {
	sdp, ok := p.(sdtypes.ShutdownProposal)
	if !ok {
		return false
	}
	for _, r := range sdp.MsgRoutes {
		if r.Route == "cdp" && r.Msg == "MsgCDPDeposit" {
			return true
		}
	}
	return false
}

// Same as above but the route isn't static
type GeneralShutdownPermission struct {
	MsgRoute cbtypes.MsgRoute
}

var _ types.Permission = GeneralShutdownPermission

func (perm GeneralShutdownPermission) Allows(p gov.Content) bool {
	sdp, ok := p.Content.(sdtypes.ShutdownProposal)
	if !ok {
		return false
	}
	for _, r := range sdp.MsgRoutes {
		if r == perm.MsgRoute {
			return true
		}
	}
	return false
}
