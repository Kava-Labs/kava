package util

import (
	"context"
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
	return &Util{
		query:          query,
		encodingConfig: app.MakeEncodingConfig(),
	}
}

func (u *Util) CtxAtHeight(height int64) context.Context {
	heightStr := strconv.FormatInt(height, 10)
	return metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, heightStr)
}
