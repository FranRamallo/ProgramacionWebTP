# ==============================================================================
# 0. HEALTHCHECK
# ==============================================================================
GET http://localhost:8080/ping
HTTP 200
[Asserts]
jsonpath "$.status" == "GoFolio API funcionando"


# ==============================================================================
# 1. SECTORES (Crear, Listar, Obtener, Actualizar, Validaciones)
# ==============================================================================

# 1.1 Listar inicialmente vacío (debe devolver array [])
GET http://localhost:8080/sectores
HTTP 200
[Asserts]
jsonpath "$" count == 0

# 1.2 Crear Sector (201 Created)
POST http://localhost:8080/sectores
Content-Type: application/json
{
    "nombre": "Tecnología"
}
HTTP 201
[Captures]
sector_id: jsonpath "$.id"
[Asserts]
jsonpath "$.nombre" == "Tecnología"

# 1.3 Validación: Crear Sector duplicado (400 Bad Request)
POST http://localhost:8080/sectores
Content-Type: application/json
{
    "nombre": "Tecnología"
}
HTTP 400

# 1.4 Listar Sectores (200 OK)
GET http://localhost:8080/sectores
HTTP 200
[Asserts]
jsonpath "$" count == 1

# 1.5 Obtener Sector por ID (200 OK)
GET http://localhost:8080/sectores/{{sector_id}}
HTTP 200
[Asserts]
jsonpath "$.id" == {{sector_id}}
jsonpath "$.nombre" == "Tecnología"

# 1.6 Actualizar Sector (200 OK)
PUT http://localhost:8080/sectores/{{sector_id}}
Content-Type: application/json
{
    "nombre": "Tecnología y Software"
}
HTTP 200
[Asserts]
jsonpath "$.nombre" == "Tecnología y Software"

# 1.7 Obtener Sector inexistente (404 Not Found)
GET http://localhost:8080/sectores/99999
HTTP 404


# ==============================================================================
# 2. ACTIVOS (Crear, Listar, Obtener, Actualizar)
# ==============================================================================

# 2.1 Crear Activo vinculado al Sector creado (201 Created)
POST http://localhost:8080/activos
Content-Type: application/json
{
    "ticker": "aapl",
    "nombre_empresa": "Apple Inc.",
    "sector_id": {{sector_id}},
    "bolsa": "NASDAQ",
    "tipo_activo": "Acción"
}
HTTP 201
[Captures]
activo_id: jsonpath "$.id"
[Asserts]
jsonpath "$.ticker" == "AAPL"
jsonpath "$.nombre_empresa" == "Apple Inc."

# 2.2 Listar Activos (200 OK)
GET http://localhost:8080/activos
HTTP 200
[Asserts]
jsonpath "$" count == 1

# 2.3 Obtener Activo por ID (200 OK)
GET http://localhost:8080/activos/{{activo_id}}
HTTP 200
[Asserts]
jsonpath "$.id" == {{activo_id}}
jsonpath "$.ticker" == "AAPL"

# 2.4 Actualizar Activo (200 OK)
PUT http://localhost:8080/activos/{{activo_id}}
Content-Type: application/json
{
    "ticker": "AAPL",
    "nombre_empresa": "Apple Corporation",
    "sector_id": {{sector_id}},
    "bolsa": "NASDAQ",
    "tipo_activo": "Acción"
}
HTTP 200
[Asserts]
jsonpath "$.nombre_empresa" == "Apple Corporation"


# ==============================================================================
# 3. COTIZACIÓN ACTUAL (Crear/Upsert, Listar, Obtener, Actualizar, Eliminar)
# ==============================================================================

# 3.1 Crear Cotización Actual (201 Created)
POST http://localhost:8080/cotizaciones-actuales
Content-Type: application/json
{
    "activo_id": {{activo_id}},
    "precio": "185.5000",
    "cierre_anterior": "180.0000",
    "variacion_dia_pct": "3.0556"
}
HTTP 201
[Asserts]
jsonpath "$.activo_id" == {{activo_id}}
jsonpath "$.precio" == "185.5000"

# 3.2 Listar Cotizaciones Actuales (200 OK)
GET http://localhost:8080/cotizaciones-actuales
HTTP 200
[Asserts]
jsonpath "$" count == 1

# 3.3 Obtener Cotización Actual por Activo ID (200 OK)
GET http://localhost:8080/cotizaciones-actuales/{{activo_id}}
HTTP 200
[Asserts]
jsonpath "$.precio" == "185.5000"

