# Start with go container
FROM golang:alpine AS builder
WORKDIR /go/src/github.com/kava-labs/kava

# Install go package manager
#RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh - doesn't work as alpine has no curl
RUN apk add --no-cache git && go get -u github.com/golang/dep/cmd/dep

# Install go packages (without updating Gopkg, as there is no source code to update from)(also with -v for verbose)
ADD Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only -v

# Copy in app code and build
COPY . .
RUN go build ./cmd/kvd && go build ./cmd/kvcli

# Copy app binary over to small container.
# Using alpine instad of scratch to aid in debugging and avoid complicated compile
# note the home directory for alpine is /root/
FROM alpine
COPY --from=builder /go/src/github.com/kava-labs/kava/kvd /go/src/github.com/kava-labs/kava/kvcli /usr/bin/
CMD ["kvd", "start"]
