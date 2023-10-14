FROM ubuntu:mantic-20230624

WORKDIR /
RUN apt-get update && apt-get install -y curl jq bsdmainutils

RUN mkdir -p kafka
WORKDIR /kafka

RUN ls
COPY ./scripts/kafka/auto_restart_connector.sh ./
RUN chmod +x ./auto_restart_connector.sh