# 3.4 Actualizar Cotización Actual (200 OK)
PUT http://localhost:8080/cotizaciones-actuales/{{activo_id}}
Content-Type: application/json
{
    "precio": "190.0000",
    "cierre_anterior": "180.0000",
    "variacion_dia_pct": "5.5556"
}
HTTP 200
[Asserts]
jsonpath "$.precio" == "190.0000"

# 3.5 Eliminar Cotización Actual (204 No Content)
DELETE http://localhost:8080/cotizaciones-actuales/{{activo_id}}
HTTP 204

# 3.6 Verificar que fue eliminada (404 Not Found)
GET http://localhost:8080/cotizaciones-actuales/{{activo_id}}
HTTP 404


# ==============================================================================
# 4. COTIZACIÓN HISTÓRICA (Crear, Listar, Obtener, Actualizar, Eliminar)
# ==============================================================================

# 4.1 Crear Cotización Histórica (201 Created)
POST http://localhost:8080/cotizaciones-historicas
Content-Type: application/json
{
    "activo_id": {{activo_id}},
    "fecha": "2026-03-30",
    "precio_apertura": "180.0000",
    "precio_max": "186.0000",
    "precio_min": "179.5000",
    "precio_cierre": "185.5000",
    "volumen": 54000000
}
HTTP 201
[Captures]
cot_hist_id: jsonpath "$.id"
[Asserts]
jsonpath "$.activo_id" == {{activo_id}}
jsonpath "$.volumen" == 54000000

# 4.2 Validación de coherencia OHLC: mínimo mayor al máximo (400 Bad Request)
POST http://localhost:8080/cotizaciones-historicas
Content-Type: application/json
{
    "activo_id": {{activo_id}},
    "fecha": "2026-03-31",
    "precio_apertura": "180.0000",
    "precio_max": "170.0000",
    "precio_min": "190.0000",
    "precio_cierre": "185.0000",
    "volumen": 1000
}
HTTP 400

# 4.3 Listar Cotizaciones Históricas por Activo (200 OK)
GET http://localhost:8080/cotizaciones-historicas?activo_id={{activo_id}}
HTTP 200
[Asserts]
jsonpath "$" count == 1

# 4.4 Obtener Cotización Histórica por ID (200 OK)
GET http://localhost:8080/cotizaciones-historicas/{{cot_hist_id}}
HTTP 200
[Asserts]
jsonpath "$.id" == {{cot_hist_id}}

# 4.5 Actualizar Cotización Histórica (200 OK)
PUT http://localhost:8080/cotizaciones-historicas/{{cot_hist_id}}
Content-Type: application/json
{
    "precio_apertura": "180.0000",
    "precio_max": "188.0000",
    "precio_min": "179.5000",
    "precio_cierre": "187.0000",
    "volumen": 60000000
}
HTTP 200
[Asserts]
jsonpath "$.precio_cierre" == "187.0000"
jsonpath "$.volumen" == 60000000

# 4.6 Eliminar Cotización Histórica (204 No Content)
DELETE http://localhost:8080/cotizaciones-historicas/{{cot_hist_id}}
HTTP 204

# 4.7 Verificar eliminación (404 Not Found)
GET http://localhost:8080/cotizaciones-historicas/{{cot_hist_id}}
HTTP 404


# ==============================================================================
# 5. MÉTRICA FUNDAMENTAL (Crear, Listar, Obtener, Actualizar, Eliminar)
# ==============================================================================

# 5.1 Crear Métrica Fundamental (201 Created)
POST http://localhost:8080/metricas-fundamentales
Content-Type: application/json
{
    "activo_id": {{activo_id}},
    "fecha_reporte": "2026-03-31",
    "pe_ratio": "28.5000",
    "dividend_yield": "0.550",
    "deuda_capital": "1.4500",
    "roe": "145.200",
    "eps": "6.4200"
}
HTTP 201
[Captures]
metrica_id: jsonpath "$.id"
[Asserts]
jsonpath "$.activo_id" == {{activo_id}}

# 5.2 Listar Métricas Fundamentales por Activo (200 OK)
GET http://localhost:8080/metricas-fundamentales?activo_id={{activo_id}}
HTTP 200
[Asserts]
jsonpath "$" count == 1

# 5.3 Obtener Métrica Fundamental por ID (200 OK)
GET http://localhost:8080/metricas-fundamentales/{{metrica_id}}
HTTP 200
[Asserts]
jsonpath "$.id" == {{metrica_id}}

# 5.4 Actualizar Métrica Fundamental (200 OK)
PUT http://localhost:8080/metricas-fundamentales/{{metrica_id}}
Content-Type: application/json
{
    "pe_ratio": "30.0000",
    "dividend_yield": "0.600",
    "deuda_capital": "1.4000",
    "roe": "150.000",
    "eps": "6.8000"
}
HTTP 200
[Asserts]
jsonpath "$.pe_ratio.String" == "30.0000"

