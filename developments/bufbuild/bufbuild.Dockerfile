FROM node:16-alpine3.15

RUN apk update && apk add --no-cache git && apk add bash

ARG GITHUB_TOKEN
ENV GITHUB_TOKEN $GITHUB_TOKEN

RUN git config --global --add url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/manabie-com".insteadOf "https://github.com/manabie-com"

WORKDIR /bufbuild

COPY ./project/.npmrc /bufbuild/
COPY ./project/package.json /bufbuild/
COPY ./project/yarn.lock /bufbuild/

RUN yarn install

COPY project /bufbuild
