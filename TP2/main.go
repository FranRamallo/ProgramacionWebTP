package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"tp2.com/TP2/repository"
)

func main() {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=mi_db sslmode=disable"

	db, err := sql.Open("postgres", dsn) //crea el objeto
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexión a PostgreSQL exitosa")

	repo := repository.NewUserRepository(db)

	user := &repository.User{
		Name:  "Fran",
		Email: "fran@gmail.com",
	}

	err = repo.CreateUser(user)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Usuario creado: %+v\n", user)
}
