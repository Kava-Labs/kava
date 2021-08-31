package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/validator-vesting/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

const SafuFund int64 = 10000000 // 10 million KAVA

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryCirculatingSupply:
			return queryGetCirculatingSupply(ctx, req, keeper)
		case types.QueryTotalSupply:
			return queryGetTotalSupply(ctx, req, keeper)
		case types.QueryCirculatingSupplyHARD:
			return getCirculatingSupplyHARD(ctx, req, keeper)
		case types.QueryCirculatingSupplyUSDX:
			return getCirculatingSupplyUSDX(ctx, req, keeper)
		case types.QueryCirculatingSupplySWP:
			return getCirculatingSupplySWP(ctx, req, keeper)
		case types.QueryTotalSupplyHARD:
			return getTotalSupplyHARD(ctx, req, keeper)
		case types.QueryTotalSupplyUSDX:
			return getCirculatingSupplyUSDX(ctx, req, keeper) // Intentional - USDX total supply is the circulating supply
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

func queryGetTotalSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	totalSupply := keeper.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("ukava")
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt64()
	bz, err := types.ModuleCdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

func queryGetCirculatingSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	supplyInt := getCirculatingSupply(ctx.BlockTime())
	bz, err := keeper.cdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

func getCirculatingSupply(blockTime time.Time) sdk.Int {
	vestingDates := []time.Time{
		time.Date(2020, 9, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2020, 11, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2021, 2, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2021, 5, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2021, 8, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2021, 11, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 5, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 8, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 11, 5, 14, 0, 0, 0, time.UTC),
	}

	switch {
	case blockTime.Before(vestingDates[0]):
		return sdk.NewInt(27190672)
	case blockTime.After(vestingDates[0]) && blockTime.Before(vestingDates[1]) || blockTime.Equal(vestingDates[0]):
		return sdk.NewInt(29442227)
	case blockTime.After(vestingDates[1]) && blockTime.Before(vestingDates[2]) || blockTime.Equal(vestingDates[1]):
		return sdk.NewInt(46876230)
	case blockTime.After(vestingDates[2]) && blockTime.Before(vestingDates[3]) || blockTime.Equal(vestingDates[2]):
		return sdk.NewInt(58524186)
	case blockTime.After(vestingDates[3]) && blockTime.Before(vestingDates[4]) || blockTime.Equal(vestingDates[3]):
		safuFundInitTime := time.Date(2021, 6, 14, 14, 0, 0, 0, time.UTC)
		safuFundFilledTime := time.Date(2021, 7, 14, 14, 0, 0, 0, time.UTC)
		switch {
		case blockTime.Before(safuFundInitTime):
			return sdk.NewInt(70172142)
		case blockTime.After(safuFundInitTime) && blockTime.Before(safuFundFilledTime):
			days := blockTime.Sub(safuFundInitTime).Hours() / 24
			currSafuFundAmt := int64(days) * (SafuFund / 30)
			return sdk.NewInt(70172142 + currSafuFundAmt)
		default:
			return sdk.NewInt(70172142 + SafuFund)
		}
	case blockTime.After(vestingDates[4]) && blockTime.Before(vestingDates[5]) || blockTime.Equal(vestingDates[4]):
		return sdk.NewInt(81443180 + SafuFund)
	case blockTime.After(vestingDates[5]) && blockTime.Before(vestingDates[6]) || blockTime.Equal(vestingDates[5]):
		return sdk.NewInt(90625000 + SafuFund)
	case blockTime.After(vestingDates[6]) && blockTime.Before(vestingDates[7]) || blockTime.Equal(vestingDates[6]):
		return sdk.NewInt(92968750 + SafuFund)
	case blockTime.After(vestingDates[7]) && blockTime.Before(vestingDates[8]) || blockTime.Equal(vestingDates[7]):
		return sdk.NewInt(95312500 + SafuFund)
	case blockTime.After(vestingDates[8]) && blockTime.Before(vestingDates[9]) || blockTime.Equal(vestingDates[8]):
		return sdk.NewInt(97656250 + SafuFund)
	default:
		return sdk.NewInt(100000000)
	}
}

