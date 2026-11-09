# ==============================================================================
# GoFolio — pruebas HTTP de las páginas (HTML renderizado en el servidor)
# Se ejecutan con `make test` (ver test.sh) contra http://localhost:8080 y la base
# recién creada: db/seed/seed.sql deja 5 activos y 5 operaciones del usuario demo
# (la más reciente es de JPM).
# ==============================================================================

# 0. Healthcheck
GET http://localhost:8080/ping
HTTP 200
[Asserts]
body contains "ok"

# 1. Página principal: formulario + historial con los datos del seed
GET http://localhost:8080/
HTTP 200
[Captures]
tx_id: xpath "string(//tbody/tr[1]//input[@name='id']/@value)"
[Asserts]
header "Content-Type" contains "text/html"
body contains "Registrar operación"
body contains "Historial de operaciones"
body contains "Toma de ganancias parcial"
xpath "//select[@name='activo_id']/option" count == 6
xpath "//tbody/tr" count == 5
xpath "string(//tbody/tr[1]/td[2]/strong)" == "JPM"

# 2. Archivos estáticos
GET http://localhost:8080/static/styles.css
HTTP 200
[Asserts]
header "Content-Type" contains "text/css"

# 3. Ruta inexistente
GET http://localhost:8080/no-existe
HTTP 404

# 4. Alta válida => Post/Redirect/Get: 303 hacia "/"
POST http://localhost:8080/transacciones
[FormParams]
activo_id: 2
tipo_operacion: compra
cantidad: 3
precio_unitario: 123.45
comision: 0.5
notas: Compra de prueba
HTTP 303
[Asserts]
header "Location" == "/"

# 5. La operación nueva aparece en el historial
GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "//tbody/tr" count == 6
body contains "$123.45"
body contains "Compra de prueba"

# 6. Validación: cantidad negativa => 400, con mensaje y conservando lo escrito
POST http://localhost:8080/transacciones
[FormParams]
activo_id: 1
tipo_operacion: compra
cantidad: -5
precio_unitario: 10
comision: 0
HTTP 400
[Asserts]
body contains "la cantidad debe ser un número válido mayor a 0"
xpath "string(//input[@name='cantidad']/@value)" == "-5"

# 7. Validación: activo inválido => 400
POST http://localhost:8080/transacciones
[FormParams]
activo_id: abc
tipo_operacion: compra
cantidad: 1
precio_unitario: 10
comision: 0
HTTP 400
[Asserts]
body contains "Elegí un activo de la lista"

# 8. Borrar una operación => 303 hacia "/"
POST http://localhost:8080/transacciones/eliminar
[FormParams]
id: {{tx_id}}
HTTP 303
[Asserts]
header "Location" == "/"

# 9. Borrarla de nuevo (ya no existe) no rompe: vuelve a "/"
POST http://localhost:8080/transacciones/eliminar
[FormParams]
id: {{tx_id}}
HTTP 303

# 10. Ya no está en el historial (6 - 1 = 5)
GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "//tbody/tr" count == 5

# 11. ID inválido => 400
POST http://localhost:8080/transacciones/eliminar
[FormParams]
id: no-es-un-uuid
HTTP 400

# 12. Método no permitido sobre "/" => 405
POST http://localhost:8080/
HTTP 405