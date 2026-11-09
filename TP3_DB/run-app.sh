#!/usr/bin/env bash
# GoFolio — run-app.sh (también: make run-app)
#   1) Tareas previas: generar código (templ + sqlc), compilar, recrear la BD y esperar a que esté lista.
#   2) Ejecutar la aplicación.
#   3) Tareas posteriores: detener la app y borrar contenedores y volúmenes (corre siempre, gracias a trap).

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

# 3) Tareas posteriores: se registran primero, así corren pase lo que pase (Ctrl+C, error o kill).
cleanup() {
    echo
    echo "[run-app] Limpiando: deteniendo la app y borrando contenedores y volúmenes..."
    if [ -n "${APP_PID:-}" ]; then kill "$APP_PID" 2>/dev/null || true; fi
    rm -rf ./bin
    docker compose down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT
trap 'exit 0' INT TERM

# 1) Tareas previas
echo "[run-app] 1/4 - Generando vistas (templ) y código de acceso a datos (sqlc)..."
go tool templ generate
if command -v sqlc >/dev/null 2>&1; then
    sqlc generate
else
    go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
fi

echo "[run-app] 2/4 - Compilando..."
go build -o ./bin/gofolio ./cmd/api

echo "[run-app] 3/4 - Recreando la base de datos (--wait espera a que esté lista)..."
docker compose down -v --remove-orphans
docker compose up -d --wait

# 2) Ejecutar la aplicación
echo "[run-app] 4/4 - Iniciando GoFolio en http://localhost:8080 (Ctrl+C para detener y limpiar)"
./bin/gofolio &
APP_PID=$!
wait "$APP_PID"