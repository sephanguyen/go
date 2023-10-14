FROM node:16-alpine3.15

COPY project /graphq-mesh

WORKDIR /graphq-mesh

RUN yarn install
