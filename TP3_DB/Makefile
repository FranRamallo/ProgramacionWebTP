.PHONY: run-app generate db-up db-down test

# Corre todo: genera código, compila, levanta la BD, ejecuta la app y limpia al salir.
run-app:
	@bash run-app.sh

# Regenera las vistas (templ) y el código de acceso a datos (sqlc).
generate:
	go tool templ generate
	sqlc generate

# Solo la base de datos (para desarrollo manual: después `go run ./cmd/api`).
db-up:
	docker compose up -d --wait

db-down:
	docker compose down -v --remove-orphans

# Tests: pruebas HTTP de las páginas (Hurl) + tests de Go. Ver test.sh.
test:
	@bash test.sh