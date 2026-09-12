-- GoFolio — db/queries/queries.sql
-- Consultas CRUD para sqlc. Una seccion por entidad, en el mismo
-- orden que db/schema/schema.sql.
-- Convenciones (segun catedra):
--   * columnas explicitas en SELECT/RETURNING, nunca *
--   * Create        -> :one  (INSERT ... RETURNING)
--   * Get           -> :one
--   * List          -> :many
--   * Update/Delete -> :exec (no devuelven filas)

-- =========================================================
-- sectores
-- =========================================================

-- name: CreateSector :one
INSERT INTO sector (nombre)
VALUES ($1)
RETURNING id, nombre;

-- name: GetSector :one
SELECT id, nombre
FROM sector
WHERE id = $1;

-- name: ListSectores :many
SELECT id, nombre
FROM sector
ORDER BY nombre;

-- name: UpdateSector :exec
UPDATE sector
SET nombre = $2
WHERE id = $1;

-- name: DeleteSector :exec
DELETE FROM sector
WHERE id = $1;


-- =========================================================
-- activos
-- =========================================================

-- name: CreateActivo :one
INSERT INTO activo (ticker, nombre_empresa, sector_id, bolsa, tipo_activo)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, ticker, nombre_empresa, sector_id, bolsa, tipo_activo;

-- name: GetActivo :one
SELECT id, ticker, nombre_empresa, sector_id, bolsa, tipo_activo
FROM activo
WHERE id = $1;

-- name: ListActivos :many
SELECT id, ticker, nombre_empresa, sector_id, bolsa, tipo_activo
FROM activo
ORDER BY ticker;

-- name: UpdateActivo :exec
UPDATE activo
SET ticker = $2,
    nombre_empresa = $3,
    sector_id = $4,
    bolsa = $5,
    tipo_activo = $6
WHERE id = $1;

-- name: DeleteActivo :exec
DELETE FROM activo
WHERE id = $1;


-- =========================================================
-- cotizaciones_actuales
-- PK = activo_id (relacion 1:1 con activos). En produccion el job
-- de sincronizacion usa un solo UPSERT (INSERT ... ON CONFLICT),
-- no Create + Update por separado; haciendo una unica consulta en vez de dos
-- =========================================================

