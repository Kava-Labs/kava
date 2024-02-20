package keeper

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/validator-vesting/types"
)

type queryServer struct {
	bk types.BankKeeper
}

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(bk types.BankKeeper) types.QueryServer {
	return &queryServer{bk: bk}
}

// CirculatingSupply implements the gRPC service handler for querying the circulating supply of the kava token.
func (s queryServer) CirculatingSupply(c context.Context, req *types.QueryCirculatingSupplyRequest) (*types.
	QueryCirculatingSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	totalSupply := s.bk.GetSupply(ctx, "ukava").Amount
	supplyInt := getCirculatingSupply(ctx.BlockTime(), totalSupply)
	return &types.QueryCirculatingSupplyResponse{
		Amount: supplyInt,
	}, nil
}

// TotalSupply returns the total amount of ukava tokens
func (s queryServer) TotalSupply(c context.Context, req *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	totalSupply := s.bk.GetSupply(ctx, "ukava").Amount
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt()
	return &types.QueryTotalSupplyResponse{
		Amount: supplyInt,
	}, nil
}

// CirculatingSupplyHARD returns the total amount of hard tokens in circulation
func (s queryServer) CirculatingSupplyHARD(c context.Context, req *types.QueryCirculatingSupplyHARDRequest) (*types.QueryCirculatingSupplyHARDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

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
		circSupply = sdkmath.NewInt(0)
	case blockTime.After(supplyIncreaseDates[0]) && blockTime.Before(supplyIncreaseDates[1]) || blockTime.Equal(supplyIncreaseDates[0]):
		circSupply = sdkmath.NewInt(30000000) // Start year ONE
	case blockTime.After(supplyIncreaseDates[1]) && blockTime.Before(supplyIncreaseDates[2]) || blockTime.Equal(supplyIncreaseDates[1]):
		circSupply = sdkmath.NewInt(35000000)
	case blockTime.After(supplyIncreaseDates[2]) && blockTime.Before(supplyIncreaseDates[3]) || blockTime.Equal(supplyIncreaseDates[2]):
		circSupply = sdkmath.NewInt(40000000)
	case blockTime.After(supplyIncreaseDates[3]) && blockTime.Before(supplyIncreaseDates[4]) || blockTime.Equal(supplyIncreaseDates[3]):
		circSupply = sdkmath.NewInt(47708334)
	case blockTime.After(supplyIncreaseDates[4]) && blockTime.Before(supplyIncreaseDates[5]) || blockTime.Equal(supplyIncreaseDates[4]):
		circSupply = sdkmath.NewInt(51041667)
	case blockTime.After(supplyIncreaseDates[5]) && blockTime.Before(supplyIncreaseDates[6]) || blockTime.Equal(supplyIncreaseDates[5]):
		circSupply = sdkmath.NewInt(54375000)
	case blockTime.After(supplyIncreaseDates[6]) && blockTime.Before(supplyIncreaseDates[7]) || blockTime.Equal(supplyIncreaseDates[6]):
		circSupply = sdkmath.NewInt(61250000)
	case blockTime.After(supplyIncreaseDates[7]) && blockTime.Before(supplyIncreaseDates[8]) || blockTime.Equal(supplyIncreaseDates[7]):
		circSupply = sdkmath.NewInt(63750000)
	case blockTime.After(supplyIncreaseDates[8]) && blockTime.Before(supplyIncreaseDates[9]) || blockTime.Equal(supplyIncreaseDates[8]):
		circSupply = sdkmath.NewInt(66250000)
	case blockTime.After(supplyIncreaseDates[9]) && blockTime.Before(supplyIncreaseDates[10]) || blockTime.Equal(supplyIncreaseDates[9]):
		circSupply = sdkmath.NewInt(73125000)
	case blockTime.After(supplyIncreaseDates[10]) && blockTime.Before(supplyIncreaseDates[11]) || blockTime.Equal(supplyIncreaseDates[10]):
		circSupply = sdkmath.NewInt(75625000)
	case blockTime.After(supplyIncreaseDates[11]) && blockTime.Before(supplyIncreaseDates[12]) || blockTime.Equal(supplyIncreaseDates[11]):
		circSupply = sdkmath.NewInt(78125000) // End year ONE
	case blockTime.After(supplyIncreaseDates[12]) && blockTime.Before(supplyIncreaseDates[13]) || blockTime.Equal(supplyIncreaseDates[12]):
		circSupply = sdkmath.NewInt(91666667) // Start year TWO
	case blockTime.After(supplyIncreaseDates[13]) && blockTime.Before(supplyIncreaseDates[14]) || blockTime.Equal(supplyIncreaseDates[13]):
		circSupply = sdkmath.NewInt(94166667)
	case blockTime.After(supplyIncreaseDates[14]) && blockTime.Before(supplyIncreaseDates[15]) || blockTime.Equal(supplyIncreaseDates[14]):
		circSupply = sdkmath.NewInt(96666667)
	case blockTime.After(supplyIncreaseDates[15]) && blockTime.Before(supplyIncreaseDates[16]) || blockTime.Equal(supplyIncreaseDates[15]):
		circSupply = sdkmath.NewInt(105208334)
	case blockTime.After(supplyIncreaseDates[16]) && blockTime.Before(supplyIncreaseDates[17]) || blockTime.Equal(supplyIncreaseDates[16]):
		circSupply = sdkmath.NewInt(107708334)
	case blockTime.After(supplyIncreaseDates[17]) && blockTime.Before(supplyIncreaseDates[18]) || blockTime.Equal(supplyIncreaseDates[17]):
		circSupply = sdkmath.NewInt(110208334)
	case blockTime.After(supplyIncreaseDates[18]) && blockTime.Before(supplyIncreaseDates[19]) || blockTime.Equal(supplyIncreaseDates[18]):
		circSupply = sdkmath.NewInt(118750000)
	case blockTime.After(supplyIncreaseDates[19]) && blockTime.Before(supplyIncreaseDates[20]) || blockTime.Equal(supplyIncreaseDates[19]):
		circSupply = sdkmath.NewInt(121250000)
	case blockTime.After(supplyIncreaseDates[20]) && blockTime.Before(supplyIncreaseDates[21]) || blockTime.Equal(supplyIncreaseDates[20]):
		circSupply = sdkmath.NewInt(123750000)
	case blockTime.After(supplyIncreaseDates[21]) && blockTime.Before(supplyIncreaseDates[22]) || blockTime.Equal(supplyIncreaseDates[21]):
		circSupply = sdkmath.NewInt(132291668)
	case blockTime.After(supplyIncreaseDates[22]) && blockTime.Before(supplyIncreaseDates[23]) || blockTime.Equal(supplyIncreaseDates[22]):
		circSupply = sdkmath.NewInt(134791668)
	case blockTime.After(supplyIncreaseDates[23]) && blockTime.Before(supplyIncreaseDates[24]) || blockTime.Equal(supplyIncreaseDates[23]):
		circSupply = sdkmath.NewInt(137291668) // End year TWO
	case blockTime.After(supplyIncreaseDates[24]) && blockTime.Before(supplyIncreaseDates[25]) || blockTime.Equal(supplyIncreaseDates[24]):
		circSupply = sdkmath.NewInt(145833335) // Start year THREE
	case blockTime.After(supplyIncreaseDates[25]) && blockTime.Before(supplyIncreaseDates[26]) || blockTime.Equal(supplyIncreaseDates[25]):
		circSupply = sdkmath.NewInt(148333335)
	case blockTime.After(supplyIncreaseDates[26]) && blockTime.Before(supplyIncreaseDates[27]) || blockTime.Equal(supplyIncreaseDates[26]):
		circSupply = sdkmath.NewInt(150833335)
	case blockTime.After(supplyIncreaseDates[27]) && blockTime.Before(supplyIncreaseDates[28]) || blockTime.Equal(supplyIncreaseDates[27]):
		circSupply = sdkmath.NewInt(155000000)
	case blockTime.After(supplyIncreaseDates[28]) && blockTime.Before(supplyIncreaseDates[29]) || blockTime.Equal(supplyIncreaseDates[28]):
		circSupply = sdkmath.NewInt(157500000)
	case blockTime.After(supplyIncreaseDates[29]) && blockTime.Before(supplyIncreaseDates[30]) || blockTime.Equal(supplyIncreaseDates[29]):
		circSupply = sdkmath.NewInt(160000000)
	case blockTime.After(supplyIncreaseDates[30]) && blockTime.Before(supplyIncreaseDates[31]) || blockTime.Equal(supplyIncreaseDates[30]):
		circSupply = sdkmath.NewInt(164166669)
	case blockTime.After(supplyIncreaseDates[31]) && blockTime.Before(supplyIncreaseDates[32]) || blockTime.Equal(supplyIncreaseDates[31]):
		circSupply = sdkmath.NewInt(166666669)
	case blockTime.After(supplyIncreaseDates[32]) && blockTime.Before(supplyIncreaseDates[33]) || blockTime.Equal(supplyIncreaseDates[32]):
		circSupply = sdkmath.NewInt(169166669)
	case blockTime.After(supplyIncreaseDates[33]) && blockTime.Before(supplyIncreaseDates[34]) || blockTime.Equal(supplyIncreaseDates[33]):
		circSupply = sdkmath.NewInt(173333336)
	case blockTime.After(supplyIncreaseDates[34]) && blockTime.Before(supplyIncreaseDates[35]) || blockTime.Equal(supplyIncreaseDates[34]):
		circSupply = sdkmath.NewInt(175833336)
	case blockTime.After(supplyIncreaseDates[35]) && blockTime.Before(supplyIncreaseDates[36]) || blockTime.Equal(supplyIncreaseDates[35]):
		circSupply = sdkmath.NewInt(178333336) // End year THREE
	case blockTime.After(supplyIncreaseDates[36]) && blockTime.Before(supplyIncreaseDates[37]) || blockTime.Equal(supplyIncreaseDates[36]):
		circSupply = sdkmath.NewInt(181666670) // Start year FOUR
	case blockTime.After(supplyIncreaseDates[37]) && blockTime.Before(supplyIncreaseDates[38]) || blockTime.Equal(supplyIncreaseDates[37]):
		circSupply = sdkmath.NewInt(183333337)
	case blockTime.After(supplyIncreaseDates[38]) && blockTime.Before(supplyIncreaseDates[39]) || blockTime.Equal(supplyIncreaseDates[38]):
		circSupply = sdkmath.NewInt(185000000)
	case blockTime.After(supplyIncreaseDates[39]) && blockTime.Before(supplyIncreaseDates[40]) || blockTime.Equal(supplyIncreaseDates[39]):
		circSupply = sdkmath.NewInt(186666670)
	case blockTime.After(supplyIncreaseDates[40]) && blockTime.Before(supplyIncreaseDates[41]) || blockTime.Equal(supplyIncreaseDates[40]):
		circSupply = sdkmath.NewInt(188333338)
	case blockTime.After(supplyIncreaseDates[41]) && blockTime.Before(supplyIncreaseDates[42]) || blockTime.Equal(supplyIncreaseDates[41]):
		circSupply = sdkmath.NewInt(190000000)
	case blockTime.After(supplyIncreaseDates[42]) && blockTime.Before(supplyIncreaseDates[43]) || blockTime.Equal(supplyIncreaseDates[42]):
		circSupply = sdkmath.NewInt(191666670)
	case blockTime.After(supplyIncreaseDates[43]) && blockTime.Before(supplyIncreaseDates[44]) || blockTime.Equal(supplyIncreaseDates[43]):
		circSupply = sdkmath.NewInt(193333339)
	case blockTime.After(supplyIncreaseDates[44]) && blockTime.Before(supplyIncreaseDates[45]) || blockTime.Equal(supplyIncreaseDates[44]):
		circSupply = sdkmath.NewInt(195000000)
	case blockTime.After(supplyIncreaseDates[45]) && blockTime.Before(supplyIncreaseDates[46]) || blockTime.Equal(supplyIncreaseDates[45]):
		circSupply = sdkmath.NewInt(196666670)
	case blockTime.After(supplyIncreaseDates[46]) && blockTime.Before(supplyIncreaseDates[47]) || blockTime.Equal(supplyIncreaseDates[46]):
		circSupply = sdkmath.NewInt(198333340)
	case blockTime.After(supplyIncreaseDates[47]) && blockTime.Before(supplyIncreaseDates[48]) || blockTime.Equal(supplyIncreaseDates[47]):
		circSupply = sdkmath.NewInt(200000000) // End year FOUR
	default:
		circSupply = sdkmath.NewInt(200000000)
	}

	return &types.QueryCirculatingSupplyHARDResponse{
		Amount: circSupply,
	}, nil
}

