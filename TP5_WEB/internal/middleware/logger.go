package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging es un middleware que intercepta las peticiones HTTP para registrar
// qué método se llamó, a qué ruta, y cuánto tiempo tardó en responder.
func Logging(next http.Handler) http.Handler {
	// Devolvemos una función anónima convertida al tipo http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Anotamos la hora exacta de inicio
		inicio := time.Now()

		// Pasamos la petición al siguiente eslabón de la cadena
		// (que puede ser otro middleware, o tu Handler final)
		next.ServeHTTP(w, r)

		// Calculamos cuánto tiempo pasó
		duracion := time.Since(inicio)

		// Imprimimos el resultado en la consola
		log.Printf("[API] %s %s | IP: %s | Tiempo: %v",
			r.Method, r.URL.Path, r.RemoteAddr, duracion)
	})
}
