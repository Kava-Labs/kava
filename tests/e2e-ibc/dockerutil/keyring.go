package dockerutil

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/docker/docker/client"
)

// NewLocalKeyringFromDockerContainer copies the contents of the given container directory into a specified local directory.
// This allows test hosts to sign transactions on behalf of test users.
func NewLocalKeyringFromDockerContainer(ctx context.Context, dc *client.Client, localDirectory, containerKeyringDir, containerId string) (keyring.Keyring, error) {
	reader, _, err := dc.CopyFromContainer(ctx, containerId, containerKeyringDir)
	if err != nil {
		return nil, err
	}

	if err := os.Mkdir(filepath.Join(localDirectory, "keyring-test"), os.ModePerm); err != nil {
		return nil, err
	}
	tr := tar.NewReader(reader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}

		var fileBuff bytes.Buffer
		if _, err := io.Copy(&fileBuff, tr); err != nil {
			return nil, err
		}

		name := hdr.Name
		extractedFileName := path.Base(name)
		isDirectory := extractedFileName == ""
		if isDirectory {
			continue
		}

		filePath := filepath.Join(localDirectory, "keyring-test", extractedFileName)
		if err := os.WriteFile(filePath, fileBuff.Bytes(), os.ModePerm); err != nil {
			return nil, err
		}
	}

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	return keyring.New("", keyring.BackendTest, localDirectory, os.Stdin, cdc)
}
