FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP_VERSION=dev

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=${APP_VERSION}" \
    -o /bin/gorimpo ./cmd/gorimpo/main.go

RUN go build -o /bin/playwright-cli github.com/playwright-community/playwright-go/cmd/playwright

FROM ubuntu:jammy

RUN apt-get update && apt-get install -y ca-certificates curl unzip && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /bin/gorimpo .
COPY --from=builder /bin/playwright-cli .

RUN ./playwright-cli install --with-deps chromium

CMD ["./gorimpo"]