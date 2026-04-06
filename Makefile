.PHONY: build run docker-build docker-run clean dev-up dev-down prod-up

# Variáveis
IMAGE_NAME = gorimpo-test
VERSION = dev

build:
	@echo "🔨 Compilando o GOrimpo..."
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/gorimpo ./cmd/gorimpo/main.go

run: build
	@echo "🚀 Rodando o GOrimpo localmente..."
	./bin/gorimpo

run-headless:
	go run -ldflags "-X main.Version=dev-headless" ./cmd/gorimpo/main.go
	

docker-build:
	@echo "🐳 Construindo imagem Docker (isso pode demorar por causa do Chromium)..."
	docker build --build-arg APP_VERSION=dev-docker -t $(IMAGE_NAME) .
	
docker-run:
	@echo "🚀 Rodando container de teste..."
	docker run --rm \
		--env-file .env \
		-v ./data:/app/data \
		-v ./config.yaml:/app/config.yaml \
		$(IMAGE_NAME)

dev-up:
	@echo "🧪 Subindo GOrimpo + Infra local para testes..."
	docker compose -f docker-compose.dev.yml up -d --build

dev-down:
	@echo "🛑 Derrubando ambiente de testes..."
	docker compose -f docker-compose.dev.yml down

prod-up:
	@echo "🔥 Subindo Produção (Pull do GHCR + Watchtower)..."
	docker compose -f docker-compose.yml up -d

clean:
	@echo "🧹 Limpando binários e containers..."
	rm -rf bin/
	docker rmi $(IMAGE_NAME) || true