package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	db "github.com/FranRamallo/ProgramacionWebTP/db/sqlc"
)

var ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")

type UsuarioService struct {
	queries *db.Queries
}

func NewUsuarioService(queries *db.Queries) *UsuarioService {
	return &UsuarioService{queries: queries}
}

func (s *UsuarioService) Crear(ctx context.Context, nombreCompleto, email, passwordPlana string) (db.Usuario, error) {
	nombreLimpio := strings.TrimSpace(nombreCompleto)
	emailLimpio := strings.ToLower(strings.TrimSpace(email)) // Los correos siempre en minúscula
	passwordLimpia := strings.TrimSpace(passwordPlana)

	// Validaciones de campos obligatorios
	if nombreLimpio == "" || emailLimpio == "" || passwordLimpia == "" {
		return db.Usuario{}, &ErrValidacion{"el nombre, email y contraseña son obligatorios"}
	}
	if !strings.Contains(emailLimpio, "@") {
		return db.Usuario{}, &ErrValidacion{"el formato del correo electrónico no es válido"}
	}

	// Hasheamos la contraseña con Bcrypt (costo 10 es el estándar recomendado)
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(passwordLimpia), 10)
	if err != nil {
		return db.Usuario{}, errors.New("error interno al encriptar la contraseña")
	}

	usuario, err := s.queries.CreateUsuario(ctx, db.CreateUsuarioParams{
		NombreCompleto: nombreLimpio,
		Email:          emailLimpio,
		PasswordHash:   string(hashBytes),
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			// Atrapamos la restricción Unique Key [UK] del email[cite: 2]
			if pgErr.Code == "23505" {
				return db.Usuario{}, &ErrValidacion{"ya existe una cuenta registrada con este email"}
			}
		}
		return db.Usuario{}, err
	}

	return usuario, nil
}

func (s *UsuarioService) ObtenerPorID(ctx context.Context, id uuid.UUID) (db.Usuario, error) {
	usuario, err := s.queries.GetUsuario(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.Usuario{}, ErrUsuarioNoEncontrado
	}
	if err != nil {
		return db.Usuario{}, err
	}
	return usuario, nil
}

func (s *UsuarioService) Listar(ctx context.Context) ([]db.Usuario, error) {
	items, err := s.queries.ListUsuarios(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []db.Usuario{}, nil
	}
	return items, nil
}

func (s *UsuarioService) Actualizar(ctx context.Context, id uuid.UUID, nombreCompleto, email string) (db.Usuario, error) {
	nombreLimpio := strings.TrimSpace(nombreCompleto)
	emailLimpio := strings.ToLower(strings.TrimSpace(email))

	if nombreLimpio == "" || emailLimpio == "" {
		return db.Usuario{}, &ErrValidacion{"el nombre y el email son obligatorios"}
	}
	if !strings.Contains(emailLimpio, "@") {
		return db.Usuario{}, &ErrValidacion{"el formato del correo electrónico no es válido"}
	}

	err := s.queries.UpdateUsuario(ctx, db.UpdateUsuarioParams{
		ID:             id,
		NombreCompleto: nombreLimpio,
		Email:          emailLimpio,
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return db.Usuario{}, &ErrValidacion{"ya existe otra cuenta utilizando este email"}
			}
		}
		return db.Usuario{}, err
	}

	return s.ObtenerPorID(ctx, id)
}

func (s *UsuarioService) Eliminar(ctx context.Context, id uuid.UUID) error {
	// Verificamos si existe primero reutilizando tu propia lógica
	_, err := s.ObtenerPorID(ctx, id)
	if err != nil {
		// Si no existe, esto automáticamente devuelve ErrUsuarioNoEncontrado
		return err
	}

	// Si llegamos acá, sabemos que existe, procedemos a borrar
	err = s.queries.DeleteUsuario(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
