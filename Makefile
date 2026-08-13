.PHONY: help build css run generate bp check-sql check-logs

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

help:
	@echo "Sumeru Makefile targets:"
	@echo "  make generate - go generate ./cmd/sumeru (refresh cmd/sumeru/zimports.go from sumeru.conf.example; copy to sumeru.conf for make run)"
	@echo "  make bp       - scaffold addon: make bp NAME=my_module  (optional WITH_MODELS=1)"
	@echo "  make run      - generate then go run ./cmd/sumeru -- -c sumeru.conf (optional EXTRA_RUN_FLAGS)"
	@echo "  make build    - generate then go build -o sumeru ./cmd/sumeru (binary ./sumeru)"
	@echo "  make css     - reminder: styles are plain CSS (no compile step)"
	@echo "  make check-sql - static SQL injection pattern guard"
	@echo "  make check-logs - forbid stdlib log and operational fmt.Printf in server paths"
	@echo "  make help    - this message"
	@echo "See README.md for CLI flags (-d, -i, -u, -p/--http-port, --stop-after-init) and sumeru.sh."
