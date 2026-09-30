package main

import (
	"TP3/logic"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var products = []logic.Product{
	{ID: 1, Name: "Product 1", Price: 10},
	{ID: 2, Name: "Product 2", Price: 20},
	{ID: 3, Name: "Product 3", Price: 30},
}

var mu sync.RWMutex

func createProduct(w http.ResponseWriter, r *http.Request) {
	var NewProduct logic.Product
	err := json.NewDecoder(r.Body).Decode(&NewProduct)
	//NewDecoder convierte Json a Go y lo asigna a NewProduct
	//r.Body es el cuerpo de la peticion http y lo decodifica en NewProduct
	//Decode toma el slide de NewProduct y lo convierte a Json
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = logic.ValidateProduct(NewProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	NewProduct.ID = len(products) + 1
	products = append(products, NewProduct)
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) //201 Created
	json.NewEncoder(w).Encode(NewProduct)
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products) //NewEncoder convierte Go a Json y lo envía al cliente
	//Encode toma el slide de products y lo convierte a Json
	mu.Unlock()
}

func findProductByID(id int) (*logic.Product, error) {
	for product := range products {
		if products[product].ID == id {
			return &products[product], nil
		}
	}
	return nil, errors.New("product not found")
}

func findProductIndexByID(id int) (int, error) {
	for i, product := range products {
		if product.ID == id {
			return i, nil
		}
	}
	return -1, errors.New("product not found")
}

func getProduct(w http.ResponseWriter, r *http.Request, id int) {
	mu.Lock()
	product, err := findProductByID(id)
	mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func updateProduct(w http.ResponseWriter, r *http.Request, id int) {
	var product logic.Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = logic.ValidateProduct(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	index, err := findProductIndexByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	//Mantener el id original
	product.ID = id
	products[index] = product
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func deleteProduct(w http.ResponseWriter, r *http.Request, id int) {
	mu.Lock()
	index, err := findProductIndexByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	products = append(products[:index], products[index+1:]...) //Eliminar el producto
	w.WriteHeader(http.StatusNoContent)
	defer mu.Unlock()
}

// /products/{id}
func productHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	//Extraer ID del path
	parts := strings.Split(path, "/")

	if len(parts) == 2 {
		switch r.Method {
		case http.MethodPost:
			createProduct(w, r)
		case http.MethodGet:
			getProducts(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if len(parts) == 3 && parts[1] == "products" {
		id, err := strconv.Atoi(parts[2]) //Agarra el id del path y lo pasa a int

		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			getProduct(w, r, id)
		case http.MethodPut:
			updateProduct(w, r, id)
		case http.MethodDelete:
			deleteProduct(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	http.Handle("/products/", loggingMiddleware(http.HandlerFunc(productHandler)))
	http.Handle("/products", loggingMiddleware(http.HandlerFunc(productHandler)))
	log.Println("Server started: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
