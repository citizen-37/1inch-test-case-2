ARG GO_VERSION=1.20

FROM golang:${GO_VERSION} AS build

RUN apt update && apt install protobuf-compiler -y

RUN wget https://github.com/ethereum/solidity/releases/download/v0.8.21/solc-static-linux -O /bin/solc && \
    chmod +x /bin/solc

RUN git clone https://github.com/ethereum/go-ethereum.git &&  \
    cd go-ethereum &&  \
    make devtools

WORKDIR /app