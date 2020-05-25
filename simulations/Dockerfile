FROM golang:1.13-alpine AS build-env

# Set up dependencies
# bash for debugging
# git, make for installation
# libc-dev, gcc, linux-headers, eudev-dev are used for cgo and ledger installation (possibly)
RUN apk add bash git make libc-dev gcc linux-headers eudev-dev jq

# Install aws cli
RUN apk add python py-pip
RUN pip install awscli

# Set working directory for the build
WORKDIR /root/kava
# default home directory is /root

COPY go.mod .
COPY go.sum .

RUN go mod download

# Add source files
COPY .git .git
COPY app app
COPY cli_test cli_test
COPY cmd cmd
COPY app app
COPY x x
COPY Makefile .

# Install kvd, kvcli
ENV LEDGER_ENABLED False
RUN make install

# Copy in simulation script after to decrease image build time
COPY simulations simulations

# Run kvd by default, omit entrypoint to ease using container with kvcli
CMD ["kvd"]
