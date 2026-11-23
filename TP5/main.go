package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	db "tp5/db/sqlc"
	"tp5/views"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 1. Abrir la conexión con SQLite.
	conn, err := sql.Open("sqlite3", "products.db")
	if err != nil {
		log.Fatal("Error al abrir la base de datos:", err)
	}
	defer conn.Close()

	// Comprobar que podemos conectarnos.
	if err := conn.Ping(); err != nil {
		log.Fatal("Error al conectar con SQLite:", err)
	}

	// 2. Leer y ejecutar el esquema SQL.
	schema, err := os.ReadFile("db/schema/schema.sql")
	if err != nil {
		log.Fatal("Error al leer schema.sql:", err)
	}

	if _, err := conn.Exec(string(schema)); err != nil {
		log.Fatal("Error al ejecutar schema.sql:", err)
	}

	// 3. Crear las consultas generadas por sqlc.
	queries := db.New(conn)

	// Ruta principal: mostrar productos.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		products, err := queries.ListProducts(r.Context())
		if err != nil {
			log.Println("Error al obtener productos:", err)
			http.Error(w, "Error al obtener productos", http.StatusInternalServerError)
			return
		}

		page := views.IndexPage("Lista de productos", products)

		if err := page.Render(r.Context(), w); err != nil {
			log.Println("Error al renderizar la página:", err)
		}
	})

	// Ruta para agregar productos.

	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		// Verificar que sea una petición POST.
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// 1. Parsear los datos del formulario.
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error al procesar el formulario", http.StatusBadRequest)
			return
		}

		// 2. Obtener los valores enviados.
		name := strings.TrimSpace(r.FormValue("name"))
		priceStr := r.FormValue("price")

		// 3. Validar el nombre.
		if name == "" {
			http.Error(w, "El nombre es obligatorio", http.StatusBadRequest)
			return
		}

		// 4. Convertir el precio a float64.
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || price < 0 {
			http.Error(w, "El precio debe ser un número válido y no negativo", http.StatusBadRequest)
			return
		}

		// 5. Guardar el producto utilizando sqlc.
		err = queries.CreateProduct(r.Context(), db.CreateProductParams{
			Name:  name,
			Price: price,
		})

		if err != nil {
			log.Println("Error al crear el producto:", err)
			http.Error(w, "Error al guardar el producto", http.StatusInternalServerError)
			return
		}

		// 6. Obtener la lista actualizada de productos.
		products, err := queries.ListProducts(r.Context())
		if err != nil {
			log.Println("Error al obtener productos:", err)
			http.Error(w, "Error al obtener productos", http.StatusInternalServerError)
			return
		}

		// 7. Renderizar la lista actualizada.
		if err := views.ProductList(products).Render(r.Context(), w); err != nil {
			log.Println("Error al renderizar la lista de productos:", err)
			return
		}

		// 8. Devolver un formulario vacío mediante Out-of-Band Swap.
		if err := views.ProductFormOOB().Render(r.Context(), w); err != nil {
			log.Println("Error al renderizar el formulario:", err)
			return
		}
	})

	// Ruta para eliminar productos.
	http.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		// Verificar que sea una petición DELETE.
		if r.Method != http.MethodDelete {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// 1. Extraer el ID desde la URL.
		idStr := strings.TrimPrefix(r.URL.Path, "/products/")

		// 2. Convertir el ID a int64.
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "ID de producto inválido", http.StatusBadRequest)
			return
		}

		// 3. Eliminar el producto utilizando sqlc.
		err = queries.DeleteProduct(r.Context(), id)
		if err != nil {
			log.Println("Error al eliminar el producto:", err)
			http.Error(w, "Error al eliminar el producto", http.StatusInternalServerError)
			return
		}

		// 4. Devolver una respuesta vacía con estado 200 OK.
		w.WriteHeader(http.StatusOK)
	})

	// 5. Iniciar el servidor HTTP.
	log.Println("Servidor iniciado en http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}
