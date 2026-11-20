# ==============================================================================
# GoFolio — pruebas HTTP de las páginas (HTML renderizado en el servidor)
# Se ejecutan con `make test` (ver test.sh) contra http://localhost:8080 y la base
# recién creada: db/seed/seed.sql deja 5 activos y 5 operaciones del usuario demo
# (la más reciente es de JPM).
#
# Orden: 0-12 flujo normal (alta, validación, baja) · 13-16 escape de HTML (XSS)
#        17-24 estado vacío (se borran TODAS las operaciones, por eso va al final).
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
tx_path: xpath "string(//tbody/tr[1]//button/@hx-delete)"
[Asserts]
header "Content-Type" contains "text/html"
body contains "Registrar operación"
body contains "Historial de operaciones"
body contains "Toma de ganancias parcial"
xpath "//select[@name='activo_id']/option" count == 6
xpath "//tbody/tr[not(@class='vacio')]" count == 5
xpath "//button[@hx-delete]" count == 5
xpath "//tbody/tr[@class='vacio']" count == 1
xpath "string(//tbody/tr[1]/td[2]/strong)" == "JPM"

# 2. Archivos estáticos
GET http://localhost:8080/static/styles.css
HTTP 200
[Asserts]
header "Content-Type" contains "text/css"

# 3. Ruta inexistente
GET http://localhost:8080/no-existe
HTTP 404

# 4. Alta válida => 200 con SOLO el fragmento de la lista (sin redirección)
POST http://localhost:8080/transacciones
[FormParams]
activo_id: 2
tipo_operacion: compra
cantidad: 3
precio_unitario: 123.45
comision: 0.5
notas: Compra de prueba
HTTP 200
[Asserts]
body not contains "<html"
xpath "//article[@id='lista-transacciones']" count == 1
xpath "//tbody/tr[not(@class='vacio')]" count == 6

# 5. La operación nueva aparece en el historial
GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "//tbody/tr[not(@class='vacio')]" count == 6
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

# 8. Borrar una operación => 200 con el cuerpo vacío (HTMX saca la fila de la tabla)
DELETE http://localhost:8080{{tx_path}}
HTTP 200
[Asserts]
body == ""

# 9. Borrarla de nuevo (ya no existe) no rompe: también 200
DELETE http://localhost:8080{{tx_path}}
HTTP 200

# 10. Ya no está en el historial (6 - 1 = 5)
GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "//tbody/tr[not(@class='vacio')]" count == 5

# 11. ID inválido => 400
DELETE http://localhost:8080/transacciones/no-es-un-uuid
HTTP 400

# 12. Método no permitido sobre "/" => 405
POST http://localhost:8080/
HTTP 405

# ==============================================================================
# Escape de HTML (XSS): lo que escribe el usuario se muestra como TEXTO, nunca se
# interpreta como HTML. templ escapa `{ ... }`. Se mira la estructura (el DOM), no solo el texto: 
# si el escape falla, aparece un <script> o un atributo nuevo que antes no existía.
# ==============================================================================

# 13. Una etiqueta <script> en las notas se escapa (en el texto y en el title)
POST http://localhost:8080/transacciones
[FormParams]
activo_id: 1
tipo_operacion: compra
cantidad: 1
precio_unitario: 10
comision: 0
notas: <script>alert(1)</script>
HTTP 200
[Asserts]
body not contains "<script>alert(1)</script>"
body contains "&lt;script&gt;alert(1)&lt;/script&gt;"
xpath "//tbody//script" count == 0
xpath "//tbody/tr[not(@class='vacio')]" count == 6

# 14. Salirse del atributo title con una comilla no agrega atributos (onmouseover)
POST http://localhost:8080/transacciones
[FormParams]
activo_id: 1
tipo_operacion: compra
cantidad: 1
precio_unitario: 10
comision: 0
notas: " onmouseover="alert(1)
HTTP 200
[Asserts]
xpath "//*[@onmouseover]" count == 0
xpath "//small[@class='notas'][contains(@title, 'onmouseover')]" count == 1
xpath "//tbody/tr[not(@class='vacio')]" count == 7

# 15. En el formulario que vuelve con error (400) el valor escrito también se escapa
POST http://localhost:8080/transacciones
[FormParams]
activo_id: 1
tipo_operacion: compra
cantidad: "><script>alert(1)</script>
precio_unitario: 10
comision: 0
HTTP 400
[Asserts]
body not contains "<script>alert(1)</script>"
xpath "//form//script" count == 0
xpath "string(//input[@name='cantidad']/@value)" == "\"><script>alert(1)</script>"

# 16. Al leer de la base (GET /) las notas guardadas siguen saliendo escapadas
#     y se capturan los 7 ids para vaciar el historial más abajo
GET http://localhost:8080/
HTTP 200
[Captures]
id1: xpath "string(//tbody/tr[1]//button/@hx-delete)"
id2: xpath "string(//tbody/tr[2]//button/@hx-delete)"
id3: xpath "string(//tbody/tr[3]//button/@hx-delete)"
id4: xpath "string(//tbody/tr[4]//button/@hx-delete)"
id5: xpath "string(//tbody/tr[5]//button/@hx-delete)"
id6: xpath "string(//tbody/tr[6]//button/@hx-delete)"
id7: xpath "string(//tbody/tr[7]//button/@hx-delete)"
[Asserts]
body contains "&lt;script&gt;alert(1)&lt;/script&gt;"
xpath "//tbody//script" count == 0
xpath "//*[@onmouseover]" count == 0
xpath "//tbody/tr[not(@class='vacio')]" count == 7

# ==============================================================================
# Estado vacío: sin operaciones se ve el mensaje "Todavía no registraste ninguna
# operación." (la fila .vacio, que el CSS muestra solo cuando es la única fila) y
# el resto de la página sigue entera. Va al final porque borra todo.
# ==============================================================================

# 17-23. Borrar las 7 operaciones, una por una
DELETE http://localhost:8080{{id1}}
HTTP 200

DELETE http://localhost:8080{{id2}}
HTTP 200

DELETE http://localhost:8080{{id3}}
HTTP 200

DELETE http://localhost:8080{{id4}}
HTTP 200

DELETE http://localhost:8080{{id5}}
HTTP 200

DELETE http://localhost:8080{{id6}}
HTTP 200

DELETE http://localhost:8080{{id7}}
HTTP 200

# 24. Historial vacío: solo queda la fila del mensaje y la página no se rompe
GET http://localhost:8080/
HTTP 200
[Asserts]
xpath "//tbody/tr[not(@class='vacio')]" count == 0
xpath "//tbody/tr" count == 1
xpath "//tbody/tr[@class='vacio']" count == 1
xpath "string(//tbody/tr[@class='vacio'])" contains "Todavía no registraste ninguna operación"
xpath "//button[@hx-delete]" count == 0
xpath "//article[@id='lista-transacciones']" count == 1
xpath "//select[@name='activo_id']/option" count == 6
body contains "Registrar operación"