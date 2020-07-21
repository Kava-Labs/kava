FROM golang:1.13-alpine AS build-env

# Set up dependencies
# bash, jq, curl for debugging
# git, make for installation
# libc-dev, gcc, linux-headers, eudev-dev are used for cgo and ledger installation
RUN apk add bash git make libc-dev gcc linux-headers eudev-dev jq curl


# Set working directory for the build
WORKDIR /root/kava
# default home directory is /root

COPY go.mod .
COPY go.sum .

RUN go mod download

# Add source files
COPY . .

# Install kvd, kvcli
#ENV LEDGER_ENABLED False
RUN make install

# Run kvd by default, omit entrypoint to ease using container with kvcli
CMD ["kvd"]
