CGO_CFLAGS="-I/opt/homebrew/opt/rocksdb/include" \
CGO_LDFLAGS="-L/opt/homebrew/opt/rocksdb/lib -lrocksdb -lstdc++ -lm -lz -L/opt/homebrew/opt/snappy/lib -L/opt/homebrew/opt/lz4/lib -L/opt/homebrew/opt/zstd/lib" \
  go test -v -tags=rocksdb
