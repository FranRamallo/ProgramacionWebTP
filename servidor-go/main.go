package main

import (
	"fmt"
	"net/http"
)

// serveForm: Maneja GET / para mostrar el formulario
// func serveForm(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/" {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 	fmt.Fprint(w, loginForm)
// }

// handleLogin: Maneja POST /contacto para procesar los datos
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parsear los datos del formulario
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error al procesar formulario", http.StatusBadRequest)
		return
	}

	// 2. Extraer los valores (atento al name="user" del input)
	username := r.FormValue("user")
	email := r.FormValue("email")
	messageText := r.FormValue("message")

	// Validación básica
	if username == "" || email == "" || messageText == "" {
		http.Error(w, "Campos incompletos: los campos son obligatorios", http.StatusBadRequest)
		return
	}

	// 3. Generar y enviar respuesta HTML dinámica
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Bienvenido</title></head>
<body>
	<h1>¡Hola, %s!</h1>
	<p>Tu email es %s</p>
	<p>Mensaje: %s</p>
	<a href="/">Volver al inicio</a>
</body>
</html>`, username, email, messageText)
}

func main() {
	//http.HandleFunc("/", serveForm)
	loginForm := "./static"
	file := http.FileServer(http.Dir(loginForm))
	http.Handle("/", file)
	http.HandleFunc("/contacto", handleLogin)

	port := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Error fatal en el servidor: %s\n", err)
	}
}
