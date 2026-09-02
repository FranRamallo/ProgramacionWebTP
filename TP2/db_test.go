package main

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"

	sqlc "tp2.com/TP2/internal/db"
)

func TestQueries_CRUD(t *testing.T) {

	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=mi_db sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("error al abrir la conexión: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("error al conectarse a PostgreSQL: %v", err)
	}

	queries := sqlc.New(db)

	ctx := context.Background()

	// ==========================================
	// 1. CREATE
	// ==========================================

	createdUser, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		Name:  "Usuario Test",
		Email: "usuario_test_sqlc@gmail.com",
	})

	if err != nil {
		t.Fatalf("error al crear usuario: %v", err)
	}

	if createdUser.Name != "Usuario Test" {
		t.Errorf("nombre incorrecto")
	}

	if createdUser.Email != "usuario_test_sqlc@gmail.com" {
		t.Errorf("email incorrecto")
	}

	t.Logf("Usuario creado: %+v", createdUser)

	// Guardamos el ID para todo el flujo
	userID := createdUser.ID

	// ==========================================
	// 2. GET
	// ==========================================

	t.Run("GetUser", func(t *testing.T) {

		user, err := queries.GetUserByID(ctx, userID)

		if err != nil {
			t.Fatalf("error al obtener usuario: %v", err)
		}

		if user.ID != userID {
			t.Errorf("ID incorrecto")
		}

		if user.Name != "Usuario Test" {
			t.Errorf("nombre incorrecto")
		}

		if user.Email != "usuario_test_sqlc@gmail.com" {
			t.Errorf("email incorrecto")
		}

		t.Logf("Usuario obtenido: %+v", user)
	})

	// ==========================================
	// 3. UPDATE
	// ==========================================

	t.Run("UpdateUser", func(t *testing.T) {

		user, err := queries.UpdateUser(ctx, sqlc.UpdateUserParams{
			ID:    userID,
			Name:  "Usuario Actualizado",
			Email: "usuario_actualizado_sqlc@gmail.com",
		})

		if err != nil {
			t.Fatalf("error al actualizar usuario: %v", err)
		}

		if user.Name != "Usuario Actualizado" {
			t.Errorf("el nombre no fue actualizado")
		}

		if user.Email != "usuario_actualizado_sqlc@gmail.com" {
			t.Errorf("el email no fue actualizado")
		}

		t.Logf("Usuario actualizado: %+v", user)
	})

	// ==========================================
	// 4. LIST
	// ==========================================

	t.Run("ListUsers", func(t *testing.T) {

		users, err := queries.ListUsers(ctx)

		if err != nil {
			t.Fatalf("error al listar usuarios: %v", err)
		}

		encontrado := false

		for _, user := range users {

			if user.ID == userID {

				encontrado = true

				if user.Name != "Usuario Actualizado" {
					t.Errorf("el usuario de la lista no está actualizado")
				}

				if user.Email != "usuario_actualizado_sqlc@gmail.com" {
					t.Errorf("el email de la lista no está actualizado")
				}

				break
			}
		}

		if !encontrado {
			t.Errorf("el usuario no fue encontrado en la lista")
		}
	})

	// ==========================================
	// 5. DELETE
	// ==========================================

	t.Run("DeleteUser", func(t *testing.T) {

		err := queries.DeleteUser(ctx, userID)

		if err != nil {
			t.Fatalf("error al eliminar usuario: %v", err)
		}

		// Intentamos obtener el usuario eliminado
		_, err = queries.GetUserByID(ctx, userID)

		if err != sql.ErrNoRows {
			t.Errorf(
				"se esperaba sql.ErrNoRows después de eliminar, se obtuvo: %v",
				err,
			)
		}
	})
}
