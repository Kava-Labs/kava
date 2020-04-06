package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/simulation"
	"github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/stretchr/testify/suite"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type decoderTest struct {
	name        string
	expectedLog string
}

type DecoderTestSuite struct {
	suite.Suite

	tests []decoderTest
}

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	types.RegisterCodec(cdc)
	return
}

func (suite *DecoderTestSuite) TestDecodeStore() {
	cdc := makeTestCodec()
	price := types.NewCurrentPrice("bnb:usd", sdk.MustNewDecFromStr("12.0"))
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	rawPrices := types.PostedPrices{
		types.NewPostedPrice("bnb:usd", addrs[0], sdk.MustNewDecFromStr("12.0"), tmtime.Now().Add(time.Hour*2)),
	}
	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: types.CurrentPriceKey("bnb:usd"), Value: cdc.MustMarshalBinaryBare(price)},
		cmn.KVPair{Key: types.RawPriceKey("bnb:usd"), Value: cdc.MustMarshalBinaryBare(rawPrices)},
	}

	decoderTests := []decoderTest{
		decoderTest{"current price", fmt.Sprintf("%s\n%s", price, price)},
		decoderTest{"raw prices", fmt.Sprintf("%s\n%s", rawPrices, rawPrices)},
	}

	for i, t := range decoderTests {
		suite.Run(t.name, func() {
			suite.Equal(t.expectedLog, simulation.DecodeStore(cdc, kvPairs[i], kvPairs[i]))
		})
	}
}

func TestDecoderTestSuite(t *testing.T) {
	suite.Run(t, new(DecoderTestSuite))
}
