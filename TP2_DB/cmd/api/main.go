package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	// Driver de PostgreSQL (el guion bajo indica que solo necesitamos que se inicialice)
	_ "github.com/lib/pq"

	// Paquete generado por sqlc y capas internas del proyecto
	db "github.com/FranRamallo/ProgramacionWebTP/db/sqlc"
	"github.com/FranRamallo/ProgramacionWebTP/internal/handlers"
	"github.com/FranRamallo/ProgramacionWebTP/internal/middleware"
	"github.com/FranRamallo/ProgramacionWebTP/internal/services"
)

func main() {
	// Configuración de la conexión a la base de datos
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "postgres"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "apirest"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "7777"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	//Establecer la conexión con PostgreSQL
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error al abrir la conexión a la BD: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("La base de datos no responde: %v", err)
	}
	log.Println("Conexión a PostgreSQL establecida correctamente.")

	// Crear la instancia del repositorio generado por sqlc
	queries := db.New(conn)

	// Inicializar la capa de Servicios (Inyección de dependencias)
	sectorSvc := services.NewSectorService(queries)
	activoSvc := services.NewActivoService(queries)
	usuarioSvc := services.NewUsuarioService(queries)
	transaccionSvc := services.NewTransaccionService(queries)
	metricaSvc := services.NewMetricaFundamentalService(queries)
	cotActualSvc := services.NewCotizacionActualService(queries)
	cotHistoricaSvc := services.NewCotizacionHistoricaService(queries)

	// Inicializar la capa de Handlers
	sectorHandler := handlers.NewSectorHandler(sectorSvc)
	activoHandler := handlers.NewActivoHandler(activoSvc)
	usuarioHandler := handlers.NewUsuarioHandler(usuarioSvc)
	transaccionHandler := handlers.NewTransaccionHandler(transaccionSvc)
	metricaHandler := handlers.NewMetricaFundamentalHandler(metricaSvc)
	cotActualHandler := handlers.NewCotizacionActualHandler(cotActualSvc)
	cotHistoricaHandler := handlers.NewCotizacionHistoricaHandler(cotHistoricaSvc)

	// Configurar el enrutador (ServeMux)
	mux := http.NewServeMux()

	// Endpoint de salud
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "GoFolio API funcionando"}`))
	})

	// Registro de rutas (con y sin barra final para soportar tanto colecciones como recursos por ID)
	mux.HandleFunc("/sectores", sectorHandler.SectoresHandler)
	mux.HandleFunc("/sectores/", sectorHandler.SectoresHandler)

	mux.HandleFunc("/activos", activoHandler.ActivosHandler)
	mux.HandleFunc("/activos/", activoHandler.ActivosHandler)

	mux.HandleFunc("/usuarios", usuarioHandler.UsuariosHandler)
	mux.HandleFunc("/usuarios/", usuarioHandler.UsuariosHandler)

	mux.HandleFunc("/transacciones", transaccionHandler.TransaccionesHandler)
	mux.HandleFunc("/transacciones/", transaccionHandler.TransaccionesHandler)

	mux.HandleFunc("/metricas-fundamentales", metricaHandler.MetricasFundamentalesHandler)
	mux.HandleFunc("/metricas-fundamentales/", metricaHandler.MetricasFundamentalesHandler)

	mux.HandleFunc("/cotizaciones-actuales", cotActualHandler.CotizacionesActualesHandler)
	mux.HandleFunc("/cotizaciones-actuales/", cotActualHandler.CotizacionesActualesHandler)

	mux.HandleFunc("/cotizaciones-historicas", cotHistoricaHandler.CotizacionesHistoricasHandler)
	mux.HandleFunc("/cotizaciones-historicas/", cotHistoricaHandler.CotizacionesHistoricasHandler)

	// Envolver todo el enrutador con el middleware de Logging y levantar el servidor
	handlerConMiddleware := middleware.Logging(mux)

	port := ":8080"
	log.Printf("Servidor corriendo en http://localhost%s", port)
	if err := http.ListenAndServe(port, handlerConMiddleware); err != nil {
		log.Fatalf("Error crítico en el servidor: %v", err)
	}
}
