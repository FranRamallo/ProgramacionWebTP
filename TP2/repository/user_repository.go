package repository

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
}

type UserRepository struct {
	db *sql.DB //representa el acceso a la base de datos.
}

// construir el repositorio
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Crear usuario
func (r *UserRepository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (name, email)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	// QueryRow hace que el driver se encargue de pasar los valores como parámetros.
	return r.db.QueryRow(
		query,
		user.Name,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
}

// Buscar por id
func (r *UserRepository) GetUserByID(id int) (*User, error) {
	query := `
		SELECT id, name, email, created_at
		FROM users
		WHERE id = $1
	`

	user := &User{}

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	return user, nil
}

// Listar usuarios
func (r *UserRepository) ListUsers() ([]*User, error) {
	query := `
		SELECT id, name, email, created_at
		FROM users
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //garantiza que los recursos asociados a rows se liberen cuando termine la función.

	var users []*User

	for rows.Next() {
		user := &User{}

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Actualizar usuario
func (r *UserRepository) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(
		query,
		user.Name,
		user.Email,
		user.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected() //nos permite saber si realmente se encontró un usuario para modificar.
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Eleminar usuario
func (r *UserRepository) DeleteUser(id int) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