# 5.5 Eliminar Métrica Fundamental (204 No Content)
DELETE http://localhost:8080/metricas-fundamentales/{{metrica_id}}
HTTP 204

# 5.6 Verificar eliminación (404 Not Found)
GET http://localhost:8080/metricas-fundamentales/{{metrica_id}}
HTTP 404


# ==============================================================================
# 6. USUARIOS (Crear, Listar, Obtener, Actualizar)
# ==============================================================================

# 6.1 Crear Usuario (201 Created)
POST http://localhost:8080/usuarios
Content-Type: application/json
{
    "nombre_completo": "Juan Perez",
    "email": "juan.perez@example.com",
    "password": "miPasswordSegura123"
}
HTTP 201
[Captures]
usuario_id: jsonpath "$.id"
[Asserts]
jsonpath "$.nombre_completo" == "Juan Perez"
jsonpath "$.email" == "juan.perez@example.com"

# 6.2 Listar Usuarios (200 OK)
GET http://localhost:8080/usuarios
HTTP 200
[Asserts]
jsonpath "$" count == 1

# 6.3 Obtener Usuario por UUID (200 OK)
GET http://localhost:8080/usuarios/{{usuario_id}}
HTTP 200
[Asserts]
jsonpath "$.id" == "{{usuario_id}}"

# 6.4 Actualizar Usuario (200 OK)
PUT http://localhost:8080/usuarios/{{usuario_id}}
Content-Type: application/json
{
    "nombre_completo": "Juan Ignacio Perez",
    "email": "juan.perez@example.com"
}
HTTP 200
[Asserts]
jsonpath "$.nombre_completo" == "Juan Ignacio Perez"


# ==============================================================================
# 7. TRANSACCIONES (Crear, Listar, Obtener, Actualizar, Eliminar)
# ==============================================================================

# 7.1 Crear Transacción (201 Created)
POST http://localhost:8080/transacciones
Content-Type: application/json
{
    "usuario_id": "{{usuario_id}}",
    "activo_id": {{activo_id}},
    "tipo_operacion": "compra",
    "cantidad": "10.500000",
    "precio_unitario": "185.5000",
    "comision": "1.50",
    "notas": "Compra inicial para largo plazo"
}
HTTP 201
[Captures]
transaccion_id: jsonpath "$.id"
[Asserts]
jsonpath "$.tipo_operacion" == "compra"
jsonpath "$.cantidad" == "10.500000"

# 7.2 Listar Transacciones del Usuario (200 OK)
GET http://localhost:8080/transacciones?usuario_id={{usuario_id}}
HTTP 200
[Asserts]
jsonpath "$" count == 1

# 7.3 Obtener Transacción por ID y Usuario (200 OK)
GET http://localhost:8080/transacciones/{{transaccion_id}}?usuario_id={{usuario_id}}
HTTP 200
[Asserts]
jsonpath "$.id" == "{{transaccion_id}}"

# 7.4 Actualizar Transacción (200 OK)
PUT http://localhost:8080/transacciones/{{transaccion_id}}
Content-Type: application/json
{
    "usuario_id": "{{usuario_id}}",
    "tipo_operacion": "compra",
    "cantidad": "15.000000",
    "precio_unitario": "185.5000",
    "comision": "2.00",
    "notas": "Compra ajustada"
}
HTTP 200
[Asserts]
jsonpath "$.cantidad" == "15.000000"

# 7.5 Eliminar Transacción (204 No Content)
DELETE http://localhost:8080/transacciones/{{transaccion_id}}?usuario_id={{usuario_id}}
HTTP 204

# 7.6 Verificar eliminación de Transacción (404 Not Found)
GET http://localhost:8080/transacciones/{{transaccion_id}}?usuario_id={{usuario_id}}
HTTP 404


# ==============================================================================
# 8. ELIMINACIÓN DE ENTIDADES PADRE (Usuario, Activo y Sector)
# ==============================================================================

# 8.1 Eliminar Usuario (204 No Content) y verificar 404
DELETE http://localhost:8080/usuarios/{{usuario_id}}
HTTP 204

GET http://localhost:8080/usuarios/{{usuario_id}}
HTTP 404

# 8.2 Eliminar Activo (204 No Content) y verificar 404
DELETE http://localhost:8080/activos/{{activo_id}}
HTTP 204

GET http://localhost:8080/activos/{{activo_id}}
HTTP 404

# 8.3 Eliminar Sector (204 No Content) y verificar 404
DELETE http://localhost:8080/sectores/{{sector_id}}
HTTP 204

GET http://localhost:8080/sectores/{{sector_id}}
HTTP 404