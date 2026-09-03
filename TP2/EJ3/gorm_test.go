package main

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tp2.com/EJ3/models"
)

func TestGORM_CRUD(t *testing.T) {

	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=mi_db sslmode=disable"

	// Conectar a PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("error al conectar con PostgreSQL: %v", err)
	}

	// Crear/verificar el esquema
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("error en AutoMigrate: %v", err)
	}

	// Limpiar los registros antes de comenzar
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.User{})

	// =====================================================
	// CREAR
	// =====================================================

	user := models.User{
		Name:  "Usuario Test",
		Email: "usuario_test@gmail.com",
	}

	result := db.Create(&user)

	if result.Error != nil {
		t.Fatalf("error al crear usuario: %v", result.Error)
	}

	// Verificar que GORM haya generado un ID
	if user.ID == 0 {
		t.Error("se esperaba que el usuario tuviera un ID")
	}

	t.Logf("Usuario creado: %+v", user)

	// Guardamos el ID para utilizarlo en las siguientes operaciones
	userID := user.ID

	// =====================================================
	// LEER
	// =====================================================

	var readUser models.User

	result = db.First(&readUser, userID)

	if result.Error != nil {
		t.Fatalf("error al leer usuario: %v", result.Error)
	}

	if readUser.ID != userID {
		t.Errorf("ID incorrecto: esperado %d, obtenido %d",
			userID, readUser.ID)
	}

	if readUser.Name != "Usuario Test" {
		t.Errorf("nombre incorrecto: esperado %q, obtenido %q",
			"Usuario Test", readUser.Name)
	}

	if readUser.Email != "usuario_test@gmail.com" {
		t.Errorf("email incorrecto: esperado %q, obtenido %q",
			"usuario_test@gmail.com", readUser.Email)
	}

	t.Logf("Usuario leído: %+v", readUser)

	// =====================================================
	// ACTUALIZAR
	// =====================================================

	result = db.Model(&readUser).Update("Name", "Usuario Actualizado")

	if result.Error != nil {
		t.Fatalf("error al actualizar usuario: %v", result.Error)
	}

	// Volvemos a leer desde la BD para comprobar que realmente
	// se haya guardado el cambio
	var updatedUser models.User

	result = db.First(&updatedUser, userID)

	if result.Error != nil {
		t.Fatalf("error al leer usuario actualizado: %v", result.Error)
	}

	if updatedUser.Name != "Usuario Actualizado" {
		t.Errorf("nombre no actualizado: obtenido %q",
			updatedUser.Name)
	}

	t.Logf("Usuario actualizado: %+v", updatedUser)

	// =====================================================
	// ELIMINAR
	// =====================================================

	result = db.Delete(&models.User{}, userID)

	if result.Error != nil {
		t.Fatalf("error al eliminar usuario: %v", result.Error)
	}

	// Intentamos buscar nuevamente el usuario
	var deletedUser models.User

	result = db.First(&deletedUser, userID)

	// Como usamos gorm.Model, Delete hace soft delete.
	// Por eso First no encuentra el registro.
	if result.Error != gorm.ErrRecordNotFound {
		t.Errorf(
			"se esperaba gorm.ErrRecordNotFound después de eliminar, se obtuvo: %v",
			result.Error,
		)
	}

	t.Log("Usuario eliminado correctamente")
}
