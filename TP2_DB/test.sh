#!/usr/bin/env bash
# GoFolio — test.sh
# Corre todo lo necesario para probar el proyecto de punta a punta:
#   1) tareas previas  (sqlc, compilación, contenedores limpios, esperar la base, cargar schema)
#   2) tests
#   3) tareas posteriores (bajar contenedores y borrar volúmenes)
#
# Uso:
#   ./test.sh
#   make test          (llama a este mismo script)

set -uo pipefail

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

COMPOSE_FILE="docker-compose.yml"
DB_SERVICE="database"
DB_USER="postgres"
DB_NAME="apirest"
SCHEMA_FILE="db/schema/schema.sql"

MAX_WAIT_SECONDS=60

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

log()  { echo "[test.sh] $*"; }
fail() { echo "[test.sh] ERROR: $*" >&2; exit 1; }

# Detecta si el binario disponible es "docker compose" (v2) o "docker-compose" (v1)
if docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
    DOCKER_COMPOSE=(docker-compose)
else
    fail "No se encontró 'docker compose' ni 'docker-compose'. Instalá Docker Compose para continuar."
fi

dc() { "${DOCKER_COMPOSE[@]}" -f "$COMPOSE_FILE" "$@"; }

cleanup() {
    log "Tareas posteriores: bajando contenedores y borrando volúmenes..."
    dc down -v
}
# Se ejecuta siempre al salir del script, pase lo que pase (tests OK, tests
# fallidos, o un error en cualquier paso previo).
trap cleanup EXIT

# ---------------------------------------------------------------------------
# 0. Verificación de herramientas
# ---------------------------------------------------------------------------

command -v go   >/dev/null 2>&1 || fail "go no está instalado o no está en el PATH."
command -v docker >/dev/null 2>&1 || fail "docker no está instalado o no está en el PATH."
command -v sqlc >/dev/null 2>&1 || fail "sqlc no está instalado. Instalalo con: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"

# ---------------------------------------------------------------------------
# 1. Tareas previas
# ---------------------------------------------------------------------------

log "1/6 - Generando código Go con sqlc (sqlc generate)..."
sqlc generate || fail "sqlc generate falló."

log "2/6 - Compilando el proyecto (go build)..."
go build ./... || fail "go build falló."

log "3/6 - Bajando contenedores y volúmenes previos (por si quedó algo de una corrida anterior)..."
dc down -v

log "4/6 - Levantando contenedores..."
dc up -d || fail "docker compose up falló."

log "5/6 - Esperando a que Postgres esté listo..."
elapsed=0
until dc exec -T "$DB_SERVICE" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; do
    elapsed=$((elapsed + 1))
    if [ "$elapsed" -ge "$MAX_WAIT_SECONDS" ]; then
        fail "Postgres no respondió después de ${MAX_WAIT_SECONDS}s."
    fi
    sleep 1
done
log "    Postgres listo (tardó ${elapsed}s)."

# ---------------------------------------------------------------------------
# 2. Ejecución de tests
# ---------------------------------------------------------------------------

log "Corriendo tests (go test ./...)..."
go test ./... -v
TEST_EXIT_CODE=$?

# ---------------------------------------------------------------------------
# 3. Tareas posteriores
# ---------------------------------------------------------------------------
# El trap EXIT (arriba) ya baja los contenedores y borra los volúmenes acá,
# se hayan pasado los tests o no.

if [ "$TEST_EXIT_CODE" -eq 0 ]; then
    log "Tests OK."
else
    log "Tests FALLARON (código ${TEST_EXIT_CODE})."
fi

exit "$TEST_EXIT_CODE"