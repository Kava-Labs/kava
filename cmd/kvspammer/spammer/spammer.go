package spammer

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	amino "github.com/tendermint/go-amino"

	"github.com/kava-labs/kava/cmd/kvspammer/txs"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

var msg []sdk.Msg

// SpamTxCDP sends a transaction to the CDP module
func SpamTxCDP(
	rpcURL string,
	chainID string,
	from string,
	passphrase string,
	collateralDenom string,
	principalDenom string,
	maxCollateral int64,
	appCodec *amino.Codec,
	cliCtx *context.CLIContext,
	accAddress sdk.AccAddress,
) error {
	// Set up new source of randomness
	randSource := rand.New(rand.NewSource(int64(time.Now().Unix())))

	// Attempt to locate existing CDP for this user/collateral denom
	cdp, found, err := txs.QueryCDP(appCodec, cliCtx, accAddress, collateralDenom)
	if err != nil {
		return err
	}

	// If no existing CDP is found, create new CDP
	if !found {
		// Create collateral and principal coin
		collateralAmount := sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, int(maxCollateral))))
		collateral := sdk.NewCoin(collateralDenom, collateralAmount)
		principal := sdk.NewCoin(principalDenom, collateralAmount.Quo(sdk.NewInt(2)))
		fmt.Printf("Creating new CDP. Collateral: %s, Principal: %s...\n", collateral, principal)
		msg = []sdk.Msg{cdptypes.NewMsgCreateCDP(accAddress, sdk.NewCoins(collateral), sdk.NewCoins(principal))}
	} else {
		// Calculate a random percentage and load current CDP values for coin amount generation
		randPercentage := sdk.NewDec(int64(simulation.RandIntBetween(randSource, 1, 25))).QuoInt(sdk.NewInt(100)) // TODO: calculate off % split
		currCollateral, currDebt, collateralizationRatio := loadCDPData(cdp)

		// If collateralization ratio above limit, withdraw colllateral or draw principal
		// If collateralization ratio is below limit, deposit collateral or repay principal
		if collateralizationRatio.GTE(sdk.NewDec(int64(220)).QuoInt(sdk.NewInt(100))) { // TODO: Parameterize 220%
			if randSource.Int63()%2 == 0 {
				// Generate random amount of coins
				coin := sdk.NewCoin(collateralDenom, sdk.NewInt(randPercentage.Mul(currCollateral).Int64()))
				// Build sdk.Msg
				msg = []sdk.Msg{cdptypes.NewMsgWithdraw(accAddress, accAddress, sdk.NewCoins(coin))}
				fmt.Printf("Attempting to withdraw %s collateral...\n", coin)
			} else {
				// Generate random amount of coins
				coin := sdk.NewCoin(principalDenom, randPercentage.MulInt(currDebt).TruncateInt())
				// Build sdk.Msg
				msg = []sdk.Msg{cdptypes.NewMsgDrawDebt(accAddress, collateralDenom, sdk.NewCoins(coin))}
				fmt.Printf("Attempting to draw %s principal...\n", coin)
			}
		} else {
			if randSource.Int63()%2 == 0 {
				// Deposit collateral 1-20%
				coin := sdk.NewCoin(collateralDenom, sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, 20))))
				fmt.Printf("Attempting to deposit %s collateral...\n", coin)
				msg = []sdk.Msg{cdptypes.NewMsgDeposit(accAddress, accAddress, sdk.NewCoins(coin))}
			} else {
				// Generate random amount of coins
				coin := sdk.NewCoin(principalDenom, randPercentage.MulInt(currDebt).TruncateInt())
				// Build sdk.Msg
				msg = []sdk.Msg{cdptypes.NewMsgRepayDebt(accAddress, principalDenom, sdk.NewCoins(coin))}
				fmt.Printf("Attempting to repay %s principal...\n", coin)
			}
		}
	}

	// Send tx containing the msg
	txRes, sdkErr := txs.SendTxRPC(chainID, appCodec, accAddress, from, passphrase, *cliCtx, msg, rpcURL)
	if sdkErr != nil {
		return sdkErr
	}

	fmt.Println("Tx hash:", txRes.TxHash)
	fmt.Println("Tx logs:", txRes.Logs)

	return nil

}

func loadCDPData(cdp cdptypes.CDP) (sdk.Dec, sdk.Int, sdk.Dec) {
	fmt.Println("Current CDP:", cdp)

	currCollateral := sdk.NewDec(int64(0))
	currPrincipal := sdk.NewInt(0)
	currFees := sdk.NewInt(0)

	// Error checking in case any value is empty
	if len(cdp.Collateral) > 0 {
		currCollateral = sdk.NewDec(cdp.Collateral[0].Amount.Int64())
	}
	if len(cdp.Principal) > 0 {
		currPrincipal = cdp.Principal[0].Amount
	}
	if len(cdp.AccumulatedFees) > 0 {
		currFees = cdp.AccumulatedFees[0].Amount
	}

	// Calculate current debt = principal + fees as sdk.Int
	currDebt := currPrincipal.Add(currFees)

	// Handle edge case: divide by 0
	var collateralizationRatio sdk.Dec
	if currDebt.IsPositive() {
		collateralizationRatio = currCollateral.QuoInt(currDebt)
	} else {
		// There is no principal or fees, CDP has excess collateral
		collateralizationRatio = sdk.NewDec(int64(1000))
	}
	fmt.Printf("\tCollateralization ratio: %s\n", collateralizationRatio.String())

	return currCollateral, currDebt, collateralizationRatio
}
