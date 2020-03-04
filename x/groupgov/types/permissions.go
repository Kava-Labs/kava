package types

// EXAMPLE PERMISSIONS ------------------------------
type InflationRateChangePermission uint8

func (InflationRateChangePermission) Allows(p gov.Proposal) bool {
	pcp, _ := p.Content.(params.ParameterChangeProposal)
	for pc, _ := range pcp.Changes {
		if pc.Key == "inflation_rate" {
			return true
		}
	}
	return false
}

type CircuitBreakCDPDepsitPermission uint8

func (CircuitBreakCDPDepsitPermission) Allows(p gov.Proposal) bool {
	cbp, _ := p.Content.(CircuitBreakProposal)
	if cbp.Route == "cdp" && cbp.Msg == "MsgCDPDeposit" {
		return true
	}
	return false
}
