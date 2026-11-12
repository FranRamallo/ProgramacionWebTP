# ProgramacionWebTP — GoFolio

**Dominio elegido:** GoFolio, una aplicación web para registrar transacciones y gestionar un portafolio de inversiones financieras.

**Integrantes:** Franco Ramallo, Jeronimo Pourtau Fleischmacher, Juan Bautista Galvan

> **Documentación completa** (arquitectura, rutas, flujos, decisiones de diseño, modelo de datos y pruebas): [docs/DOCUMENTACION.md](docs/DOCUMENTACION.md)

| TP | Qué se construyó |
|---|---|
| TP2 | Modelo de datos en PostgreSQL y acceso a datos con `sqlc` |
| TP3 | API REST (JSON) con `net/http`, capas handlers → services → sqlc (rama `tp3`) |
| TP4 | Frontend con HTML + JavaScript que consumía la API |
| TP5 | Interfaz renderizada en el servidor con [templ](https://templ.guide), sin JavaScript (rama `tp5`) |
| **TP6** | **Interfaz reactiva con [HTMX](https://htmx.org): registrar y borrar operaciones sin recargar la página** (rama `tp6`) |

## Qué hace la aplicación (TP6)

Una página única donde el usuario administra las operaciones (compras y ventas) de su portafolio:

* **Historial de operaciones:** fecha, activo, tipo, cantidad, precio, comisión y notas.
* **Registrar una operación:** al enviar el formulario la lista se actualiza sola, **sin recargar la página**. Si los datos no son válidos, aparece el mensaje de error debajo del formulario.
* **Borrar una operación:** el botón pide confirmación y la fila desaparece sin recargar.

El servidor genera el HTML con **templ** y los estilos usan **Pico.css**. La interactividad la aporta **HTMX** (librería cargada por CDN): el navegador pide HTML en segundo plano y reemplaza solo la parte que cambió. No hay archivos JavaScript: la interactividad se declara con atributos `hx-*` directamente en el HTML y **el repositorio no contiene archivos `.js`**.

## Cómo ejecutar

**Requisitos:** Go 1.21 o superior (el `go.mod` pide la 1.26.5: si tu versión es menor, Go descarga solo ese toolchain la primera vez), Docker con Docker Compose v2 (`docker compose`), `make` e internet (imágenes de Docker, módulos de Go y, por CDN, Pico.css y HTMX). No hace falta instalar `sqlc` ni `templ`.

```bash
git clone -b tp6 --single-branch https://github.com/FranRamallo/ProgramacionWebTP.git
cd ProgramacionWebTP
make run-app            # equivalente: bash TP5_WEB/run-app.sh
```

Cuando aparezca `Servidor corriendo en http://localhost:8080`, abrí esa dirección en el navegador.
Para detener la aplicación: `Ctrl+C` (limpia todo solo). La primera ejecución tarda más porque descarga imágenes y módulos.

El proyecto vive en `TP5_WEB/`; el `Makefile` de la raíz solo reenvía los comandos a esa carpeta (también se puede trabajar desde adentro: `cd TP5_WEB && make run-app`).

### Qué hace `run-app.sh`

| Fase | Pasos |
|---|---|
| **1. Tareas previas** | `templ generate` (vistas) → `sqlc generate` (acceso a datos; si `sqlc` no está instalado se ejecuta con `go run`) → compilar a `bin/gofolio` → borrar contenedores y volúmenes viejos → `docker compose up -d --wait` (levanta PostgreSQL y espera a que esté listo) |
| **2. Ejecución** | Inicia la aplicación en primer plano |
| **3. Tareas posteriores** | Siempre al salir (`Ctrl+C`, error o `kill`): detiene la app, borra `bin/` y ejecuta `docker compose down -v` (contenedores y volúmenes) |

La base se crea desde cero en cada ejecución con `db/schema/schema.sql` y los datos de ejemplo de
`db/seed/seed.sql` (5 activos y 5 operaciones de un usuario demo), así que el historial nunca arranca vacío.

El código generado (`db/sqlc/*.go` y `views/*_templ.go`) **no se sube al repositorio**: lo produce
`run-app.sh` en cada ejecución. En un clon recién hecho, el editor marca errores en esas carpetas hasta
correr `make run-app` o `make generate`.

### Configuración (opcional)

Sin configurar nada funciona. Para cambiar un valor, exportalo antes de ejecutar
(por ejemplo `DB_PORT=5433 make run-app`):

| Variable | Por defecto | Uso |
|---|---|---|
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `postgres` / `postgres` / `apirest` | Credenciales de PostgreSQL |
| `DB_HOST` / `DB_PORT` | `localhost` / `7777` | Dónde escucha PostgreSQL (puerto publicado por Docker) |
| `PORT` | `8080` | Puerto HTTP de la aplicación |

## Pruebas

```bash
make test
```

Genera, compila, levanta una base nueva y ejecuta las pruebas HTTP de las páginas (`web.hurl`, 13 peticiones con [Hurl](https://hurl.dev)) y los tests de Go de la capa de datos (`go test`). Al terminar limpia todo. Hurl solo hace falta para los tests: `brew install hurl` (macOS), `pacman -S hurl` (Arch) o el `.deb` desde [Hurl Releases](https://github.com/Orange-OpenSource/hurl/releases) (Ubuntu/Debian). Detalle en [docs/DOCUMENTACION.md](docs/DOCUMENTACION.md#ejecución-y-pruebas).

## Estructura del repositorio

```
.
├── Makefile                # make run-app / test / generate (reenvía a TP5_WEB/)
├── README.md               # cómo ejecutar
├── docs/
│   └── DOCUMENTACION.md    # documentación completa
└── TP5_WEB/
    ├── cmd/api/main.go     # arma dependencias y rutas, levanta el servidor
    ├── views/              # vistas templ (layout, formulario, lista)
    ├── internal/           # handlers, services (lógica de negocio) y middleware
    ├── db/                 # schema, seed, queries (sqlc) y tests de la capa de datos
    ├── static/styles.css   # estilos propios
    ├── web.hurl            # pruebas HTTP
    └── run-app.sh · test.sh · Makefile · docker-compose.yml · sqlc.yaml
```

## Problemas frecuentes

| Síntoma | Solución |
|---|---|
| `port is already allocated` / `address already in use` | Otro programa usa el puerto 7777 u 8080: cerralo o cambiá `DB_PORT` / `PORT` (ver Configuración) |
| `Cannot connect to the Docker daemon` | Iniciá Docker (Docker Desktop o `systemctl start docker`) |
| `docker: 'compose' is not a docker command` | Falta Docker Compose v2: https://docs.docker.com/compose/install/ |
| `Permission denied` al ejecutar `./run-app.sh` | Usá `make run-app` o `bash TP5_WEB/run-app.sh` |
| `Falta instalar Hurl` al correr `make test` | Instalá Hurl (ver [Pruebas](#pruebas)); solo hace falta para los tests, no para ejecutar la app |
| Los botones no hacen nada o la lista no se actualiza sin recargar | HTMX no se cargó (sin internet o bloqueado): en `F12` → Network buscá `htmx.min.js` |
| Cambié `styles.css` y el navegador sigue mostrando los estilos viejos | Recargá sin caché con `Ctrl+Shift+R` |