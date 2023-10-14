# syntax=docker/dockerfile:1.3
# Base image used for compiling our servers' code.
FROM asia.gcr.io/student-coach-e1e95/tools:0.0.13 AS builder

ENV GOPRIVATE github.com/manabie-com
ARG GITHUB_TOKEN=$GITHUB_TOKEN

WORKDIR /backend/

COPY ./go.mod .
COPY ./go.sum .
RUN git config --global --add url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/manabie-com".insteadOf "https://github.com/manabie-com"
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
RUN --mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /server ./cmd/server

RUN chmod +x /server

#--------------------------------------------
# This image is where our servers run. For Github CI and production environments.
FROM alpine:3.18.2 AS runner

WORKDIR /

RUN apk --no-cache add ca-certificates tzdata netcat-openbsd curl

RUN wget https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64 -O cloud_sql_proxy
RUN chmod +x cloud_sql_proxy

RUN wget https://storage.googleapis.com/alloydb-auth-proxy/v0.5.0/alloydb-auth-proxy.linux.amd64 -O alloydb-auth-proxy
RUN chmod +x alloydb-auth-proxy

RUN wget https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.0.0-preview.2/cloud-sql-proxy.linux.amd64 -O cloud-sql-proxy
RUN mv cloud-sql-proxy /usr/bin/
RUN chmod +x /usr/bin/cloud-sql-proxy

COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=builder /server ./
COPY ./migrations /migrations
COPY ./scripts/wait-for.sh ./scripts/wait-for.sh
COPY ./accesscontrol /accesscontrol

ENTRYPOINT [ "/server" ]

#-------------------------------
# This image is for building j4 
FROM asia.gcr.io/student-coach-e1e95/tools:0.0.13 AS j4-builder

ENV GOPRIVATE github.com/manabie-com
ARG GITHUB_TOKEN=$GITHUB_TOKEN


WORKDIR /backend/

COPY ./go.mod .
COPY ./go.sum .
RUN git config --global --add url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/manabie-com".insteadOf "https://github.com/manabie-com"
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./j4 ./j4
COPY ./pkg ./pkg

RUN --mount=type=cache,target=/root/.cache/go-build \
 	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-extldflags=-static" -v -o /j4cli ./cmd/j4/


#--------------------------------------------
# This image is for running stress tests (j4)
FROM debian:12.0 AS j4-runner

RUN apt-get update && apt-get install -y ca-certificates netcat-openbsd wget && update-ca-certificates
RUN wget https://storage.googleapis.com/cloudsql-proxy/v1.31.1/cloud_sql_proxy.linux.amd64 -O cloud_sql_proxy
RUN chmod +x cloud_sql_proxy
WORKDIR /
COPY --from=j4-builder /j4cli ./j4

ENTRYPOINT [ "/j4" ]

#--------------------------------------------
# This image is used in initContainers for hasura-related services, so that
# we can ensure hasura is migrated correctly before the services run.
# Use `make build-hasura-migrator` to build this image.
#
# This builds into "asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v1.3.3.cli-migrations-v2-20230411".
FROM hasura/graphql-engine:v1.3.3.cli-migrations-v2 AS hasura-migrator

WORKDIR /
RUN wget https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64 -O cloud_sql_proxy && chmod +x cloud_sql_proxy
RUN mkdir -p /usr/bin
RUN wget https://storage.googleapis.com/cloudsql-proxy/v1.33.5/cloud_sql_proxy.linux.amd64 -O /usr/bin/cloud_sql_proxy && chmod +x /usr/bin/cloud_sql_proxy
RUN wget https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.1.2/cloud-sql-proxy.linux.amd64 -O /usr/bin/cloud_sql_proxy2 && chmod +x /usr/bin/cloud_sql_proxy2
ENTRYPOINT [ "/bin/docker-entrypoint.sh" ]

#--------------------------------------------
# Similar to hasura-migrator but for hasura v2.
FROM hasura/graphql-engine:v2.8.1.cli-migrations-v3 AS hasura-v2-migrator

WORKDIR /

RUN apt update && apt install -y curl
RUN curl -Lo /usr/bin/cloud_sql_proxy https://storage.googleapis.com/cloudsql-proxy/v1.29.0/cloud_sql_proxy.linux.amd64
RUN chmod +x /usr/bin/cloud_sql_proxy
ENTRYPOINT [ "/bin/docker-entrypoint.sh" ]
