package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Driver de PostgreSQL (el guion bajo indica que solo necesitamos que se inicialice)
	_ "github.com/lib/pq"

	"github.com/google/uuid"

	// Paquete generado por sqlc y capas internas del proyecto
	db "github.com/FranRamallo/ProgramacionWebTP/db/sqlc"
	"github.com/FranRamallo/ProgramacionWebTP/internal/handlers"
	"github.com/FranRamallo/ProgramacionWebTP/internal/middleware"
	"github.com/FranRamallo/ProgramacionWebTP/internal/services"
)

// usuarioDemoID es el usuario con el que opera la app mientras no exista login.
// Tiene que coincidir con el que inserta db/seed/seed.sql.
var usuarioDemoID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

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

	// Postgres puede tardar unos segundos en aceptar conexiones después de
	// levantar el contenedor, así que reintentamos antes de rendirnos.
	var errPing error
	for intento := 1; intento <= 30; intento++ {
		if errPing = conn.Ping(); errPing == nil {
			break
		}
		log.Printf("Esperando a la base de datos (%d/30): %v", intento, errPing)
		time.Sleep(time.Second)
	}
	if errPing != nil {
		log.Fatalf("La base de datos no responde: %v", errPing)
	}
	log.Println("Conexión a PostgreSQL establecida correctamente.")

	// Crear la instancia del repositorio generado por sqlc
	queries := db.New(conn)

	// Capa de Servicios
	activoSvc := services.NewActivoService(queries)
	transaccionSvc := services.NewTransaccionService(queries)

	// Capa de Presentación: handler que responde HTML renderizado con templ
	webHandler := handlers.NewWebHandler(transaccionSvc, activoSvc, usuarioDemoID)

	// Configurar el enrutador (ServeMux)
	mux := http.NewServeMux()

	// Endpoint de salud
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Archivos estáticos (CSS) bajo /static/
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Páginas (HTML renderizado en el servidor)
	mux.HandleFunc("GET /{$}", webHandler.Index)
	mux.HandleFunc("POST /transacciones", webHandler.Crear)
	mux.HandleFunc("DELETE /transacciones/{id}", webHandler.Eliminar)

	// Envolver todo el enrutador con el middleware de Logging y levantar el servidor
	handlerConMiddleware := middleware.Logging(mux)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}
	log.Printf("Servidor corriendo en http://localhost%s", port)
	if err := http.ListenAndServe(port, handlerConMiddleware); err != nil {
		log.Fatalf("Error crítico en el servidor: %v", err)
	}
}
