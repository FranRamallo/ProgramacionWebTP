# ProgramacionWebTP
# TP1 - GoFolio

Dominio elegido: GoFolio, una aplicación web para registrar transacciones y gestionar un portafolio de inversiones.

## Entidades Principales del Dominio

Para representar el sistema, guardaremos la información estructurada en las siguientes entidades:

* Usuarios: Datos de la cuenta del inversor (ID, email, credenciales).
* Activos: Catálogo de instrumentos financieros operables (Ticker, nombre, tipo).
* Sectores: Clasificación de la industria para analizar la diversificación del portafolio (ej: Tecnología, Finanzas).
* Transacciones: Registro inmutable de compras y ventas (usuario, activo, cantidad, precio, fecha, tipo).
* Cotizaciones Actuales: Último precio de mercado conocido por cada activo.
* Cotizaciones Históricas: Serie de precios diarios (apertura, cierre, volumen) para graficar rendimientos.
* Métricas Fundamentales: Datos financieros trimestrales de las empresas (P/E Ratio, ROE, Deuda/Capital) para análisis.
* Alertas: Reglas configuradas por el usuario para vigilar precios o métricas (ej: "Avisar si AAPL cae de $150").
* Notificaciones: Registro de los avisos al usuario.


# Cómo ejecutar el proyecto

1. Asegúrate de tener Go instalado.
2. Cambiar a la rama tp1 (git switch tp1)
3. Abre una terminal en esta carpeta y ejecuta el siguiente comando:
   go run main.go
