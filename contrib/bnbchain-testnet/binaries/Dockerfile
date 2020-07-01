FROM ubuntu:19.10

ARG BNC_VERSION=0.6.3

RUN apt-get update && \
    apt-get install -y git git-lfs

WORKDIR /binaries

RUN git clone --depth 1 https://github.com/binance-chain/node-binary.git .

RUN git lfs pull -I fullnode/testnet/${BNC_VERSION}-hotfix/linux
RUN git lfs pull -I cli/testnet/${BNC_VERSION}/linux

RUN ./fullnode/testnet/${BNC_VERSION}-hotfix/linux/bnbchaind testnet --acc-prefix tbnb --chain-id Binance-Chain-Nile --v 1

RUN sed -i "s/BEP12Height = 9223372036854775807/BEP12Height = 1/" ./mytestnet/node0/gaiad/config/app.toml