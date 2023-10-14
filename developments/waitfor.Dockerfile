#--------------------------------------------
# Simple image containings the wait-for.sh script
FROM alpine:3.18.2 AS wait-for
COPY ./scripts/wait-for.sh ./scripts/wait-for.sh
