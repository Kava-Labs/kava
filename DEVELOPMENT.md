# Notes for developers

# Protobuf code generation

https://developers.google.com/protocol-buffers/docs/reference/go-generated
https://dev.to/techschoolguru/how-to-define-a-protobuf-message-and-generate-go-code-4g4e

```proto
option go_package = "github.com/kava-labs/kava/x/auction/legacy/v0_16";

package kava.auction.v16;
```

```sh
brew install protobuf
brew link protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```


```sh
protoc --proto_path=./proto/kava/auction/v16/ --go_out=x/auction/legacy/v16 --go_opt=paths=source_relative auction.proto genesis.proto query.proto tx.proto

```

## Linting


```sh
brew install golangci-lint
```

```
make lint
```

## Formating

```sh
go install github.com/client9/misspell/cmd/misspell@latest
go install golang.org/x/tools/cmd/goimports@latest
```
The below make

```sh
make format
```



aws s3 cp export-genesis.json s3://levi.testing.kava.io/kava-9-4-19-export-genesis.json --acl public-read
