FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod go.sum ./
COPY src/ ./src
COPY config/ ./config
COPY templates/ ./templates

RUN go build -o service ./src

EXPOSE 8000

CMD [ "./service" ]
