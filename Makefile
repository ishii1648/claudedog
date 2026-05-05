.PHONY: build install uninstall grafana-fixtures grafana-up grafana-up-e2e grafana-down grafana-screenshot lint-dashboard

PREFIX ?= $(HOME)/.local
BIN_DIR := $(PREFIX)/bin
BIN_NAME := hitl-metrics

# 実データ表示用 DB パス（grafana-up が参照）。上書き可。
HITL_METRICS_DB ?= $(HOME)/.claude/hitl-metrics.db

# grafana-up（実データ）と grafana-up-e2e（fixture）を並行起動できるよう、
# compose project とポートを分離する。互いに独立したスタックとして扱われ、
# 片方を立ち上げてももう片方のコンテナを巻き込まない。
# 既定ポートは ssh トンネル等でよく使われる 13000 / 13001 を避けて 13010+ に置く。
GRAFANA_PORT         ?= 13010
GRAFANA_E2E_PORT     ?= 13011
COMPOSE_PROJECT_REAL ?= agent-telemetry-real
COMPOSE_PROJECT_E2E  ?= agent-telemetry-e2e

build:
	CGO_ENABLED=0 go build -o bin/$(BIN_NAME) ./cmd/hitl-metrics/

install:
	@mkdir -p "$(BIN_DIR)"
	CGO_ENABLED=0 go build -o "$(BIN_DIR)/$(BIN_NAME)" ./cmd/hitl-metrics/
	@echo "Installed: $(BIN_DIR)/$(BIN_NAME)"
	@case ":$$PATH:" in *":$(BIN_DIR):"*) ;; *) echo "Warning: $(BIN_DIR) is not in PATH";; esac

uninstall:
	rm -f "$(BIN_DIR)/$(BIN_NAME)"
	@echo "Removed: $(BIN_DIR)/$(BIN_NAME)"

grafana-fixtures:
	CGO_ENABLED=0 GOTOOLCHAIN=local go test -run TestGenTestDB -v ./e2e/

grafana-up:
	@if [ ! -f "$(HITL_METRICS_DB)" ]; then \
		echo "DB not found: $(HITL_METRICS_DB)"; \
		echo "Run 'hitl-metrics sync-db' first, or override: make grafana-up HITL_METRICS_DB=/path/to/db"; \
		exit 1; \
	fi
	HITL_METRICS_DB=$(HITL_METRICS_DB) GRAFANA_PORT=$(GRAFANA_PORT) \
	    docker compose -p $(COMPOSE_PROJECT_REAL) up -d
	@echo "Waiting for Grafana to be ready..."
	@for i in $$(seq 1 60); do \
		if curl -sf http://localhost:$(GRAFANA_PORT)/api/health > /dev/null 2>&1; then \
			echo "Grafana is ready at http://localhost:$(GRAFANA_PORT)"; \
			echo "Showing data from: $(HITL_METRICS_DB)"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "Grafana failed to start within 60s"; exit 1

grafana-up-e2e: grafana-fixtures
	HITL_METRICS_DB=$(CURDIR)/e2e/testdata/hitl-metrics.db GRAFANA_PORT=$(GRAFANA_E2E_PORT) \
	    docker compose -p $(COMPOSE_PROJECT_E2E) up -d
	@echo "Waiting for Grafana to be ready..."
	@for i in $$(seq 1 60); do \
		if curl -sf http://localhost:$(GRAFANA_E2E_PORT)/api/health > /dev/null 2>&1; then \
			echo "Grafana is ready at http://localhost:$(GRAFANA_E2E_PORT) (e2e fixtures)"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "Grafana failed to start within 60s"; exit 1

grafana-down:
	-docker compose -p $(COMPOSE_PROJECT_REAL) down
	-docker compose -p $(COMPOSE_PROJECT_E2E) down

grafana-screenshot: grafana-up-e2e
	GRAFANA_PORT=$(GRAFANA_E2E_PORT) bash e2e/screenshot.sh .outputs/grafana-screenshots

lint-dashboard:
	go run github.com/grafana/dashboard-linter@latest lint --strict --config grafana/dashboards/.lint grafana/dashboards/hitl-metrics.json
