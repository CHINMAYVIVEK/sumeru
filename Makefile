.PHONY: help build css run generate bp check-sql check-logs db-check i18n-export i18n-import migrate module shell test-db test-integration swc swc-check swc-test

# Extra flags for `make run`, e.g. `make run EXTRA_RUN_FLAGS='-p 9090 -d sumeru_staging'`
EXTRA_RUN_FLAGS ?=

check-sql:
	@bash scripts/check_sql_safety.sh

check-logs:
	@bash scripts/check_no_stdlog.sh

generate:
	go generate ./cmd/sumeru

# Default dev server (config: sumeru.conf, cwd: repo root).
run: generate
	go run ./cmd/sumeru -- -c sumeru.conf $(EXTRA_RUN_FLAGS)

# Production-style binary next to Makefile.
build: generate
	go build -o sumeru ./cmd/sumeru

# Scaffold a new addon (strict layout). Example: make bp NAME=parts_vendor
WITH_MODELS ?=
bp:
	@test -n "$(NAME)" || (echo 'usage: make bp NAME=my_module  (optional WITH_MODELS=1)' >&2 && exit 1)
	go run ./cmd/sumeru-bp -name $(NAME) $(if $(WITH_MODELS),-with-models,)

# Plain CSS only — edit core/engine/assets/css/ (no Sass pipeline).
css:
	@echo "No CSS build step — edit core/engine/assets/css/*.css"

SWC_DIR := core/swc

swc:
	cd $(SWC_DIR) && npm install && npm run build

swc-check:
	cd $(SWC_DIR) && npm install && npm run check

swc-test:
	cd $(SWC_DIR) && npm install && npm run test

db-check:
	go run ./cmd/sumeru-db-check -- -c sumeru.conf

i18n-export:
	go run ./cmd/sumeru-i18n-export -- -c sumeru.conf -o translations.csv

i18n-import:
	go run ./cmd/sumeru-i18n-import -- -c sumeru.conf -i translations.csv

migrate:
	go run ./cmd/sumeru-migrate -- -c sumeru.conf

module:
	go run ./cmd/sumeru-module -- -c sumeru.conf $(ARGS)

shell:
	go run ./cmd/sumeru-shell -- -c sumeru.conf

test-db:
	docker compose -f docker-compose.test.yml up -d --wait

test-integration: test-db
	SUMERU_TEST_DSN='host=localhost port=5433 user=postgres password=postgres dbname=sumeru_test sslmode=disable' \
		go test -tags=integration ./test/integration/... -count=1

help:
	@echo "Sumeru Makefile targets:"
	@echo "  make generate - go generate ./cmd/sumeru (refresh cmd/sumeru/zimports.go from sumeru.conf.example; copy to sumeru.conf for make run)"
	@echo "  make bp       - scaffold addon: make bp NAME=my_module  (optional WITH_MODELS=1)"
	@echo "  make run      - generate then go run ./cmd/sumeru -- -c sumeru.conf (optional EXTRA_RUN_FLAGS)"
	@echo "  make build    - generate then go build -o sumeru ./cmd/sumeru (binary ./sumeru)"
	@echo "  make css     - reminder: styles are plain CSS (no compile step)"
	@echo "  make swc     - build SWC bundle to core/engine/assets/swc/swc.js"
	@echo "  make swc-check - TypeScript 7 strict check (tsc --noEmit)"
	@echo "  make swc-test  - vitest unit tests for SWC"
	@echo "  make check-sql - static SQL injection pattern guard"
	@echo "  make check-logs - forbid stdlib log and operational fmt.Printf in server paths"
	@echo "  make db-check  - validate sumeru.conf and PostgreSQL connectivity"
	@echo "  make i18n-export - export sys.translation rows to translations.csv"
	@echo "  make i18n-import - import translations.csv into sys.translation"
	@echo "  make migrate     - run pending module SQL migrations"
	@echo "  make module      - module CLI (ARGS='list' | 'depends-tree' | 'install sales' | ...)"
	@echo "  make shell       - interactive ORM REPL (sumeru-shell)"
	@echo "  make test-db     - start PostgreSQL via docker-compose.test.yml"
	@echo "  make test-integration - run go test -tags=integration against test DB"
	@echo "  make help    - this message"
	@echo "See README.md for CLI flags (-d, -i, -u, -p/--http-port, --stop-after-init) and sumeru.sh."
