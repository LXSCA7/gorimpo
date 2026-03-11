FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -o /bin/playwright-cli github.com/playwright-community/playwright-go/cmd/playwright

FROM ubuntu:jammy AS setup

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/playwright-cli /app/playwright-cli
RUN ./playwright-cli install --with-deps chromium firefox webkit \
    && rm -f /app/playwright-cli

FROM golang:1.25 AS gorimpobuilder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP_VERSION=dev

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=${APP_VERSION}" \
    -o /bin/gorimpo ./cmd/gorimpo/main.go

FROM setup AS final

WORKDIR /app

COPY --from=gorimpobuilder /bin/gorimpo /app/gorimpo

CMD ["./gorimpo"]