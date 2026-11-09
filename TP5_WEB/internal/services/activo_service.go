package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	db "github.com/FranRamallo/ProgramacionWebTP/db/sqlc"
	"github.com/lib/pq"
)

var ErrActivoNoEncontrado = errors.New("activo no encontrado")

type ActivoService struct {
	queries *db.Queries
}

func NewActivoService(queries *db.Queries) *ActivoService {
	return &ActivoService{queries: queries}
}

func (s *ActivoService) Crear(ctx context.Context, ticker, nombreEmpresa string, sectorID *int32, bolsa, tipoActivo string) (db.Activo, error) {
	// Limpieza de datos
	tickerLimpio := strings.ToUpper(strings.TrimSpace(ticker))
	nombreLimpio := strings.TrimSpace(nombreEmpresa)
	bolsaLimpia := strings.TrimSpace(bolsa)
	tipoLimpio := strings.ToUpper(strings.TrimSpace(tipoActivo))

	// Validaciones de negocio: Campos obligatorios
	if tickerLimpio == "" || nombreLimpio == "" || bolsaLimpia == "" || tipoLimpio == "" {
		return db.Activo{}, &ErrValidacion{"ticker, nombre, bolsa y tipo de activo son obligatorios"}
	}

	// Traducción de puntero a tipo de sqlc
	var nullSectorID sql.NullInt32
	if sectorID != nil {
		nullSectorID = sql.NullInt32{Int32: *sectorID, Valid: true}
	}

	// Inserción
	activo, err := s.queries.CreateActivo(ctx, db.CreateActivoParams{
		Ticker:        tickerLimpio,
		NombreEmpresa: nombreLimpio,
		SectorID:      nullSectorID,
		Bolsa:         bolsaLimpia,
		TipoActivo:    tipoLimpio,
	})

	// Traducción de errores de infraestructura a errores de dominio
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // foreign_key_violation
				return db.Activo{}, &ErrValidacion{"el sector indicado no existe en el sistema"}
			case "23505": // unique_violation
				return db.Activo{}, &ErrValidacion{"ya existe un activo registrado con ese ticker"}
			}
		}
		// Si es cualquier otro error (ej: se cayó la red), lo propagamos
		return db.Activo{}, err
	}

	return activo, nil
}

func (s *ActivoService) ObtenerPorID(ctx context.Context, id int64) (db.Activo, error) {
	activo, err := s.queries.GetActivo(ctx, id)

	if errors.Is(err, sql.ErrNoRows) {
		return db.Activo{}, ErrActivoNoEncontrado
	}
	if err != nil {
		return db.Activo{}, err
	}

	return activo, nil
}

func (s *ActivoService) Listar(ctx context.Context) ([]db.Activo, error) {
	items, err := s.queries.ListActivos(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []db.Activo{}, nil
	}
	return items, nil
}

func (s *ActivoService) Actualizar(ctx context.Context, id int64, ticker, nombreEmpresa string, sectorID *int32, bolsa, tipoActivo string) (db.Activo, error) {
	tickerLimpio := strings.ToUpper(strings.TrimSpace(ticker))
	nombreLimpio := strings.TrimSpace(nombreEmpresa)
	bolsaLimpia := strings.TrimSpace(bolsa)
	tipoLimpio := strings.ToUpper(strings.TrimSpace(tipoActivo))

	if tickerLimpio == "" || nombreLimpio == "" || bolsaLimpia == "" || tipoLimpio == "" {
		return db.Activo{}, &ErrValidacion{"ticker, nombre, bolsa y tipo de activo son obligatorios"}
	}

	var nullSectorID sql.NullInt32
	if sectorID != nil {
		nullSectorID = sql.NullInt32{Int32: *sectorID, Valid: true}
	}

	err := s.queries.UpdateActivo(ctx, db.UpdateActivoParams{
		ID:            id,
		Ticker:        tickerLimpio,
		NombreEmpresa: nombreLimpio,
		SectorID:      nullSectorID,
		Bolsa:         bolsaLimpia,
		TipoActivo:    tipoLimpio,
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return db.Activo{}, &ErrValidacion{"el sector indicado no existe en el sistema"}
			case "23505":
				return db.Activo{}, &ErrValidacion{"ya existe otro activo utilizando ese ticker"}
			}
		}
		return db.Activo{}, err
	}

	return s.ObtenerPorID(ctx, id)
}

func (s *ActivoService) Eliminar(ctx context.Context, id int64) error {
	// 1. Verificamos si existe primero reutilizando tu propia lógica
	_, err := s.ObtenerPorID(ctx, id)
	if err != nil {
		// Si no existe, esto automáticamente devuelve ErrActivoNoEncontrado
		return err
	}

	// 2. Si llegamos acá, sabemos que existe, procedemos a borrar
	err = s.queries.DeleteActivo(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
