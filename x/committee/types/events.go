package types

// Module event types
const (
	EventTypeProposalSubmit = "proposal_submit"
	EventTypeProposalClose  = "proposal_close"
	EventTypeProposalVote   = "proposal_vote"

	AttributeValueCategory          = "committee"
	AttributeKeyCommitteeID         = "committee_id"
	AttributeKeyProposalID          = "proposal_id"
	AttributeKeyDeadline            = "deadline"
	AttributeKeyProposalCloseStatus = "status"
	AttributeKeyVoter               = "voter"
	AttributeKeyVote                = "vote"
	AttributeKeyProposalOutcome     = "proposal_outcome"
	AttributeKeyProposalTally       = "proposal_tally"
)
