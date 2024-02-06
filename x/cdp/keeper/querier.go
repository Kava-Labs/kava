package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// FilterCDPs queries the store for all CDPs that match query params
func FilterCDPs(ctx sdk.Context, k Keeper, params types.QueryCdpsParams) (types.AugmentedCDPs, error) {
	var matchCollateralType, matchOwner, matchID, matchRatio types.CDPs

	// match cdp owner (if supplied)
	if len(params.Owner) > 0 {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			cdp, found := k.GetCdpByOwnerAndCollateralType(ctx, params.Owner, denom)
			if found {
				matchOwner = append(matchOwner, cdp)
			}
		}
	}

	// match cdp collateral denom (if supplied)
	if len(params.CollateralType) > 0 {
		// if owner is specified only iterate over already matched cdps for efficiency
		if len(params.Owner) > 0 {
			for _, cdp := range matchOwner {
				if cdp.Type == params.CollateralType {
					matchCollateralType = append(matchCollateralType, cdp)
				}
			}
		} else {
			_, found := k.GetCollateral(ctx, params.CollateralType)
			if !found {
				return nil, fmt.Errorf("invalid collateral type")
			}
			matchCollateralType = k.GetAllCdpsByCollateralType(ctx, params.CollateralType)
		}
	}

	// match cdp ID (if supplied)
	if params.ID != 0 {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			cdp, found := k.GetCDP(ctx, denom, params.ID)
			if found {
				matchID = append(matchID, cdp)
			}
		}
	}

	// match cdp ratio (if supplied)
	if !params.Ratio.IsNil() && params.Ratio.GT(sdk.ZeroDec()) {
		denoms := k.GetCollateralTypes(ctx)
		for _, denom := range denoms {
			ratio, err := k.CalculateCollateralizationRatioFromAbsoluteRatio(ctx, denom, params.Ratio, "liquidation")
			if err != nil {
				continue
			}
			cdpsUnderRatio := k.GetAllCdpsByCollateralTypeAndRatio(ctx, denom, ratio)
			matchRatio = append(matchRatio, cdpsUnderRatio...)
		}
	}

	var commonCDPs types.CDPs
	// If no params specified, fetch all CDPs
	if params.CollateralType == "" && len(params.Owner) == 0 && params.ID == 0 && params.Ratio.Equal(sdk.ZeroDec()) {
		commonCDPs = k.GetAllCdps(ctx)
	}

	// Find the intersection of any matched CDPs
	if params.CollateralType != "" {
		if len(matchCollateralType) == 0 {
			return nil, nil
		}

		commonCDPs = matchCollateralType
	}

	if len(params.Owner) > 0 {
		if len(matchCollateralType) > 0 {
			if len(commonCDPs) > 0 {
				commonCDPs = FindIntersection(commonCDPs, matchOwner)
			} else {
				commonCDPs = matchOwner
			}
		} else {
			commonCDPs = matchOwner
		}
	}

	if params.ID != 0 {
		if len(matchID) == 0 {
			return nil, nil
		}

		if len(commonCDPs) > 0 {
			commonCDPs = FindIntersection(commonCDPs, matchID)
		} else {
			commonCDPs = matchID
		}
	}

	if !params.Ratio.IsNil() && params.Ratio.GT(sdk.ZeroDec()) {
		if len(matchRatio) == 0 {
			return nil, nil
		}

		if len(commonCDPs) > 0 {
			commonCDPs = FindIntersection(commonCDPs, matchRatio)
		} else {
			commonCDPs = matchRatio
		}
	}

	// Load augmented CDPs
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range commonCDPs {
		augmentedCDP := k.LoadAugmentedCDP(ctx, cdp)
		augmentedCDPs = append(augmentedCDPs, augmentedCDP)
	}

	// Apply page and limit params
	start, end := client.Paginate(len(augmentedCDPs), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		return nil, nil
	}

	return augmentedCDPs[start:end], nil
}

// FindIntersection finds the intersection of two CDP arrays in linear time complexity O(n + n)
func FindIntersection(x types.CDPs, y types.CDPs) types.CDPs {
	cdpSet := make(types.CDPs, 0)
	cdpMap := make(map[uint64]bool)

	for i := 0; i < len(x); i++ {
		cdp := x[i]
		cdpMap[cdp.ID] = true
	}

	for i := 0; i < len(y); i++ {
		cdp := y[i]
		if _, found := cdpMap[cdp.ID]; found {
			cdpSet = append(cdpSet, cdp)
		}
	}

	return cdpSet
}
