package util

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

// NewGrpcConnection parses a GRPC endpoint and creates a connection to it
func NewGrpcConnection(endpoint string) (*grpc.ClientConn, error) {
	grpcUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	var secureOpt grpc.DialOption
	switch grpcUrl.Scheme {
	case "http":
		secureOpt = grpc.WithInsecure()
	case "https":
		creds := credentials.NewTLS(&tls.Config{})
		secureOpt = grpc.WithTransportCredentials(creds)
	default:
		return nil, fmt.Errorf("unknown grpc url scheme: %s", grpcUrl.Scheme)
	}

	grpcConn, err := grpc.Dial(grpcUrl.Host, secureOpt)
	if err != nil {
		return nil, err
	}

	return grpcConn, nil
}

func CtxAtHeight(height int64) context.Context {
	heightStr := strconv.FormatInt(height, 10)
	return metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, heightStr)
}
