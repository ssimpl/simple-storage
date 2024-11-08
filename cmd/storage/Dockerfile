FROM golang:1.23-alpine3.20 AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go install ./cmd/storage/...

FROM alpine:3.20 AS main

WORKDIR /usr/local/bin
COPY --from=builder /go/bin/storage .

CMD ["storage"]
