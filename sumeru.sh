#!/usr/bin/env bash
set -euo pipefail
#
# Sumeru — run from the repository root so sumeru.conf paths resolve
# (addons_path, core/engine/assets, logo_path, etc.).
#
# Examples (see README.md for flag mapping and full use cases):
#
#   ./sumeru.sh
#       → go run ./cmd/sumeru -- -c "${SUMERU_CONF:-sumeru.conf}"
#
#   SUMERU_CONF=my.conf ./sumeru.sh
#       → same as ./sumeru.sh -c my.conf
#
#   ./sumeru.sh -c /abs/path/sumeru.conf
#       → explicit config file
#
#   ./sumeru.sh -p 9090
#   ./sumeru.sh --http-port 9090
#       → override listen port (--http-port / -p)
#
#   ./sumeru.sh -d sumeru_staging
#   ./sumeru.sh --database sumeru_staging
#       → override PostgreSQL dbname (-d / --database)
#
#   ./sumeru.sh -p 9090 -d sumeru_dev -i sales,crm
#       → custom port + database + install multiple modules, then start server
#
#   ./sumeru.sh -i sales
#   ./sumeru.sh -i sales,crm
#       → install one or many modules (-i), then start server
#
#   ./sumeru.sh -u sales
#   ./sumeru.sh -u sales,sale_demo_inherit --stop-after-init
#   ./sumeru.sh -u all --stop-after-init
#       → update one or many modules (-u); -u all = every installed module; --stop-after-init exits (no HTTP)
#
#   ./sumeru.sh -i company,user --stop-after-init -p 9090 -d mydb
#       → install then exit; port/db flags still apply for the short-lived process
#
# Startup wiring lives in sumeru/core/server (server.Run). This script only runs
#   go run ./cmd/sumeru -- "$@"
#
if [ "$#" -eq 0 ]; then
  set -- -c "${SUMERU_CONF:-sumeru.conf}"
fi
go generate ./cmd/sumeru
exec go run ./cmd/sumeru -- "$@"
