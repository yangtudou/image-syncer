FROM golang:1.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -ldflags="-s -w" -o image-syncer .

FROM alpine:latest

WORKDIR /bin

COPY --from=builder /src/image-syncer ./image-syncer

RUN chmod +x ./image-syncer \
    && apk add --no-cache ca-certificates \
    && update-ca-certificates

ENTRYPOINT ["./image-syncer"]