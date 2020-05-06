package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
)

type HashTestSuite struct {
	suite.Suite
	addrs      []sdk.AccAddress
	timestamps []int64
}

func (suite *HashTestSuite) SetupTest() {
	// Generate 10 addresses
	_, addrs := app.GeneratePrivKeyAddressPairs(10)

	// Generate 10 timestamps
	var timestamps []int64
	for i := 0; i < 10; i++ {
		timestamps = append(timestamps, ts(i))
	}

	suite.addrs = addrs
	suite.timestamps = timestamps
	return
}

func (suite *HashTestSuite) TestGenerateSecureRandomNumber() {
	secureRandomNumber, err := types.GenerateSecureRandomNumber()
	suite.Nil(err)
	suite.NotNil(secureRandomNumber)
	suite.Equal(64, len(secureRandomNumber))
}

func (suite *HashTestSuite) TestCalculateRandomHash() {
	randomNumber, _ := types.GenerateSecureRandomNumber()
	hash := types.CalculateRandomHash(randomNumber, suite.timestamps[0])
	suite.NotNil(hash)
	suite.Equal(32, len(hash))
}

func (suite *HashTestSuite) TestCalculateSwapID() {
	randomNumber, _ := types.GenerateSecureRandomNumber()
	hash := types.CalculateRandomHash(randomNumber, suite.timestamps[3])
	swapID := types.CalculateSwapID(hash, suite.addrs[3], suite.addrs[5].String())
	suite.NotNil(swapID)
	suite.Equal(32, len(swapID))

	diffHash := types.CalculateRandomHash(randomNumber, suite.timestamps[2])
	diffSwapID := types.CalculateSwapID(diffHash, suite.addrs[3], suite.addrs[5].String())
	suite.NotEqual(swapID, diffSwapID)
}

func TestHashTestSuite(t *testing.T) {
	suite.Run(t, new(HashTestSuite))
}