-- name: UpsertCotizacionActual :one
INSERT INTO cotizacion_actual (
    activo_id, precio, cierre_anterior, variacion_dia_pct, actualizado_en
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (activo_id) 
DO UPDATE SET 
    precio = EXCLUDED.precio,
    cierre_anterior = EXCLUDED.cierre_anterior,
    variacion_dia_pct = EXCLUDED.variacion_dia_pct,
    actualizado_en = EXCLUDED.actualizado_en
RETURNING activo_id, precio, cierre_anterior, variacion_dia_pct, actualizado_en;

-- name: GetCotizacionActual :one
SELECT activo_id, precio, cierre_anterior, variacion_dia_pct, actualizado_en
FROM cotizacion_actual
WHERE activo_id = $1;

-- name: ListCotizacionesActuales :many
SELECT activo_id, precio, cierre_anterior, variacion_dia_pct, actualizado_en
FROM cotizacion_actual
ORDER BY actualizado_en DESC;

-- name: DeleteCotizacionActual :exec
DELETE FROM cotizacion_actual
WHERE activo_id = $1;


-- =========================================================
-- cotizaciones_historicas
-- =========================================================

-- name: CreateCotizacionHistorica :one
INSERT INTO cotizacion_historica (
    activo_id, fecha, precio_apertura, precio_max, precio_min, precio_cierre, volumen
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, activo_id, fecha, precio_apertura, precio_max, precio_min, precio_cierre, volumen;

-- name: GetCotizacionHistorica :one
SELECT id, activo_id, fecha, precio_apertura, precio_max, precio_min, precio_cierre, volumen
FROM cotizacion_historica
WHERE id = $1;

-- name: ListCotizacionesHistoricas :many
SELECT id, activo_id, fecha, precio_apertura, precio_max, precio_min, precio_cierre, volumen
FROM cotizacion_historica
WHERE activo_id = $1
ORDER BY fecha DESC;

-- name: UpdateCotizacionHistorica :exec
UPDATE cotizacion_historica
SET precio_apertura = $2,
    precio_max = $3,
    precio_min = $4,
    precio_cierre = $5,
    volumen = $6
WHERE id = $1;

-- name: DeleteCotizacionHistorica :exec
DELETE FROM cotizacion_historica
WHERE id = $1;


-- =========================================================
-- metricas_fundamentales
-- =========================================================

-- name: CreateMetricaFundamental :one
INSERT INTO metrica_fundamental (
    activo_id, fecha_reporte, pe_ratio, dividend_yield, deuda_capital, roe, eps
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, activo_id, fecha_reporte, pe_ratio, dividend_yield, deuda_capital, roe, eps;

-- name: GetMetricaFundamental :one
SELECT id, activo_id, fecha_reporte, pe_ratio, dividend_yield, deuda_capital, roe, eps
FROM metrica_fundamental
WHERE id = $1;

-- name: ListMetricasFundamentales :many
SELECT id, activo_id, fecha_reporte, pe_ratio, dividend_yield, deuda_capital, roe, eps
FROM metrica_fundamental
WHERE activo_id = $1
ORDER BY fecha_reporte DESC;

-- name: UpdateMetricaFundamental :exec
UPDATE metrica_fundamental
SET pe_ratio = $2,
    dividend_yield = $3,
    deuda_capital = $4,
    roe = $5,
    eps = $6
WHERE id = $1;

-- name: DeleteMetricaFundamental :exec
DELETE FROM metrica_fundamental
WHERE id = $1;


-- =========================================================
-- usuarios
-- =========================================================

-- name: CreateUsuario :one
INSERT INTO usuario (nombre_completo, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, nombre_completo, email, password_hash, fecha_registro;

-- name: GetUsuario :one
SELECT id, nombre_completo, email, password_hash, fecha_registro
FROM usuario
WHERE id = $1;

-- name: ListUsuarios :many
SELECT id, nombre_completo, email, password_hash, fecha_registro
FROM usuario
ORDER BY fecha_registro DESC;

-- name: UpdateUsuario :exec
UPDATE usuario
SET nombre_completo = $2,
    email = $3
WHERE id = $1;

-- name: DeleteUsuario :exec
DELETE FROM usuario
WHERE id = $1;

-- =========================================================
-- transacciones
-- =========================================================


-- name: CreateTransaccion :one
INSERT INTO transaccion (
    usuario_id, activo_id, tipo_operacion, cantidad,
    precio_unitario, comision, fecha_operacion, notas
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, usuario_id, activo_id, tipo_operacion, cantidad,
          precio_unitario, comision, fecha_operacion, notas;

-- name: GetTransaccion :one
SELECT id, usuario_id, activo_id, tipo_operacion, cantidad,
       precio_unitario, comision, fecha_operacion, notas
FROM transaccion
WHERE id = $1 AND usuario_id = $2;

-- name: ListTransaccionesPorUsuario :many
SELECT id, usuario_id, activo_id, tipo_operacion, cantidad,
       precio_unitario, comision, fecha_operacion, notas
FROM transaccion
WHERE usuario_id = $1
ORDER BY fecha_operacion DESC;

-- name: UpdateTransaccion :exec
UPDATE transaccion
SET tipo_operacion = $2,
    cantidad = $3,
    precio_unitario = $4,
    comision = $5,
    fecha_operacion = $6,
    notas = $7
WHERE id = $1 AND usuario_id = $8;

-- name: DeleteTransaccion :exec
DELETE FROM transaccion 
WHERE id = $1 AND usuario_id = $2;