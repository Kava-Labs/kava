package util

import (
	"context"
	"errors"
	"time"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func WaitForProposalStatus(
	govClient govtypesv1.QueryClient,
	proposalId uint64,
	status govtypesv1.ProposalStatus,
	timeout time.Duration,
) (uint64, error) {
	var err error
	var passedProposalId uint64
	outOfTime := time.After(timeout)
	for {
		select {
		case <-outOfTime:
			err = errors.New("timed out waiting for proposal to be passed")
		default:
			resp, proposalErr := govClient.Proposal(
				context.Background(),
				&govtypesv1.QueryProposalRequest{ProposalId: proposalId},
			)
			if proposalErr != nil {
				err = proposalErr
				break
			}
			if resp.Proposal.GetStatus() == status {
				passedProposalId = resp.Proposal.GetId()
				break
			}
			continue
		}
		break
	}
	return passedProposalId, err
}
