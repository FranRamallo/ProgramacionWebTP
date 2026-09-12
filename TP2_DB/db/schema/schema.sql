-- Extension necesaria para gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE sector (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE activo (
    id BIGSERIAL PRIMARY KEY,
    ticker VARCHAR(10) NOT NULL UNIQUE,
    nombre_empresa VARCHAR(255) NOT NULL,
    sector_id INT REFERENCES sector(id) ON DELETE SET NULL,
    bolsa VARCHAR(20) NOT NULL,
    tipo_activo VARCHAR(50) NOT NULL
);

CREATE INDEX idx_activos_sector_id ON activo(sector_id);

CREATE TABLE cotizacion_actual (
    activo_id BIGINT PRIMARY KEY REFERENCES activo(id) ON DELETE CASCADE,
    precio DECIMAL(15, 4) NOT NULL,
    cierre_anterior DECIMAL(15, 4) NOT NULL,
    variacion_dia_pct DECIMAL(8, 4) NOT NULL,
    actualizado_en TIMESTAMPTZ NOT NULL
);

CREATE TABLE cotizacion_historica (
    id BIGSERIAL PRIMARY KEY,
    activo_id BIGINT NOT NULL REFERENCES activo(id) ON DELETE CASCADE,
    fecha DATE NOT NULL,
    precio_apertura DECIMAL(15, 4) NOT NULL,
    precio_max DECIMAL(15, 4) NOT NULL,
    precio_min DECIMAL(15, 4) NOT NULL,
    precio_cierre DECIMAL(15, 4) NOT NULL,
    volumen BIGINT NOT NULL,
    UNIQUE (activo_id, fecha)
);

CREATE INDEX idx_cotizacion_historica_activo_id ON cotizacion_historica(activo_id);

CREATE TABLE metrica_fundamental (
    id BIGSERIAL PRIMARY KEY,
    activo_id BIGINT NOT NULL REFERENCES activo(id) ON DELETE CASCADE,
    fecha_reporte DATE NOT NULL,
    pe_ratio DECIMAL(10, 4),
    dividend_yield DECIMAL(6, 3),
    deuda_capital DECIMAL(10, 4),
    roe DECIMAL(6, 3),
    eps DECIMAL(10, 4),
    UNIQUE (activo_id, fecha_reporte)
);

CREATE INDEX idx_metrica_fundamental_activo_id ON metrica_fundamental(activo_id);

CREATE TABLE usuario (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre_completo VARCHAR(150) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    fecha_registro TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transaccion (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    activo_id BIGINT NOT NULL REFERENCES activo(id) ON DELETE RESTRICT,
    tipo_operacion VARCHAR(10) NOT NULL CHECK (tipo_operacion IN ('compra', 'venta')),
    cantidad DECIMAL(18, 6) NOT NULL,
    precio_unitario DECIMAL(15, 4) NOT NULL,
    comision DECIMAL(15, 2) NOT NULL,
    fecha_operacion TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    notas TEXT
);

CREATE INDEX idx_transaccion_usuario_id ON transaccion(usuario_id);
CREATE INDEX idx_transaccion_activo_id ON transaccion(activo_id);