package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// Convertion between native gas KAVA (6) to EVM (18).
// 12 decimal difference, so 1_000_000_000_000.
var conversionMultiplier = sdk.NewInt(1_000_000_000_000)

type EVMBankKeeper struct {
	bankKeeper bankkeeper.Keeper
}

var _ evmtypes.BankKeeper = (*EVMBankKeeper)(nil)

func NewEVMBankKeeper(bk bankkeeper.Keeper) EVMBankKeeper {
	return EVMBankKeeper{
		bankKeeper: bk,
	}
}

func (bk EVMBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	bal := bk.bankKeeper.GetBalance(ctx, addr, denom)

	return convertCoinToEvm(bal)
}

func (bk EVMBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	actualAmt := convertCoinsFromEvm(amt)

	return bk.bankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, actualAmt)
}

func (bk EVMBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	actualAmt := convertCoinsFromEvm(amt)

	return bk.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, actualAmt)
}

func (bk EVMBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	actualAmt := convertCoinsFromEvm(amt)

	return bk.bankKeeper.MintCoins(ctx, moduleName, actualAmt)
}

func (bk EVMBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	actualAmt := convertCoinsFromEvm(amt)

	return bk.bankKeeper.BurnCoins(ctx, moduleName, actualAmt)
}

// convertCoinToEvm converts a sdk.Coin with native decimals to an sdk.Coin with evm decimals
func convertCoinToEvm(coin sdk.Coin) sdk.Coin {
	coin.Amount = coin.Amount.Mul(conversionMultiplier)

	return coin
}

func convertCoinFromEvm(coin sdk.Coin) sdk.Coin {
	coin.Amount = coin.Amount.Quo(conversionMultiplier)

	return coin
}

func convertCoinsFromEvm(coins sdk.Coins) sdk.Coins {
	for _, coin := range coins {
		coin = convertCoinFromEvm(coin)
	}

	return coins
}