func getCirculatingSupplyHARD(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	supplyIncreaseDates := []time.Time{
		time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC), // + 30,000,000 *** Year ONE ***
		time.Date(2020, 11, 15, 14, 0, 0, 0, time.UTC), // + 5,000,000
		time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC), // + 5,000,000
		time.Date(2021, 1, 15, 14, 0, 0, 0, time.UTC),  // + 7,708,334
		time.Date(2021, 2, 15, 14, 0, 0, 0, time.UTC),  // + 3,333,333
		time.Date(2021, 3, 15, 14, 0, 0, 0, time.UTC),  // + 3,333,333
		time.Date(2021, 4, 15, 14, 0, 0, 0, time.UTC),  // + 6,875,000
		time.Date(2021, 5, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2021, 6, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2021, 7, 15, 14, 0, 0, 0, time.UTC),  // + 6,875,000
		time.Date(2021, 8, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2021, 9, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000  *** Year ONE ***
		time.Date(2021, 10, 15, 14, 0, 0, 0, time.UTC), // + 13,541,667 *** Year TWO ***
		time.Date(2021, 11, 15, 14, 0, 0, 0, time.UTC), // + 2,500,000
		time.Date(2021, 12, 15, 14, 0, 0, 0, time.UTC), // + 2,500,000
		time.Date(2022, 1, 15, 14, 0, 0, 0, time.UTC),  // + 8,541,667
		time.Date(2022, 2, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2022, 3, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2022, 4, 15, 14, 0, 0, 0, time.UTC),  // + 8,541,667
		time.Date(2022, 5, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2022, 6, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2022, 7, 15, 14, 0, 0, 0, time.UTC),  // + 8,541,667
		time.Date(2022, 8, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2022, 9, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000  *** Year TWO ***
		time.Date(2022, 10, 15, 14, 0, 0, 0, time.UTC), // + 8,541,667  *** Year THREE ***
		time.Date(2022, 11, 15, 14, 0, 0, 0, time.UTC), // + 2,500,000
		time.Date(2022, 12, 15, 14, 0, 0, 0, time.UTC), // + 2,500,000
		time.Date(2023, 1, 15, 14, 0, 0, 0, time.UTC),  // + 4,166,667
		time.Date(2023, 2, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2023, 3, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2023, 4, 15, 14, 0, 0, 0, time.UTC),  // + 4,166,667
		time.Date(2023, 5, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2023, 6, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2023, 7, 15, 14, 0, 0, 0, time.UTC),  // + 4,166,667
		time.Date(2023, 8, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000
		time.Date(2023, 9, 15, 14, 0, 0, 0, time.UTC),  // + 2,500,000  *** Year THREE ***
		time.Date(2023, 10, 15, 14, 0, 0, 0, time.UTC), // + 3,333,334  *** Year FOUR ***
		time.Date(2023, 11, 15, 14, 0, 0, 0, time.UTC), // + 1,666,667
		time.Date(2023, 12, 15, 14, 0, 0, 0, time.UTC), // + 1,666,667
		time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667
		time.Date(2024, 2, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667
		time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667
		time.Date(2024, 4, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667
		time.Date(2024, 5, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667
		time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667
		time.Date(2024, 7, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667
		time.Date(2024, 8, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667
		time.Date(2024, 9, 15, 14, 0, 0, 0, time.UTC),  // + 1,666,667  *** Year FOUR ***
	}

	circSupply := sdk.ZeroInt()
	blockTime := ctx.BlockTime()
	switch {
	case blockTime.Before(supplyIncreaseDates[0]):
		circSupply = sdk.NewInt(0)
	case blockTime.After(supplyIncreaseDates[0]) && blockTime.Before(supplyIncreaseDates[1]) || blockTime.Equal(supplyIncreaseDates[0]):
		circSupply = sdk.NewInt(30000000) // Start year ONE
	case blockTime.After(supplyIncreaseDates[1]) && blockTime.Before(supplyIncreaseDates[2]) || blockTime.Equal(supplyIncreaseDates[1]):
		circSupply = sdk.NewInt(35000000)
	case blockTime.After(supplyIncreaseDates[2]) && blockTime.Before(supplyIncreaseDates[3]) || blockTime.Equal(supplyIncreaseDates[2]):
		circSupply = sdk.NewInt(40000000)
	case blockTime.After(supplyIncreaseDates[3]) && blockTime.Before(supplyIncreaseDates[4]) || blockTime.Equal(supplyIncreaseDates[3]):
		circSupply = sdk.NewInt(47708334)
	case blockTime.After(supplyIncreaseDates[4]) && blockTime.Before(supplyIncreaseDates[5]) || blockTime.Equal(supplyIncreaseDates[4]):
		circSupply = sdk.NewInt(51041667)
	case blockTime.After(supplyIncreaseDates[5]) && blockTime.Before(supplyIncreaseDates[6]) || blockTime.Equal(supplyIncreaseDates[5]):
		circSupply = sdk.NewInt(54375000)
	case blockTime.After(supplyIncreaseDates[6]) && blockTime.Before(supplyIncreaseDates[7]) || blockTime.Equal(supplyIncreaseDates[6]):
		circSupply = sdk.NewInt(61250000)
	case blockTime.After(supplyIncreaseDates[7]) && blockTime.Before(supplyIncreaseDates[8]) || blockTime.Equal(supplyIncreaseDates[7]):
		circSupply = sdk.NewInt(63750000)
	case blockTime.After(supplyIncreaseDates[8]) && blockTime.Before(supplyIncreaseDates[9]) || blockTime.Equal(supplyIncreaseDates[8]):
		circSupply = sdk.NewInt(66250000)
	case blockTime.After(supplyIncreaseDates[9]) && blockTime.Before(supplyIncreaseDates[10]) || blockTime.Equal(supplyIncreaseDates[9]):
		circSupply = sdk.NewInt(73125000)
	case blockTime.After(supplyIncreaseDates[10]) && blockTime.Before(supplyIncreaseDates[11]) || blockTime.Equal(supplyIncreaseDates[10]):
		circSupply = sdk.NewInt(75625000)
	case blockTime.After(supplyIncreaseDates[11]) && blockTime.Before(supplyIncreaseDates[12]) || blockTime.Equal(supplyIncreaseDates[11]):
		circSupply = sdk.NewInt(78125000) // End year ONE
	case blockTime.After(supplyIncreaseDates[12]) && blockTime.Before(supplyIncreaseDates[13]) || blockTime.Equal(supplyIncreaseDates[12]):
		circSupply = sdk.NewInt(91666667) // Start year TWO
	case blockTime.After(supplyIncreaseDates[13]) && blockTime.Before(supplyIncreaseDates[14]) || blockTime.Equal(supplyIncreaseDates[13]):
		circSupply = sdk.NewInt(94166667)
	case blockTime.After(supplyIncreaseDates[14]) && blockTime.Before(supplyIncreaseDates[15]) || blockTime.Equal(supplyIncreaseDates[14]):
		circSupply = sdk.NewInt(96666667)
	case blockTime.After(supplyIncreaseDates[15]) && blockTime.Before(supplyIncreaseDates[16]) || blockTime.Equal(supplyIncreaseDates[15]):
		circSupply = sdk.NewInt(105208334)
	case blockTime.After(supplyIncreaseDates[16]) && blockTime.Before(supplyIncreaseDates[17]) || blockTime.Equal(supplyIncreaseDates[16]):
		circSupply = sdk.NewInt(107708334)
	case blockTime.After(supplyIncreaseDates[17]) && blockTime.Before(supplyIncreaseDates[18]) || blockTime.Equal(supplyIncreaseDates[17]):
		circSupply = sdk.NewInt(110208334)
	case blockTime.After(supplyIncreaseDates[18]) && blockTime.Before(supplyIncreaseDates[19]) || blockTime.Equal(supplyIncreaseDates[18]):
		circSupply = sdk.NewInt(118750000)
	case blockTime.After(supplyIncreaseDates[19]) && blockTime.Before(supplyIncreaseDates[20]) || blockTime.Equal(supplyIncreaseDates[19]):
		circSupply = sdk.NewInt(121250000)
	case blockTime.After(supplyIncreaseDates[20]) && blockTime.Before(supplyIncreaseDates[21]) || blockTime.Equal(supplyIncreaseDates[20]):
		circSupply = sdk.NewInt(123750000)
	case blockTime.After(supplyIncreaseDates[21]) && blockTime.Before(supplyIncreaseDates[22]) || blockTime.Equal(supplyIncreaseDates[21]):
		circSupply = sdk.NewInt(132291668)
	case blockTime.After(supplyIncreaseDates[22]) && blockTime.Before(supplyIncreaseDates[23]) || blockTime.Equal(supplyIncreaseDates[22]):
		circSupply = sdk.NewInt(134791668)
	case blockTime.After(supplyIncreaseDates[23]) && blockTime.Before(supplyIncreaseDates[24]) || blockTime.Equal(supplyIncreaseDates[23]):
		circSupply = sdk.NewInt(137291668) // End year TWO
	case blockTime.After(supplyIncreaseDates[24]) && blockTime.Before(supplyIncreaseDates[25]) || blockTime.Equal(supplyIncreaseDates[24]):
		circSupply = sdk.NewInt(145833335) // Start year THREE
	case blockTime.After(supplyIncreaseDates[25]) && blockTime.Before(supplyIncreaseDates[26]) || blockTime.Equal(supplyIncreaseDates[25]):
		circSupply = sdk.NewInt(148333335)
	case blockTime.After(supplyIncreaseDates[26]) && blockTime.Before(supplyIncreaseDates[27]) || blockTime.Equal(supplyIncreaseDates[26]):
		circSupply = sdk.NewInt(150833335)
	case blockTime.After(supplyIncreaseDates[27]) && blockTime.Before(supplyIncreaseDates[28]) || blockTime.Equal(supplyIncreaseDates[27]):
		circSupply = sdk.NewInt(155000000)
	case blockTime.After(supplyIncreaseDates[28]) && blockTime.Before(supplyIncreaseDates[29]) || blockTime.Equal(supplyIncreaseDates[28]):
		circSupply = sdk.NewInt(157500000)
	case blockTime.After(supplyIncreaseDates[29]) && blockTime.Before(supplyIncreaseDates[30]) || blockTime.Equal(supplyIncreaseDates[29]):
		circSupply = sdk.NewInt(160000000)
	case blockTime.After(supplyIncreaseDates[30]) && blockTime.Before(supplyIncreaseDates[31]) || blockTime.Equal(supplyIncreaseDates[30]):
		circSupply = sdk.NewInt(164166669)
	case blockTime.After(supplyIncreaseDates[31]) && blockTime.Before(supplyIncreaseDates[32]) || blockTime.Equal(supplyIncreaseDates[31]):
		circSupply = sdk.NewInt(166666669)
	case blockTime.After(supplyIncreaseDates[32]) && blockTime.Before(supplyIncreaseDates[33]) || blockTime.Equal(supplyIncreaseDates[32]):
		circSupply = sdk.NewInt(169166669)
	case blockTime.After(supplyIncreaseDates[33]) && blockTime.Before(supplyIncreaseDates[34]) || blockTime.Equal(supplyIncreaseDates[33]):
		circSupply = sdk.NewInt(173333336)
	case blockTime.After(supplyIncreaseDates[34]) && blockTime.Before(supplyIncreaseDates[35]) || blockTime.Equal(supplyIncreaseDates[34]):
		circSupply = sdk.NewInt(175833336)
	case blockTime.After(supplyIncreaseDates[35]) && blockTime.Before(supplyIncreaseDates[36]) || blockTime.Equal(supplyIncreaseDates[35]):
		circSupply = sdk.NewInt(178333336) // End year THREE
	case blockTime.After(supplyIncreaseDates[36]) && blockTime.Before(supplyIncreaseDates[37]) || blockTime.Equal(supplyIncreaseDates[36]):
		circSupply = sdk.NewInt(181666670) // Start year FOUR
	case blockTime.After(supplyIncreaseDates[37]) && blockTime.Before(supplyIncreaseDates[38]) || blockTime.Equal(supplyIncreaseDates[37]):
		circSupply = sdk.NewInt(183333337)
	case blockTime.After(supplyIncreaseDates[38]) && blockTime.Before(supplyIncreaseDates[39]) || blockTime.Equal(supplyIncreaseDates[38]):
		circSupply = sdk.NewInt(185000000)
	case blockTime.After(supplyIncreaseDates[39]) && blockTime.Before(supplyIncreaseDates[40]) || blockTime.Equal(supplyIncreaseDates[39]):
		circSupply = sdk.NewInt(186666670)
	case blockTime.After(supplyIncreaseDates[40]) && blockTime.Before(supplyIncreaseDates[41]) || blockTime.Equal(supplyIncreaseDates[40]):
		circSupply = sdk.NewInt(188333338)
	case blockTime.After(supplyIncreaseDates[41]) && blockTime.Before(supplyIncreaseDates[42]) || blockTime.Equal(supplyIncreaseDates[41]):
		circSupply = sdk.NewInt(190000000)
	case blockTime.After(supplyIncreaseDates[42]) && blockTime.Before(supplyIncreaseDates[43]) || blockTime.Equal(supplyIncreaseDates[42]):
		circSupply = sdk.NewInt(191666670)
	case blockTime.After(supplyIncreaseDates[43]) && blockTime.Before(supplyIncreaseDates[44]) || blockTime.Equal(supplyIncreaseDates[43]):
		circSupply = sdk.NewInt(193333339)
	case blockTime.After(supplyIncreaseDates[44]) && blockTime.Before(supplyIncreaseDates[45]) || blockTime.Equal(supplyIncreaseDates[44]):
		circSupply = sdk.NewInt(195000000)
	case blockTime.After(supplyIncreaseDates[45]) && blockTime.Before(supplyIncreaseDates[46]) || blockTime.Equal(supplyIncreaseDates[45]):
		circSupply = sdk.NewInt(196666670)
	case blockTime.After(supplyIncreaseDates[46]) && blockTime.Before(supplyIncreaseDates[47]) || blockTime.Equal(supplyIncreaseDates[46]):
		circSupply = sdk.NewInt(198333340)
	case blockTime.After(supplyIncreaseDates[47]) && blockTime.Before(supplyIncreaseDates[48]) || blockTime.Equal(supplyIncreaseDates[47]):
		circSupply = sdk.NewInt(200000000) // End year FOUR
	default:
		circSupply = sdk.NewInt(200000000)
	}

	bz, err := keeper.cdc.MarshalJSON(circSupply)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

func getCirculatingSupplyUSDX(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	totalSupply := keeper.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("usdx")
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt64()
	bz, err := types.ModuleCdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

func getCirculatingSupplySWP(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Start values
	year := 2021
	month := 8

	var supplyIncreaseDates []time.Time

	// Add month times for 4 years
	for i := 0; i < 12*4; i++ {
		// Always day 30 unless it's Feb
		day := 30
		if month == 2 {
			day = 28
		}

		date := time.Date(year, time.Month(month), day, 15 /* hour */, 0, 0, 0, time.UTC)
		supplyIncreaseDates = append(supplyIncreaseDates, date)

		// Update year and month
		if month == 12 {
			month = 1
			year += 1
		} else {
			month += 1
		}
	}

	// Repeated tokens released
	teamSwp := int64(5_859_375)
	treasurySwp := int64(5_859_375)
	monthlyStakersSwp := int64(520_833)
	monthlyLPIncentivesSwp := int64(2_343_750)

	// []{Ecosystem, Team, Treasury, Kava Stakers, LP Incentives}
	scheduleAmounts := [][]int64{
		{12_500_000, 0, 15_625_000, monthlyStakersSwp, monthlyLPIncentivesSwp},  // *** Year ONE ***
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 1
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 2
		{0, 0, treasurySwp, monthlyStakersSwp, monthlyLPIncentivesSwp},          // 3
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 4
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 5
		{0, 0, treasurySwp, monthlyStakersSwp, monthlyLPIncentivesSwp},          // 6
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 7
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 8
		{0, 0, treasurySwp, monthlyStakersSwp, monthlyLPIncentivesSwp},          // 9
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 10
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 11
		{0, 18_750_000, treasurySwp, monthlyStakersSwp, monthlyLPIncentivesSwp}, // *** Year TWO ***
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 13
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 14
		{0, teamSwp, treasurySwp, monthlyStakersSwp, monthlyLPIncentivesSwp},    // 15
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 16
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 17
		{0, teamSwp, treasurySwp, monthlyStakersSwp, monthlyLPIncentivesSwp},    // 18
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 19
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 20
		{0, teamSwp, treasurySwp, monthlyStakersSwp, monthlyLPIncentivesSwp},    // 21
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 22
		{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp},                    // 23
		{0, teamSwp, treasurySwp, monthlyStakersSwp, monthlyLPIncentivesSwp},    // *** Year THREE ***
	}

	// Months 25-47 are the same
	for i := 0; i < 23; i++ {
		scheduleAmounts = append(scheduleAmounts, []int64{0, 0, 0, monthlyStakersSwp, monthlyLPIncentivesSwp})
	}

	circSupply := sdk.ZeroInt()
	blockTime := ctx.BlockTime()

	for i := 0; i < len(scheduleAmounts); i++ {
		if blockTime.Before(supplyIncreaseDates[i]) {
			break
		}

		// Sum up each category of token release
		monthTotal := int64(0)
		for _, val := range scheduleAmounts[i] {
			monthTotal += val
		}

		circSupply = circSupply.Add(sdk.NewInt(monthTotal))
	}

	bz, err := keeper.cdc.MarshalJSON(circSupply)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

func getTotalSupplyHARD(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	totalSupply := keeper.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf("hard")
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt64()
	bz, err := types.ModuleCdc.MarshalJSON(supplyInt)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}