// CirculatingSupplyUSDX returns the total amount of usdx tokens in circulation
func (s queryServer) CirculatingSupplyUSDX(c context.Context, req *types.QueryCirculatingSupplyUSDXRequest) (*types.QueryCirculatingSupplyUSDXResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	totalSupply := s.bk.GetSupply(ctx, "usdx").Amount
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt()
	return &types.QueryCirculatingSupplyUSDXResponse{
		Amount: supplyInt,
	}, nil
}

// CirculatingSupplySWP returns the total amount of swp tokens in circulation
func (s queryServer) CirculatingSupplySWP(c context.Context, req *types.QueryCirculatingSupplySWPRequest) (*types.QueryCirculatingSupplySWPResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

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
	teamSwp := int64(4_687_500)
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

		circSupply = circSupply.Add(sdkmath.NewInt(monthTotal))
	}

	return &types.QueryCirculatingSupplySWPResponse{
		Amount: circSupply,
	}, nil
}

// TotalSupplyHARD returns the total amount of hard tokens
func (s queryServer) TotalSupplyHARD(c context.Context, req *types.QueryTotalSupplyHARDRequest) (*types.QueryTotalSupplyHARDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	totalSupply := s.bk.GetSupply(ctx, "hard").Amount
	supplyInt := sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt()
	return &types.QueryTotalSupplyHARDResponse{
		Amount: supplyInt,
	}, nil
}

