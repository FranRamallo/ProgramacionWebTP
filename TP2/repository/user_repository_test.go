package repository

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func TestUserRepository_CRUD(t *testing.T) {

	// Conexión a PostgreSQL
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=mi_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("error al abrir la conexión: %v", err)
	}
	defer db.Close()

	// Verificamos que podamos conectarnos
	if err := db.Ping(); err != nil {
		t.Fatalf("error al conectarse a PostgreSQL: %v", err)
	}

	// Instanciamos el repositorio
	repo := NewUserRepository(db)

	// ==========================================
	// 1. CREAR
	// ==========================================

	var userID int

	t.Run("Crear usuario", func(t *testing.T) {

		user := &User{
			Name:  "Usuario Test",
			Email: "usuario_test@gmail.com",
		}

		err := repo.CreateUser(user)

		if err != nil {
			t.Fatalf("error al crear usuario: %v", err)
		}

		if user.ID == 0 {
			t.Errorf("se esperaba que el usuario tuviera un ID")
		}

		userID = user.ID

		t.Logf("Usuario creado con ID: %d", userID)
	})

	// ==========================================
	// 2. LEER
	// ==========================================

	t.Run("Obtener usuario", func(t *testing.T) {

		user, err := repo.GetUserByID(userID)

		if err != nil {
			t.Fatalf("error al obtener usuario: %v", err)
		}

		if user.Name != "Usuario Test" {
			t.Errorf(
				"nombre incorrecto: se esperaba %q, se obtuvo %q",
				"Usuario Test",
				user.Name,
			)
		}

		if user.Email != "usuario_test@gmail.com" {
			t.Errorf(
				"email incorrecto: se esperaba %q, se obtuvo %q",
				"usuario_test@gmail.com",
				user.Email,
			)
		}

		t.Logf("Usuario obtenido correctamente: %+v", user)
	})

	// ==========================================
	// 3. ACTUALIZAR
	// ==========================================

	t.Run("Actualizar usuario", func(t *testing.T) {

		user, err := repo.GetUserByID(userID)

		if err != nil {
			t.Fatalf("error al obtener usuario antes de actualizar: %v", err)
		}

		// Cambiamos el nombre
		user.Name = "Usuario Actualizado"

		err = repo.UpdateUser(user)

		if err != nil {
			t.Fatalf("error al actualizar usuario: %v", err)
		}

		// Volvemos a obtenerlo para comprobar que se actualizó
		updatedUser, err := repo.GetUserByID(userID)

		if err != nil {
			t.Fatalf("error al obtener usuario actualizado: %v", err)
		}

		if updatedUser.Name != "Usuario Actualizado" {
			t.Errorf(
				"nombre no actualizado: se esperaba %q, se obtuvo %q",
				"Usuario Actualizado",
				updatedUser.Name,
			)
		}

		t.Logf("Usuario actualizado correctamente: %+v", updatedUser)
	})

	// ==========================================
	// 4. LISTAR
	// ==========================================

	t.Run("Listar usuarios", func(t *testing.T) {

		users, err := repo.ListUsers()

		if err != nil {
			t.Fatalf("error al listar usuarios: %v", err)
		}

		encontrado := false

		for _, user := range users {
			if user.ID == userID {
				encontrado = true

				if user.Name != "Usuario Actualizado" {
					t.Errorf(
						"nombre incorrecto en la lista: se esperaba %q, se obtuvo %q",
						"Usuario Actualizado",
						user.Name,
					)
				}

				break
			}
		}

		if !encontrado {
			t.Errorf("el usuario con ID %d no fue encontrado en la lista", userID)
		}
	})

	// ==========================================
	// 5. ELIMINAR
	// ==========================================

	t.Run("Eliminar usuario", func(t *testing.T) {

		err := repo.DeleteUser(userID)

		if err != nil {
			t.Fatalf("error al eliminar usuario: %v", err)
		}

		// Intentamos obtenerlo nuevamente
		_, err = repo.GetUserByID(userID)

		if err != sql.ErrNoRows {
			t.Errorf(
				"se esperaba sql.ErrNoRows después de eliminar, se obtuvo: %v",
				err,
			)
		}
	})
}
