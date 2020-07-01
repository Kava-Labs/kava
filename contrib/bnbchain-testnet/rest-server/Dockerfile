FROM ubuntu:19.10

ARG BNC_VERSION=0.6.3

WORKDIR /api-server

COPY --from=binance:binaries /binaries/cli/testnet/${BNC_VERSION}/linux/tbnbcli ./

RUN echo password | ./tbnbcli keys add key

EXPOSE 8080

ENTRYPOINT ["./tbnbcli", "api-server", "--chain-id", "Binance-Dev", "--laddr", "tcp://0.0.0.0:8080", "--node", "http://192.168.10.8:26657"]