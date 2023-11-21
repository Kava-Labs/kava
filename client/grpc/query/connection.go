package query

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// newGrpcConnection parses a GRPC endpoint and creates a connection to it
func newGrpcConnection(ctx context.Context, endpoint string) (*grpc.ClientConn, error) {
	grpcUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse grpc connection \"%s\": %v", endpoint, err)
	}

	var creds credentials.TransportCredentials
	switch grpcUrl.Scheme {
	case "http":
		creds = insecure.NewCredentials()
	case "https":
		creds = credentials.NewTLS(&tls.Config{})
	default:
		return nil, fmt.Errorf("unknown grpc url scheme: %s", grpcUrl.Scheme)
	}

	secureOpt := grpc.WithTransportCredentials(creds)
	grpcConn, err := grpc.DialContext(ctx, grpcUrl.Host, secureOpt)
	if err != nil {
		return nil, err
	}

	return grpcConn, nil
}
