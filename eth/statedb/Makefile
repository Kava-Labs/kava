# Set the Go compiler and linker flags
GO := go
GOFLAGS := -mod=readonly
LDFLAGS := -ldflags="-s -w"

# Define the target for running tests
test:
	$(GO) test $(GOFLAGS) ./...

.PHONY: test
