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
<img src="https://github.com/FranRamallo/ProgramacionWebTP/raw/fe6fd8f6cc68ecd0b959f3e2453ed1ccf2f845ef/MVP_Gofolio.png" width="900" alt="MVP Gofolio">

# Cómo ejecutar el proyecto

git clone https://github.com/FranRamallo/ProgramacionWebTP.git
cd ProgramacionWebTP/servidor-go
git switch tp1
go run main.go
