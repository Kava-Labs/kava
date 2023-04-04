package simulation

// import (
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// )

// func ShiftDec(x sdk.Dec, places sdkmath.Int) sdk.Dec {
// 	neg := places.IsNegative()
// 	for i := 0; i < int(abs(places.Int64())); i++ {
// 		if neg {
// 			x = x.Mul(sdk.MustNewDecFromStr("0.1"))
// 		} else {
// 			x = x.Mul(sdk.NewDecFromInt(sdkmath.NewInt(10)))
// 		}

// 	}
// 	return x
// }

// // abs returns the absolute value of x.
// func abs(x int64) int64 {
// 	if x < 0 {
// 		return -x
// 	}
// 	return x
// }
