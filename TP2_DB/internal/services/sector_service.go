package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	db "github.com/FranRamallo/ProgramacionWebTP/db/sqlc"
	"github.com/lib/pq"
)

// ErrSectorNoEncontrado es un error de dominio. El Handler decidirá devolver un 404.
var ErrSectorNoEncontrado = errors.New("sector no encontrado")

type SectorService struct {
	queries *db.Queries
}

func NewSectorService(queries *db.Queries) *SectorService {
	return &SectorService{queries: queries}
}

// Crear recibe directamente el string en lugar de un struct
func (s *SectorService) Crear(ctx context.Context, nombre string) (db.Sector, error) {
	nombreLimpio := strings.TrimSpace(nombre)
	if nombreLimpio == "" {
		return db.Sector{}, &ErrValidacion{"el nombre del sector es obligatorio"}
	}

	sector, err := s.queries.CreateSector(ctx, nombreLimpio)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return db.Sector{}, &ErrValidacion{"ya existe un sector registrado con ese nombre"}
			}
		}
		// Cualquier otro error de base de datos se propaga
		return db.Sector{}, err
	}

	return sector, nil
}

func (s *SectorService) ObtenerPorID(ctx context.Context, id int32) (db.Sector, error) {
	sector, err := s.queries.GetSector(ctx, id)

	// Traducción de error de infraestructura (sql.ErrNoRows) a error de dominio
	if errors.Is(err, sql.ErrNoRows) {
		return db.Sector{}, ErrSectorNoEncontrado
	}
	if err != nil {
		return db.Sector{}, err
	}

	return sector, nil
}

func (s *SectorService) Listar(ctx context.Context) ([]db.Sector, error) {
	items, err := s.queries.ListSectores(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []db.Sector{}, nil
	}
	return items, nil
}

// Actualizar recibe el ID y los campos modificables directamente
func (s *SectorService) Actualizar(ctx context.Context, id int32, nombre string) (db.Sector, error) {
	nombreLimpio := strings.TrimSpace(nombre)
	if nombreLimpio == "" {
		return db.Sector{}, &ErrValidacion{"el nombre del sector es obligatorio"}
	}
	update := db.UpdateSectorParams{
		ID:     id,
		Nombre: nombreLimpio,
	}

	err := s.queries.UpdateSector(ctx, update)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return db.Sector{}, &ErrValidacion{"ya existe otro sector utilizando ese nombre"}
			}
		}
		return db.Sector{}, err
	}

	return s.ObtenerPorID(ctx, id)
}

func (s *SectorService) Eliminar(ctx context.Context, id int32) error {
	// 1. Verificamos si existe primero reutilizando tu propia lógica
	_, err := s.ObtenerPorID(ctx, id)
	if err != nil {
		// Si no existe, esto automáticamente devuelve ErrSectorNoEncontrado
		return err
	}

	// 2. Si llegamos acá, sabemos que existe, procedemos a borrar
	err = s.queries.DeleteSector(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
