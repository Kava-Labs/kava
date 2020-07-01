FROM ubuntu:19.10

ARG BNC_VERSION=0.6.3

WORKDIR /bnc

COPY --from=binance:binaries /binaries/fullnode/testnet/${BNC_VERSION}-hotfix/linux/bnbchaind ./
COPY --from=binance:binaries /binaries/cli/testnet/${BNC_VERSION}/linux/tbnbcli ./
COPY --from=binance:binaries /binaries/mytestnet/node0/gaiacli ./node0/.bnbcli
COPY --from=binance:binaries /binaries/mytestnet/node0/gaiad ./node0/.bnbchaind

EXPOSE 26657

ENTRYPOINT ["./bnbchaind", "start", "--home", "/bnc/node0/.bnbchaind"]