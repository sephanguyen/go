# syntax=docker/dockerfile:1.3
FROM ubuntu:22.04 AS builder

ARG INSTALL_DIR=/usr/local/bin

RUN apt update && apt install -y curl
RUN curl -o ${INSTALL_DIR}/cloud_sql_proxy_v2 -L https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.4.0/cloud-sql-proxy.linux.amd64
RUN chmod +x ${INSTALL_DIR}/cloud_sql_proxy_v2
RUN curl -o ${INSTALL_DIR}/cloud_sql_proxy -L https://storage.googleapis.com/cloudsql-proxy/v1.33.8/cloud_sql_proxy.linux.amd64
RUN chmod +x ${INSTALL_DIR}/cloud_sql_proxy
RUN curl -o ${INSTALL_DIR}/jq -L https://github.com/jqlang/jq/releases/download/jq-1.6/jq-linux64
RUN chmod +x ${INSTALL_DIR}/jq
RUN curl -o ${INSTALL_DIR}/yq -L https://github.com/mikefarah/yq/releases/download/v4.34.1/yq_linux_amd64
RUN chmod +x ${INSTALL_DIR}/yq

#### Running container
FROM ubuntu:22.04 AS runner

ARG INSTALL_DIR=/usr/local/bin

COPY --from=builder ${INSTALL_DIR}/cloud_sql_proxy_v2 ${INSTALL_DIR}/cloud_sql_proxy_v2
COPY --from=builder ${INSTALL_DIR}/cloud_sql_proxy ${INSTALL_DIR}/cloud_sql_proxy
COPY --from=builder ${INSTALL_DIR}/jq ${INSTALL_DIR}/jq
COPY --from=builder ${INSTALL_DIR}/yq ${INSTALL_DIR}/yq

RUN apt update \
    && apt install -y postgresql-client \
    && apt install -y netcat psmisc \
    && apt install -y ca-certificates && update-ca-certificates \
    && apt clean
