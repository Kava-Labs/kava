package types

import (
	"encoding/binary"
	"fmt"
	"strings"

	binance "github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func CalculateRandomHash(randomNumber []byte, timestamp int64) []byte {
	data := make([]byte, RandomNumberLength+Int64Size)
	copy(data[:RandomNumberLength], randomNumber)
	binary.BigEndian.PutUint64(data[RandomNumberLength:], uint64(timestamp))
	return tmhash.Sum(data)
}

func CalculateSwapID(randomNumberHash []byte, sender sdk.AccAddress, senderOtherChain string) []byte {
	senderOtherChain = strings.ToLower(senderOtherChain)
	data := randomNumberHash
	data = append(data, []byte(sender)...)
	data = append(data, []byte(senderOtherChain)...)
	return tmhash.Sum(data)
}

func ConvertBinanceAddress(addr binance.AccAddress) (sdk.AccAddress, sdk.Error) {
	sdkAddr, err := sdk.AccAddressFromBech32(addr.String())
	if err != nil {
		return sdk.AccAddress{}, sdk.ErrInvalidAddress(fmt.Sprintf("%s", err))
	}
	return sdkAddr, nil
}

func ConvertBinanceAddresses(binanceAddrs []binance.AccAddress) ([]sdk.AccAddress, sdk.Error) {
	var sdkAddrs []sdk.AccAddress
	for _, addr := range binanceAddrs {
		sdkAddr, err := sdk.AccAddressFromBech32(addr.String())
		if err != nil {
			return []sdk.AccAddress{}, sdk.ErrInvalidAddress(fmt.Sprintf("%s", err))
		}
		sdkAddrs = append(sdkAddrs, sdkAddr)
	}
	return sdkAddrs, nil
}

func ConvertBinanceCoin(coin binance.Coin) sdk.Coin {
	return sdk.NewInt64Coin(coin.Denom, coin.Amount)
}

func ConvertBinanceCoins(binanceCoins binance.Coins) sdk.Coins {
	var sdkCoins sdk.Coins
	for _, binanceCoin := range binanceCoins {
		sdkCoins = append(sdkCoins, ConvertBinanceCoin(binanceCoin))
	}
	return sdkCoins
}
