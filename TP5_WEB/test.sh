#!/usr/bin/env bash
# GoFolio — test.sh (también: `make test`)
# Igual que run-app.sh, pero en vez de dejar la app corriendo ejecuta los tests:
# pruebas HTTP de las páginas (Hurl) y tests de Go. Al terminar limpia todo.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

command -v hurl >/dev/null 2>&1 || { echo "[test] Falta instalar Hurl: https://hurl.dev" >&2; exit 1; }

# Tareas posteriores: corren siempre (también si un paso falla o se corta con Ctrl+C).
cleanup() {
    echo
    echo "[test] Limpiando: deteniendo la app y borrando contenedores y volúmenes..."
    if [ -n "${APP_PID:-}" ]; then kill "$APP_PID" 2>/dev/null || true; fi
    rm -rf ./bin
    docker compose down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT
trap 'exit 130' INT TERM

echo "[test] 1/5 - Generando vistas (templ) y código de acceso a datos (sqlc)..."
go tool templ generate
if command -v sqlc >/dev/null 2>&1; then
    sqlc generate
else
    go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
fi

echo "[test] 2/5 - Compilando..."
go build -o ./bin/gofolio ./cmd/api

echo "[test] 3/5 - Recreando la base de datos (--wait espera a que esté lista)..."
docker compose down -v --remove-orphans
docker compose up -d --wait

echo "[test] 4/5 - Iniciando la aplicación y esperando a que responda..."
./bin/gofolio &
APP_PID=$!
curl -fs --retry 30 --retry-connrefused --retry-delay 1 -o /dev/null http://localhost:8080/ping

echo "[test] 5/5 - Ejecutando tests..."
status=0
hurl --test web.hurl || status=1
go test -count=1 ./... || status=1
exit "$status"