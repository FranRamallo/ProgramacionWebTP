package models

import "gorm.io/gorm"

type User struct {
	gorm.Model // Incluye ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string
	Email      string `gorm:"unique"` // Crea una restricción UNIQUE en la BD
}
