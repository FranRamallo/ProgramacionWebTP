-- GoFolio — datos de ejemplo.
-- Se ejecuta automáticamente despues de db/schema/schema.sql, solo cuando el
-- volumen de Postgres está vacío (docker-compose.yml: 01-schema.sql y 02-seed.sql).

-- Sectores
INSERT INTO sector (nombre) VALUES
    ('Tecnología'),
    ('Finanzas'),
    ('Consumo masivo'),
    ('Energía');

-- Activos
INSERT INTO activo (ticker, nombre_empresa, sector_id, bolsa, tipo_activo) VALUES
    ('AAPL', 'Apple Inc.',              (SELECT id FROM sector WHERE nombre = 'Tecnología'),    'NASDAQ', 'ACCIÓN'),
    ('MSFT', 'Microsoft Corporation',   (SELECT id FROM sector WHERE nombre = 'Tecnología'),    'NASDAQ', 'ACCIÓN'),
    ('JPM',  'JPMorgan Chase & Co.',    (SELECT id FROM sector WHERE nombre = 'Finanzas'),      'NYSE',   'ACCIÓN'),
    ('KO',   'The Coca-Cola Company',   (SELECT id FROM sector WHERE nombre = 'Consumo masivo'), 'NYSE',   'ACCIÓN'),
    ('XOM',  'Exxon Mobil Corporation', (SELECT id FROM sector WHERE nombre = 'Energía'),       'NYSE',   'ACCIÓN');

-- Cotización actual (precio, cierre anterior y variación % del día)
INSERT INTO cotizacion_actual (activo_id, precio, cierre_anterior, variacion_dia_pct, actualizado_en)
SELECT a.id, v.precio, v.cierre, ROUND((v.precio - v.cierre) / v.cierre * 100, 4), now()
FROM (VALUES
    ('AAPL', 190.50, 189.20),
    ('MSFT', 415.30, 417.10),
    ('JPM',  198.75, 197.90),
    ('KO',    62.40,  62.55),
    ('XOM',  114.20, 112.80)
) AS v(ticker, precio, cierre)
JOIN activo a ON a.ticker = v.ticker;

-- Usuario demo. La app todavía no tiene login, así que opera siempre con este
-- usuario (el mismo UUID está en cmd/api/main.go). La contraseña del demo es
-- "demo1234" y se guarda hasheada con bcrypt.
INSERT INTO usuario (id, nombre_completo, email, password_hash) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Usuario Demo', 'demo@gofolio.local', crypt('demo1234', gen_salt('bf', 10)));

-- Operaciones de ejemplo del usuario demo
INSERT INTO transaccion (usuario_id, activo_id, tipo_operacion, cantidad, precio_unitario, comision, fecha_operacion, notas)
SELECT '11111111-1111-1111-1111-111111111111'::uuid, a.id, v.tipo, v.cantidad, v.precio, v.comision, v.fecha::timestamptz, v.notas
FROM (VALUES
    ('AAPL', 'compra', 10,  172.50, 1.00, '2026-03-10 10:30:00+00', 'Primera compra'),
    ('MSFT', 'compra',  5,  405.00, 1.00, '2026-04-02 14:15:00+00', NULL),
    ('KO',   'compra', 20,   60.25, 0.50, '2026-05-20 09:45:00+00', 'Posición defensiva'),
    ('AAPL', 'venta',   4,  188.00, 1.00, '2026-08-14 16:00:00+00', 'Toma de ganancias parcial'),
    ('JPM',  'compra',  8,  195.40, 1.00, '2026-10-01 13:20:00+00', NULL)
) AS v(ticker, tipo, cantidad, precio, comision, fecha, notas)
JOIN activo a ON a.ticker = v.ticker;