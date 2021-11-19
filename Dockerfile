FROM golang:1.16-alpine AS build-env

# Set up dependencies
# bash, jq, curl for debugging
# git, make for installation
# libc-dev, gcc, linux-headers, eudev-dev are used for cgo and ledger installation
RUN apk add bash git make libc-dev gcc linux-headers eudev-dev jq curl

# Set working directory for the build
WORKDIR /root/kava
# default home directory is /root

# Speed up later builds by caching the dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Add source files
COPY . .

#ENV LEDGER_ENABLED False
RUN make install

FROM alpine:3.15

RUN apk add bash jq curl
COPY --from=build-env /go/bin/kava /bin/kava

CMD ["kava"]
