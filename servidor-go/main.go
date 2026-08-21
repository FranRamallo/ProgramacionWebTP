package main

import (
	"fmt"
	"net/http"
)

func main() {
	//http.HandleFunc("/", HandleIndex)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	port := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Error fatal en el servidor: %s\n", err)
	}
}
