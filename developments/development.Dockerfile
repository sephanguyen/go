ARG GO_VERSION
#--------------------------------------------
# Prepare curl to download other stuffs
FROM debian:12.0 AS curl-ready
RUN apt-get update && apt-get install -y curl

#--------------------------------------------
# Download grpc_health_probe
FROM curl-ready AS download-grpc-health-probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.6 && \
    curl -L -o /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

#--------------------------------------------
# Download modd
FROM curl-ready AS download-modd
RUN MODD_VERSION=0.8 && \
    curl -LO https://github.com/cortesi/modd/releases/download/v${MODD_VERSION}/modd-${MODD_VERSION}-linux64.tgz && \
    tar -xvf modd-${MODD_VERSION}-linux64.tgz && \
    mv modd-${MODD_VERSION}-linux64/modd /usr/local/bin/modd && \
    rm modd-${MODD_VERSION}-linux64.tgz

#--------------------------------------------
# A base no-op image without /server binaries
# TODO @anhpngt: instead of building -> deploying, speed up the
# process by building + deploying a no-op image in parallel, then
# finally live reload the server.
FROM debian:12.0 AS no-op

WORKDIR /

RUN apt-get update -y && apt-get install -y ca-certificates net-tools netcat-openbsd
COPY --from=download-grpc-health-probe /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=download-modd /usr/local/bin/modd /usr/local/bin/modd
ENTRYPOINT [ "modd" ]

#--------------------------------------------
# This image is where both of our servers and integration tests run.
# For local/CI development environments.
FROM debian:12.0 AS developer

WORKDIR /

RUN apt-get update -y && apt-get install -y ca-certificates net-tools curl netcat-openbsd bash
COPY --from=download-grpc-health-probe /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=download-modd /usr/local/bin/modd /usr/local/bin/modd

COPY ./build/server /server
COPY ./migrations /migrations
COPY ./features /backend/features
COPY ./build/bdd.test /backend/features/bdd.test
COPY ./build/stub /stub
COPY ./accesscontrol /accesscontrol

COPY ./scripts/wait-for.sh ./scripts/wait-for.sh
ENTRYPOINT [ "/server" ]

#--------------------------------------------
FROM debian:12.0 AS j4-runner

RUN apt-get update && apt-get install -y ca-certificates netcat-openbsd wget && update-ca-certificates
WORKDIR /
COPY ./build/j4 /j4 
COPY ./build/rqlite /rqlite 

ENTRYPOINT [ "/j4" ]

#--------------------------------------------
# Download delve for debugging
FROM golang:${GO_VERSION} AS download-delve
# Build Delve
RUN CGO_ENABLED=0 go install github.com/go-delve/delve/cmd/dlv@v1.20.2

#--------------------------------------------
# Debug session
# This require build command enable -gcflags='all=-N -l'

FROM developer AS developer-debug
COPY --from=download-delve /go/bin/dlv /

ENTRYPOINT ["/dlv","--listen=:40000", "--continue", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/server", "--"]
