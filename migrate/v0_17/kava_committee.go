package v0_17

import (
	auctiontypes "github.com/kava-labs/kava/x/auction/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
)

func migrateCommitteePermissions(genState committeetypes.GenesisState) committeetypes.GenesisState {
	var newCommittees committeetypes.Committees
	for _, committee := range genState.GetCommittees() {
		switch committee.GetDescription() {
		case "Hard Governance Committee":
			committee = fixHardPermissions(committee)
		case "Kava Stability Committee":
			committee = fixStabilityPermissions(committee)
		}
		newCommittees = append(newCommittees, committee)
	}

	packedCommittees, err := committeetypes.PackCommittees(newCommittees)
	if err != nil {
		panic(err)
	}
	genState.Committees = packedCommittees
	return genState
}

func fixStabilityPermissions(committee committeetypes.Committee) committeetypes.Committee {
	permissions := committee.GetPermissions()

	// get first params change permission in committee
	var perm *committeetypes.ParamsChangePermission
	var permIndex int
	var found bool
	for permIndex = range permissions {
		p, ok := permissions[permIndex].(*committeetypes.ParamsChangePermission)
		if ok {
			perm = p
			found = true
			break
		}
	}
	if !found {
		panic("ParamsChangePermission not found")
	}

	appendSubparamRequirement(
		perm.AllowedParamsChanges,
		hardtypes.ModuleName, string(hardtypes.KeyMoneyMarkets),
		committeetypes.SubparamRequirement{
			Key: "denom",
			Val: "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			AllowedSubparamAttrChanges: []string{
				"borrow_limit",
				"interest_rate_model",
				"keeper_reward_percentage",
				"reserve_factor",
				"spot_market_id",
			},
		},
	)

	appendSubparamRequirement(
		perm.AllowedParamsChanges,
		cdptypes.ModuleName, string(cdptypes.KeyCollateralParams),
		committeetypes.SubparamRequirement{
			Key: "type",
			Val: "ust-a",
			AllowedSubparamAttrChanges: []string{
				"auction_size",
				"check_collateralization_index_count",
				"debt_limit",
				"keeper_reward_percentage",
				"stability_fee",
			},
		},
	)

	perm.AllowedParamsChanges.Delete(auctiontypes.ModuleName, "BidDuration")

	perm.AllowedParamsChanges.Set(committeetypes.AllowedParamsChange{
		Subspace: auctiontypes.ModuleName,
		Key:      string(auctiontypes.KeyForwardBidDuration),
	})
	perm.AllowedParamsChanges.Set(committeetypes.AllowedParamsChange{
		Subspace: auctiontypes.ModuleName,
		Key:      string(auctiontypes.KeyReverseBidDuration),
	})

	// update committee
	permissions[permIndex] = perm
	committee.SetPermissions(permissions)

	return committee
}

func fixHardPermissions(committee committeetypes.Committee) committeetypes.Committee {
	permissions := committee.GetPermissions()

	// get first params change permission in committee
	var perm *committeetypes.ParamsChangePermission
	var permIndex int
	var found bool
	for permIndex = range permissions {
		p, ok := permissions[permIndex].(*committeetypes.ParamsChangePermission)
		if ok {
			perm = p
			found = true
			break
		}
	}
	if !found {
		panic("ParamsChangePermission not found")
	}

	appendSubparamRequirement(
		perm.AllowedParamsChanges,
		hardtypes.ModuleName, string(hardtypes.KeyMoneyMarkets),
		committeetypes.SubparamRequirement{
			Key: "denom",
			Val: "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			AllowedSubparamAttrChanges: []string{
				"borrow_limit",
				"interest_rate_model",
				"keeper_reward_percentage",
				"reserve_factor",
				"spot_market_id",
			},
		},
	)

	// update committee
	permissions[permIndex] = perm
	committee.SetPermissions(permissions)

	return committee
}

func appendSubparamRequirement(allowed committeetypes.AllowedParamsChanges, subspace, key string, requirement committeetypes.SubparamRequirement) {
	apc, found := allowed.Get(subspace, key)
	if !found {
		panic("AllowedParamsChange not found")
	}
	apc.MultiSubparamsRequirements = append(apc.MultiSubparamsRequirements, requirement)
	allowed.Set(apc)
}
