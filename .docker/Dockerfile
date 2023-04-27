FROM golang:1.16.6-alpine3.14 as builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/wallet ./cmd/wallet

ENTRYPOINT ["/app/wallet"]