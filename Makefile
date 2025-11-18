.PHONY: help hosts-add hosts-check dev-setup dev-certs dev prod up down logs clean test

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

hosts-add: ## Add api.starterpack.dev to /etc/hosts (safely, no duplicates)
	@if grep -q "^127.0.0.1.*api.starterpack.dev" /etc/hosts 2>/dev/null; then \
		echo "✓ api.starterpack.dev already in /etc/hosts"; \
	else \
		echo "Adding api.starterpack.dev to /etc/hosts (requires sudo)..."; \
		echo "127.0.0.1 api.starterpack.dev" | sudo tee -a /etc/hosts > /dev/null; \
		echo "✓ Added api.starterpack.dev to /etc/hosts"; \
	fi

hosts-check: ## Check if api.starterpack.dev is in /etc/hosts
	@if grep -q "^127.0.0.1.*api.starterpack.dev" /etc/hosts 2>/dev/null; then \
		echo "✓ api.starterpack.dev is configured in /etc/hosts"; \
	else \
		echo "✗ api.starterpack.dev NOT found in /etc/hosts"; \
		echo "Run: make hosts-add"; \
	fi

dev-setup: hosts-add dev-certs ## Complete development setup (hosts + certs)
	@echo ""
	@echo "✓ Development setup complete!"
	@echo ""
	@echo "Run: make dev"

dev-certs: ## Generate development certificates
	@./scripts/generate_dev_cert.sh

dev: ## Start development environment
	@docker compose up

prod: ## Start production environment (requires DOMAIN and ACME_EMAIL env vars)
	@if [ -z "$$DOMAIN" ]; then echo "Error: DOMAIN env var not set"; exit 1; fi
	@if [ -z "$$ACME_EMAIL" ]; then echo "Error: ACME_EMAIL env var not set"; exit 1; fi
	@docker compose up -d

up: ## Start docker compose in background
	@docker compose up -d

down: ## Stop docker compose
	@docker compose down

logs: ## Show docker compose logs
	@docker compose logs -f

clean: ## Remove all containers and volumes
	@docker compose down -v
	@echo "✓ Cleaned up containers and volumes"

test: ## Test API with curl (health check)
	@echo "Testing API health endpoint..."
	@curl -s http://api.starterpack.dev/health && echo " ✓ API is healthy!" || echo "\n✗ Failed. Is docker running? (make dev)"
