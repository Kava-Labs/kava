package util

import (
	"context"
	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"strconv"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
	query "github.com/kava-labs/kava/client/grpc/query"
)

// Util contains utility functions for the Kava gRPC client
type Util struct {
	query          *query.QueryClient
	encodingConfig params.EncodingConfig
}

// NewUtil creates a new Util instance
func NewUtil(query *query.QueryClient) *Util {
	encodingConfig := app.MakeEncodingConfig()
	_ = app.NewApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		app.DefaultNodeHome,
		nil,
		encodingConfig,
		app.DefaultOptions,
	)

	return &Util{
		query:          query,
		encodingConfig: encodingConfig,
	}
}

func (u *Util) CtxAtHeight(height int64) context.Context {
	heightStr := strconv.FormatInt(height, 10)
	return metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, heightStr)
}
