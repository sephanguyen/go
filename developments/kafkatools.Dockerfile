FROM alpine:3.16.0 AS runner

RUN apk add --no-cache bash curl postgresql-client jq util-linux
ENV CLOUDSQLPROXY_VERSION=v1.21.0
RUN wget "https://storage.googleapis.com/cloudsql-proxy/$CLOUDSQLPROXY_VERSION/cloud_sql_proxy.linux.amd64" -O /cloud_sql_proxy
RUN chmod +x /cloud_sql_proxy
