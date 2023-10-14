ARG GO_VERSION
FROM golang:${GO_VERSION} AS protoc_gen_go_and_web

RUN apt update && apt install -y --no-install-recommends curl make git unzip apt-utils

RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-3.14.0-linux-x86_64.zip
RUN unzip protoc-3.14.0-linux-x86_64.zip -d protoc3
RUN mv protoc3/bin/* /usr/local/bin/
RUN mv protoc3/include/* /usr/local/include/

RUN git clone https://github.com/gogo/protobuf.git $GOPATH/src/github.com/gogo/protobuf
RUN cd $GOPATH/src/github.com/gogo/protobuf && git checkout b03c65ea87cdc3521ede29f62fe3ce239267c1bc && make install

#protoc-gen-grpc-web
RUN curl -OL https://github.com/grpc/grpc-web/releases/download/1.2.0/protoc-gen-grpc-web-1.2.0-linux-x86_64
RUN mv ./protoc-gen-grpc-web-1.2.0-linux-x86_64 /usr/local/bin/protoc-gen-grpc-web
RUN chmod +x /usr/local/bin/protoc-gen-grpc-web

# check proto syntax
RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v0.7.1/buf-Linux-x86_64" -o "/usr/local/bin/buf"
RUN chmod +x "/usr/local/bin/buf"
RUN apt install -y clang-format

#protoc-gen-dart
FROM dart:2.19 AS protoc_gen_dart

RUN apt update && apt install -y --no-install-recommends curl unzip

RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-3.14.0-linux-x86_64.zip
RUN unzip protoc-3.14.0-linux-x86_64.zip -d protoc3
RUN mv protoc3/bin/* /usr/local/bin/
RUN mv protoc3/include/* /usr/local/include/

RUN git clone https://github.com/googleapis/googleapis.git /usr/local/include/googleapis
RUN git clone https://github.com/gogo/protobuf.git usr/local/include/github.com/gogo/protobuf

RUN dart pub global activate protoc_plugin 20.0.1
ENV PATH $PATH:/root/.pub-cache/bin
