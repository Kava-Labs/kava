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

// CollateralizationRatioLimit set to 220%
const CollateralizationRatioLimit = 220

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
	cliCtx context.CLIContext,
	accAddress sdk.AccAddress,
) {

	// Set up new source of randomness
	randSource := rand.New(rand.NewSource(int64(time.Now().Unix())))

	// Attempt to locate existing CDP for this user/collateral denom
	cdp, found, err := txs.QueryCDP(appCodec, cliCtx, accAddress, collateralDenom)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(cdp)
		// Load current CDP values for coin amount generation
		currCollateral, currDebt, collateralizationRatio := loadCDPData(cdp)

		// Get random amount of collateral between 1-25% current collateral
		quarterCollateral := int(currCollateral.Int64()) / 4
		amtCollateral := sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, quarterCollateral)))

		// Get random amount of debt between 1-25% current debt
		quarterDebt := int(currDebt.Int64()) / 4
		amtDebt := sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, quarterDebt)))

		// If collateralization ratio above limit, withdraw colllateral or draw principal
		// If collateralization ratio is below limit, deposit collateral or repay principal
		if collateralizationRatio.GTE(sdk.NewDec(int64(CollateralizationRatioLimit)).QuoInt(sdk.NewInt(100))) {
			if randSource.Int63()%2 == 0 {
				coin := sdk.NewCoin(collateralDenom, amtCollateral)
				msg = []sdk.Msg{cdptypes.NewMsgWithdraw(accAddress, accAddress, sdk.NewCoins(coin))}
				fmt.Printf("\nAttempting to withdraw %s collateral...\n", coin)
			} else {
				coin := sdk.NewCoin(principalDenom, amtDebt)
				msg = []sdk.Msg{cdptypes.NewMsgDrawDebt(accAddress, collateralDenom, sdk.NewCoins(coin))}
				fmt.Printf("\nAttempting to draw %s principal...\n", coin)
			}
		} else {
			if randSource.Int63()%2 == 0 {
				coin := sdk.NewCoin(collateralDenom, amtCollateral)
				msg = []sdk.Msg{cdptypes.NewMsgDeposit(accAddress, accAddress, sdk.NewCoins(coin))}
				fmt.Printf("\nAttempting to deposit %s collateral...\n", coin)
			} else {
				coin := sdk.NewCoin(principalDenom, amtDebt)
				msg = []sdk.Msg{cdptypes.NewMsgRepayDebt(accAddress, collateralDenom, sdk.NewCoins(coin))}
				fmt.Printf("\nAttempting to repay %s principal...\n", coin)
			}
		}
	}

	// Send tx containing the msg
	txRes, sdkErr := txs.SendTxRPC(chainID, appCodec, accAddress, from, passphrase, cliCtx, msg, rpcURL)
	if sdkErr != nil {
		fmt.Println(err)
	}

	fmt.Println("Tx hash:", txRes.TxHash)
	fmt.Println("Tx logs:", txRes.Logs)
	fmt.Println()

}

func loadCDPData(cdp cdptypes.CDP) (sdk.Int, sdk.Int, sdk.Dec) {
	currCollateral := sdk.NewInt(0)
	currPrincipal := sdk.NewInt(0)
	currFees := sdk.NewInt(0)

	// Error checking in case any value is empty
	if len(cdp.Collateral) > 0 {
		currCollateral = cdp.Collateral[0].Amount
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
		collateralizationRatio = sdk.NewDec(currCollateral.Int64()).QuoInt(currDebt)
	} else {
		// There is no principal or fees, CDP has excess collateral
		collateralizationRatio = sdk.NewDec(int64(1000))
	}
	fmt.Printf("\tCollateralization ratio: %s\n", collateralizationRatio.String())

	return currCollateral, currDebt, collateralizationRatio
}
