package util

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

// EvmContractMethodId encodes a method signature to the method id used in eth calldata.
func EvmContractMethodId(signature string) []byte {
	transferFnSignature := []byte(signature)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	return hash.Sum(nil)[:4]
}

func BuildErc20TransferCallData(from common.Address, to common.Address, amount *big.Int) []byte {
	methodId := EvmContractMethodId("transfer(address,uint256)")
	paddedAddress := common.LeftPadBytes(to.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodId...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	return data
}

func BuildErc20BalanceOfCallData(address common.Address) []byte {
	methodId := EvmContractMethodId("balanceOf(address)")
	paddedAddress := common.LeftPadBytes(address.Bytes(), 32)

	var data []byte
	data = append(data, methodId...)
	data = append(data, paddedAddress...)

	return data
}