// TotalSupplyUSDX returns the total amount of usdx tokens
func (s queryServer) TotalSupplyUSDX(c context.Context, req *types.QueryTotalSupplyUSDXRequest) (*types.QueryTotalSupplyUSDXResponse, error) {
	// USDX total supply is the circulating supply
	rsp, err := s.CirculatingSupplyUSDX(c, &types.QueryCirculatingSupplyUSDXRequest{})
	if err != nil {
		return nil, err
	}
	return &types.QueryTotalSupplyUSDXResponse{
		Amount: rsp.Amount,
	}, nil
}

func getCirculatingSupply(blockTime time.Time, totalSupply sdkmath.Int) sdkmath.Int {
	vestingDates := []time.Time{
		time.Date(2022, 2, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 5, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 8, 5, 14, 0, 0, 0, time.UTC),
		time.Date(2022, 11, 5, 14, 0, 0, 0, time.UTC),
	}

	switch {
	case blockTime.Before(vestingDates[0]):
		return sdk.NewDecFromInt(totalSupply.Sub(sdkmath.NewInt(9937500000000))).Mul(sdk.MustNewDecFromStr("0.000001")).RoundInt()
	case blockTime.After(vestingDates[0]) && blockTime.Before(vestingDates[1]) || blockTime.Equal(vestingDates[0]):
		return sdk.NewDecFromInt(totalSupply.Sub(sdkmath.NewInt(7453125000000))).Mul(sdk.MustNewDecFromStr("0.000001")).RoundInt()
	case blockTime.After(vestingDates[1]) && blockTime.Before(vestingDates[2]) || blockTime.Equal(vestingDates[1]):
		return sdk.NewDecFromInt(totalSupply.Sub(sdkmath.NewInt(4968750000000))).Mul(sdk.MustNewDecFromStr("0.000001")).RoundInt()
	case blockTime.After(vestingDates[2]) && blockTime.Before(vestingDates[3]) || blockTime.Equal(vestingDates[2]):
		return sdk.NewDecFromInt(totalSupply.Sub(sdkmath.NewInt(2484375000000))).Mul(sdk.MustNewDecFromStr("0.000001")).RoundInt()
	default:
		// align with total supply calculation and truncate int here instead of round
		return sdk.NewDecFromInt(totalSupply).Mul(sdk.MustNewDecFromStr("0.000001")).TruncateInt()
	}
}
