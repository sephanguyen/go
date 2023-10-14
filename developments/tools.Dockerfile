ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine3.18

ENV GO111MODULE on
ENV GOPRIVATE github.com/manabie-com

# GITHUB_TOKEN to pull private repo j4. Use `gh auth token` or create a custom token.
ARG GITHUB_TOKEN

RUN apk update && apk upgrade && \
    apk add --no-cache bash git curl make unzip musl-dev gcc postgresql-client jq util-linux

RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-3.14.0-linux-x86_64.zip
RUN unzip protoc-3.14.0-linux-x86_64.zip -d protoc3
RUN mv protoc3/bin/* /usr/local/bin/
RUN mv protoc3/include/* /usr/local/include/

RUN git clone https://github.com/gogo/protobuf.git $GOPATH/src/github.com/gogo/protobuf
RUN cd $GOPATH/src/github.com/gogo/protobuf && git checkout b03c65ea87cdc3521ede29f62fe3ce239267c1bc && make install


RUN GRPC_HEALTH_PROBE_VERSION=v0.3.6 && \
    curl -L https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 -o /bin/grpc_health_probe && \
    chmod +x /bin/grpc_health_probe

WORKDIR /backend_build_cache/

# we need to build go app during this stage so majority of packages will be cached
COPY ./go.mod .
COPY ./go.sum .
RUN git config --global --add url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/manabie-com".insteadOf "https://github.com/manabie-com"
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /server ./cmd/server
