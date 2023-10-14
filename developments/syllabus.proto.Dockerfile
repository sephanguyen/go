ARG GO_VERSION
FROM golang:${GO_VERSION} AS protoc_gen_go

RUN apt update && apt install -y --no-install-recommends curl make git unzip apt-utils
ENV GO111MODULE=on
ENV PROTOC_VERSION=3.14.0
ENV GRPC_WEB_VERSION=1.2.1
ENV BUFBUILD_VERSION=0.24.0

RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-linux-x86_64.zip
RUN unzip protoc-$PROTOC_VERSION-linux-x86_64.zip -d protoc3
RUN mv protoc3/bin/* /usr/local/bin/
RUN mv protoc3/include/* /usr/local/include/

RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.15.0
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.25.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.0.1
RUN go install github.com/envoyproxy/protoc-gen-validate@v0.9.1
RUN go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
RUN go install github.com/bold-commerce/protoc-gen-struct-transformer@v1.0.7    
RUN go install mvdan.cc/gofumpt@latest

RUN go mod download github.com/googleapis/googleapis@v0.0.0-20221209211743-f7f499371afa

ENV MOD=$GOPATH/pkg/mod
RUN mv $MOD/github.com/envoyproxy/protoc-gen-validate@v0.9.1/validate /usr/local/include/
RUN mv $MOD/github.com/googleapis/googleapis@v0.0.0-20221209211743-f7f499371afa/google/* /usr/local/include/google/

COPY ./protoc-gen-syllabus/protoc-gen-syllabus /usr/local/bin/

# protoc-gen-grpc-web
FROM protoc_gen_go AS protoc_gen_ts

RUN curl -OL https://github.com/grpc/grpc-web/releases/download/$GRPC_WEB_VERSION/protoc-gen-grpc-web-$GRPC_WEB_VERSION-linux-x86_64
RUN mv ./protoc-gen-grpc-web-$GRPC_WEB_VERSION-linux-x86_64 /usr/local/bin/protoc-gen-grpc-web
RUN chmod +x /usr/local/bin/protoc-gen-grpc-web

# protoc-gen-dart
FROM dart:2.19 AS protoc_gen_dart
ENV PROTOC_VERSION=3.14.0

RUN apt update && apt install -y --no-install-recommends curl unzip

RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-linux-x86_64.zip
RUN unzip protoc-$PROTOC_VERSION-linux-x86_64.zip -d protoc3
RUN mv protoc3/bin/* /usr/local/bin/
RUN mv protoc3/include/* /usr/local/include/

RUN git clone https://github.com/googleapis/googleapis.git /usr/local/include/googleapis
RUN mv /usr/local/include/googleapis/google/* /usr/local/include/google
RUN git clone --depth 1 --branch v0.9.1 https://github.com/envoyproxy/protoc-gen-validate.git -o validate /usr/local/include/validate

RUN dart pub global activate protoc_plugin 20.0.1
ENV PATH $PATH:/root/.pub-cache/bin

# protoc-gen-py
FROM python:3.10.5-alpine3.16 as protoc_gen_py

RUN pip install --no-cache-dir grpcio==1.46.3 grpcio-tools==1.46.3
