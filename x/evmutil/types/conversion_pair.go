package types

import (
	bytes "bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

///////////////
// EVM -> Cosmos SDK
///////////////

// NewConversionPair returns a new ConversionPair.
func NewConversionPair(address InternalEVMAddress, denom string) ConversionPair {
	return ConversionPair{
		KavaERC20Address: address.Address.Bytes(),
		Denom:            denom,
	}
}

// GetAddress returns the InternalEVMAddress of the Kava ERC20 address.
func (pair ConversionPair) GetAddress() InternalEVMAddress {
	return NewInternalEVMAddress(common.BytesToAddress(pair.KavaERC20Address))
}

// Validate returns an error if the ConversionPair is invalid.
func (pair ConversionPair) Validate() error {
	if err := sdk.ValidateDenom(pair.Denom); err != nil {
		return fmt.Errorf("conversion pair denom invalid: %v", err)
	}

	if len(pair.KavaERC20Address) != common.AddressLength {
		return fmt.Errorf("address length is %v but expected %v", len(pair.KavaERC20Address), common.AddressLength)
	}

	if bytes.Equal(pair.KavaERC20Address, common.Address{}.Bytes()) {
		return fmt.Errorf("address cannot be zero value %v", hex.EncodeToString(pair.KavaERC20Address))
	}

	return nil
}

// ConversionPairs defines a slice of ConversionPair.
type ConversionPairs []ConversionPair

// NewConversionPairs returns ConversionPairs from the provided values.
func NewConversionPairs(pairs ...ConversionPair) ConversionPairs {
	return ConversionPairs(pairs)
}

func (pairs ConversionPairs) Validate() error {
	// Check for duplicates for both addrs and denoms
	addrs := map[string]bool{}
	denoms := map[string]bool{}

	for _, pair := range pairs {
		if addrs[hex.EncodeToString(pair.KavaERC20Address)] {
			return fmt.Errorf(
				"found duplicate enabled conversion pair internal ERC20 address %s",
				hex.EncodeToString(pair.KavaERC20Address),
			)
		}

		if denoms[pair.Denom] {
			return fmt.Errorf(
				"found duplicate enabled conversion pair denom %s",
				pair.Denom,
			)
		}

		if err := pair.Validate(); err != nil {
			return err
		}

		addrs[hex.EncodeToString(pair.KavaERC20Address)] = true
		denoms[pair.Denom] = true
	}

	return nil
}

// validateConversionPairs validates an interface as ConversionPairs
func validateConversionPairs(i interface{}) error {
	pairs, ok := i.(ConversionPairs)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return pairs.Validate()
}

///////////////
// Cosmos SDK -> EVM
///////////////

// NewAllowedNativeCoinERC20Token returns an AllowedNativeCoinERC20Token
func NewAllowedNativeCoinERC20Token(
	sdkDenom, name, symbol string,
	decimal uint32,
) AllowedNativeCoinERC20Token {
	return AllowedNativeCoinERC20Token{
		SdkDenom: sdkDenom,
		Name:     name,
		Symbol:   symbol,
		Decimals: decimal,
	}
}

// Validate validates the fields of a single AllowedNativeCoinERC20Token
func (token AllowedNativeCoinERC20Token) Validate() error {
	// disallow empty string fields
	if err := sdk.ValidateDenom(token.SdkDenom); err != nil {
		return fmt.Errorf("allowed native coin erc20 token's sdk denom is invalid: %v", err)
	}

	if token.Name == "" {
		return errors.New("allowed native coin erc20 token's name cannot be empty")
	}

	if token.Symbol == "" {
		return errors.New("allowed native coin erc20 token's symbol cannot be empty")
	}

	// ensure decimals will properly cast to uint8 of erc20 spec
	if token.Decimals > math.MaxUint8 {
		return fmt.Errorf("allowed native coin erc20 token's decimals must be less than 256, found %d", token.Decimals)
	}

	return nil
}

// AllowedNativeCoinERC20Tokens defines a slice of AllowedNativeCoinERC20Token
type AllowedNativeCoinERC20Tokens []AllowedNativeCoinERC20Token

// NewAllowedNativeCoinERC20Tokens returns AllowedNativeCoinERC20Tokens from the provided values.
func NewAllowedNativeCoinERC20Tokens(pairs ...AllowedNativeCoinERC20Token) AllowedNativeCoinERC20Tokens {
	return AllowedNativeCoinERC20Tokens(pairs)
}

// Validate checks that all containing tokens are valid and that there are
// no duplicate denoms or symbols.
func (tokens AllowedNativeCoinERC20Tokens) Validate() error {
	// Disallow multiple instances of a single sdk_denom or evm symbol
	denoms := make(map[string]struct{}, len(tokens))
	symbols := make(map[string]struct{}, len(tokens))

	for i, t := range tokens {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("invalid token at index %d: %s", i, err)
		}

		if _, found := denoms[t.SdkDenom]; found {
			return fmt.Errorf("found duplicate token with sdk denom %s", t.SdkDenom)
		}
		if _, found := symbols[t.Symbol]; found {
			return fmt.Errorf("found duplicate token with symbol %s", t.Symbol)
		}

		denoms[t.SdkDenom] = struct{}{}
		symbols[t.Symbol] = struct{}{}
	}

	return nil
}

// validateAllowedNativeCoinERC20Tokens validates an interface as AllowedNativeCoinERC20Tokens
func validateAllowedNativeCoinERC20Tokens(i interface{}) error {
	pairs, ok := i.(AllowedNativeCoinERC20Tokens)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return pairs.Validate()
}
