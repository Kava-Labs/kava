package grpc

import (
	"errors"

	"github.com/kava-labs/kava/client/grpc/query"
	"github.com/kava-labs/kava/client/grpc/util"
)

// KavaGrpcClient enables the usage of kava grpc query clients and query utils
type KavaGrpcClient struct {
	config KavaGrpcClientConfig

	// Query clients for cosmos and kava modules
	Query *query.QueryClient

	// Utils for common queries (ie fetch an unpacked BaseAccount)
	*util.Util
}

// KavaGrpcClientConfig is a configuration struct for a KavaGrpcClient
type KavaGrpcClientConfig struct {
	// note: add future config options here
}

// NewClient creates a new KavaGrpcClient via a grpc url
func NewClient(grpcUrl string) (*KavaGrpcClient, error) {
	return NewClientWithConfig(grpcUrl, NewDefaultConfig())
}

// NewClientWithConfig creates a new KavaGrpcClient via a grpc url and config
func NewClientWithConfig(grpcUrl string, config KavaGrpcClientConfig) (*KavaGrpcClient, error) {
	if grpcUrl == "" {
		return nil, errors.New("grpc url cannot be empty")
	}
	query, error := query.NewQueryClient(grpcUrl)
	if error != nil {
		return nil, error
	}
	client := &KavaGrpcClient{
		Query:  query,
		Util:   util.NewUtil(query),
		config: config,
	}
	return client, nil
}

func NewDefaultConfig() KavaGrpcClientConfig {
	return KavaGrpcClientConfig{}
}
