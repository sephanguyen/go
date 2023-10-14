FROM google/cloud-sdk:323.0.0-alpine
RUN apk --no-cache add ca-certificates openssl

FROM google/cloud-sdk:323.0.0-alpine AS decrypt-secret
RUN apk --no-cache add ca-certificates openssl openjdk11-jre-headless
ARG SOPS_VERSION=3.7.1
ADD https://github.com/mozilla/sops/releases/download/v${SOPS_VERSION}/sops-v${SOPS_VERSION}.linux /usr/local/bin/sops

RUN chmod 755 /usr/local/bin/sops && \
    apk update && \
    apk add --no-cache jq && \
    rm -rf /var/cache/apk/*

ENTRYPOINT ["/usr/local/bin/sops"]
