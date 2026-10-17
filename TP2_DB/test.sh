#!/usr/bin/env bash
# GoFolio — test.sh
# Ejecución completa de tests (unitarios + API REST con Hurl) y tareas de entorno.

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuración
# ---------------------------------------------------------------------------
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

COMPOSE_FILE="docker-compose.yml"
DB_SERVICE="database"

export DB_USER="postgres"
export DB_NAME="apirest"
export DB_PASSWORD="postgres"
export DB_HOST="localhost"
export DB_PORT="7777"

MAX_WAIT_SECONDS=60
API_PID=""

# ---------------------------------------------------------------------------
# Helpers y Cleanup
# ---------------------------------------------------------------------------
log()  { echo "[test.sh] $*"; }
fail() { echo "[test.sh] ERROR: $*" >&2; exit 1; }

if docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
    DOCKER_COMPOSE=(docker-compose)
else
    fail "No se encontró Docker Compose."
fi

dc() { "${DOCKER_COMPOSE[@]}" -f "$COMPOSE_FILE" "$@"; }

cleanup() {
    log "Tareas posteriores: limpiando entorno..."
    if [ -n "$API_PID" ] && kill -0 "$API_PID" 2>/dev/null; then
        log "Deteniendo servidor API (PID: $API_PID)..."
        kill "$API_PID" 2>/dev/null || true
        wait "$API_PID" 2>/dev/null || true
    fi
    rm -rf ./bin
    dc down -v --remove-orphans >/dev/null 2>&1
    log "Limpieza completada."
}
# Garantiza que el cleanup se ejecute SIEMPRE al salir del script
trap cleanup EXIT

# ---------------------------------------------------------------------------
# 0. Verificación de herramientas
# ---------------------------------------------------------------------------
command -v go >/dev/null 2>&1 || fail "go no está instalado."
command -v docker >/dev/null 2>&1 || fail "docker no está instalado."
command -v sqlc >/dev/null 2>&1 || fail "sqlc no está instalado."
command -v hurl >/dev/null 2>&1 || fail "hurl no está instalado."
command -v curl >/dev/null 2>&1 || fail "curl no está instalado."

# ---------------------------------------------------------------------------
# 1. Tareas previas
# ---------------------------------------------------------------------------
log "1/8 - Generando código Go con sqlc..."
sqlc generate || fail "sqlc generate falló."

log "2/8 - Sincronizando dependencias..."
go mod tidy

log "3/8 - Compilando el proyecto y binario de la API..."
mkdir -p ./bin
go build ./... || fail "La compilación de paquetes falló."
go build -o ./bin/api ./cmd/api/main.go || fail "La compilación del servidor API falló."

log "4/8 - Limpiando estado previo de Docker..."
dc down -v --remove-orphans >/dev/null 2>&1

log "5/8 - Levantando contenedores (up -d)..."
dc up -d || fail "No se pudo levantar Docker."

log "6/8 - Esperando a que PostgreSQL esté 100% operativo..."
elapsed=0
until dc exec -T "$DB_SERVICE" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; do
    elapsed=$((elapsed + 1))
    if [ "$elapsed" -ge "$MAX_WAIT_SECONDS" ]; then
        fail "PostgreSQL no respondió después de ${MAX_WAIT_SECONDS}s."
    fi
    sleep 1
done
sleep 2 # Margen de seguridad para que el schema.sql termine de impactar
log "PostgreSQL listo (tardó ${elapsed}s)."

# ---------------------------------------------------------------------------
# 2. Ejecución de tests de Go (si existen) y pruebas de la API con Hurl
# ---------------------------------------------------------------------------
log "7/8 - Levantando la aplicación API en segundo plano..."
./bin/api &
API_PID=$!

elapsed=0
until curl -s http://localhost:8080/ping >/dev/null 2>&1; do
    elapsed=$((elapsed + 1))
    if [ "$elapsed" -ge "$MAX_WAIT_SECONDS" ]; then
        fail "La API no respondió en http://localhost:8080/ping después de ${MAX_WAIT_SECONDS}s."
    fi
    sleep 1
done
log "Servidor API listo y escuchando en puerto 8080 (PID: $API_PID)."

log "8/8 - Corriendo pruebas de endpoints con Hurl (requests.hurl)..."

# Apagamos el 'set -e' momentáneamente para capturar el código de salida de Hurl (y go test)
set +e
go test ./... -v
GO_TEST_EXIT_CODE=$?

hurl --test --color requests.hurl
HURL_EXIT_CODE=$?
set -e

if [ "$GO_TEST_EXIT_CODE" -eq 0 ] && [ "$HURL_EXIT_CODE" -eq 0 ]; then
    log "RESULTADO: Todos los tests (Go y Hurl API) pasaron exitosamente."
    exit 0
else
    log "RESULTADO: Se encontraron errores (Go test: ${GO_TEST_EXIT_CODE}, Hurl: ${HURL_EXIT_CODE})."
    exit 1
fi