package util

import (
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
