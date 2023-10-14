FROM golang:1.20.5-alpine3.18

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build -o /docker-gs-ping

EXPOSE 1234

CMD [ "/docker-gs-ping" ]
