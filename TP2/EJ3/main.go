package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tp2.com/EJ3/models"
)

func main() {

	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=mi_db sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("error al conectar con PostgreSQL: %v", err)
	}

	// Migrar el esquema (en un proyecto real, revisar el error).
	if err := db.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}

	fmt.Println("Conexión a PostgreSQL exitosa")

	// Evitamos que Go considere que db no se está utilizando
	_ = db

	// Crear un usuario
	user := models.User{Name: "Jinzhu", Email: "jinzhu@cn.com"}
	result := db.Create(&user) // Pasa el puntero del objeto
	if result.Error != nil {
		log.Fatalf("failed to create user: %v", result.Error)
	}
	fmt.Printf("Usuario creado con ID: %d", user.ID)

	var readUser models.User
	// Obtener el primer registro ordenado por clave primaria
	if err := db.First(&readUser, user.ID).Error; err != nil { //db.First() para recuperar un registro por su clave primaria.
		log.Fatalf("failed to read user: %v", err)
	}
	fmt.Printf("Usuario leído: %+v ", readUser)

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("failed to list users: %v", err)
	}
	fmt.Printf("Todos los usuarios: %+v ", users)

	if err := db.Model(&readUser).Update("Name", "Jinzhu 2").Error; err != nil {
		log.Fatalf("failed to update user: %v", err)
	}
	fmt.Printf("Usuario actualizado: %+v ", readUser)

	if err := db.Delete(&models.User{}, readUser.ID).Error; err != nil {
		log.Fatalf("failed to delete user: %v", err)
	}
	fmt.Println("Usuario eliminado")
}
