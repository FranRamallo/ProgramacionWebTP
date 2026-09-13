#!/usr/bin/env bash
# GoFolio — test.sh
# Ejecución completa de tests y tareas de entorno.

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

MAX_WAIT_SECONDS=60

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
    log "Tareas posteriores: limpiando entorno (down -v)..."
    dc down -v >/dev/null 2>&1
}
# Garantiza que el cleanup se ejecute SIEMPRE al salir del script
trap cleanup EXIT

# ---------------------------------------------------------------------------
# 0. Verificación de herramientas
# ---------------------------------------------------------------------------
command -v go >/dev/null 2>&1 || fail "go no está instalado."
command -v docker >/dev/null 2>&1 || fail "docker no está instalado."
command -v sqlc >/dev/null 2>&1 || fail "sqlc no está instalado."

# ---------------------------------------------------------------------------
# 1. Tareas previas
# ---------------------------------------------------------------------------
log "1/7 - Generando código Go con sqlc..."
sqlc generate || fail "sqlc generate falló."

log "2/7 - Sincronizando dependencias..."
go mod tidy

log "3/7 - Compilando el proyecto..."
go build ./... || fail "La compilación falló."

log "4/7 - Limpiando estado previo de Docker..."
dc down -v >/dev/null 2>&1

log "5/7 - Levantando contenedores (up -d)..."
dc up -d || fail "No se pudo levantar Docker."

log "6/7 - Esperando a que PostgreSQL esté 100% operativo..."
elapsed=0
until dc exec -T "$DB_SERVICE" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; do
    elapsed=$((elapsed + 1))
    if [ "$elapsed" -ge "$MAX_WAIT_SECONDS" ]; then
        fail "PostgreSQL no respondió después de ${MAX_WAIT_SECONDS}s."
    fi
    sleep 1
done
sleep 2 # Margen de seguridad para que el schema.sql termine de impactar
log "    PostgreSQL listo (tardó ${elapsed}s)."

# ---------------------------------------------------------------------------
# 2. Ejecución de tests
# ---------------------------------------------------------------------------
log "7/7 - Corriendo la suite de tests..."

# Apagamos el 'set -e' momentáneamente para que un test fallido no aborte
# el script antes de que podamos procesar el resultado de salida.
set +e
go test ./... -v
TEST_EXIT_CODE=$?
set -e

# ---------------------------------------------------------------------------
# 3. Tareas posteriores (Manejadas por el TRAP)
# ---------------------------------------------------------------------------
if [ "$TEST_EXIT_CODE" -eq 0 ]; then
    log "RESULTADO: Todos los tests pasaron exitosamente."
else
    log "RESULTADO: Se encontraron errores en los tests (Código ${TEST_EXIT_CODE})."
fi

exit "$TEST_EXIT_CODE"