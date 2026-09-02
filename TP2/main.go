package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	sqlc "tp2.com/TP2/internal/db" // generado por sqlc
	//_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=mi_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	defer db.Close()
	queries := sqlc.New(db)
	ctx := context.Background()
	createdUser, err := queries.CreateUser(ctx, // Create
		sqlc.CreateUserParams{
			Name:  "John Doe",
			Email: "john.doe@example.com",
		})
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	fmt.Printf("Created user: %+v\n", createdUser)
	user, err := queries.GetUserByID(ctx, createdUser.ID) // Read One
	if err != nil {
		log.Fatalf("failed to get user: %v", err)
	}
	fmt.Printf("Retrieved user: %+v\n", user)
	users, err := queries.ListUsers(ctx) // Read Many
	if err != nil {
		log.Fatalf("failed to list users: %v", err)
	}
	fmt.Printf("All users: %+v\n", users)
	updatedUser, err := queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:    createdUser.ID,
		Name:  "Johnny Doe",
		Email: "johnny.doe@example.com",
	})
	if err != nil {
		log.Fatalf("failed to update user: %v", err)
	}
	fmt.Println("User updated successfully")
	//updatedUser, err := queries.GetUserByID(ctx, createdUser.ID)
	if err != nil {
		log.Fatalf("failed to get updated user: %v", err)
	}
	fmt.Printf("Updated user: %+v\n", updatedUser)
	err = queries.DeleteUser(ctx, createdUser.ID) // Delete
	if err != nil {
		log.Fatalf("failed to delete user: %v", err)
	}
	fmt.Println("User deleted successfully")
	_, err = queries.GetUserByID(ctx, createdUser.ID)
	if err == sql.ErrNoRows {
		fmt.Println("User not found after deletion")
	} else if err != nil {
		log.Fatalf("failed to get user after deletion: %v", err)
	}
}
