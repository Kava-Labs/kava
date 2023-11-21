package grpc

import (
	"errors"

	"github.com/kava-labs/kava/client/grpc/query"
	"github.com/kava-labs/kava/client/grpc/util"
)

// KavaGrpcClient enables the usage of kava grpc query clients and query utils
type KavaGrpcClient struct {
	// Query clients for cosmos and kava modules
	Query *query.QueryClient
	// Utils for common queries (ie fetch an unpacked BaseAccount)
	Util *util.Util
}

// NewClient creates a new KavaGrpcClient via a grpc url
func NewClient(grpcUrl string) (*KavaGrpcClient, error) {
	if grpcUrl == "" {
		return nil, errors.New("grpc url cannot be empty")
	}
	query, error := query.NewQueryClient(grpcUrl)
	if error != nil {
		return nil, error
	}
	client := &KavaGrpcClient{
		Query: query,
		Util:  util.NewUtil(query),
	}
	return client, nil
}
