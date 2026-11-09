package services

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"

	db "github.com/FranRamallo/ProgramacionWebTP/db/sqlc"
)

var ErrCotizacionActualNoEncontrada = errors.New("cotización actual no encontrada para este activo")

type CotizacionActualService struct {
	queries *db.Queries
}

func NewCotizacionActualService(queries *db.Queries) *CotizacionActualService {
	return &CotizacionActualService{queries: queries}
}

// Guardar realiza un Upsert: crea la cotización actual si no existe, o la actualiza si ya existe
func (s *CotizacionActualService) Guardar(ctx context.Context, activoID int64, precio, cierreAnterior, variacionDiaPct string, actualizadoEn *time.Time) (db.CotizacionActual, error) {
	precioLimpio := strings.TrimSpace(precio)
	cierreLimpio := strings.TrimSpace(cierreAnterior)
	variacionLimpia := strings.TrimSpace(variacionDiaPct)

	// Validación de campos obligatorios
	if activoID <= 0 || precioLimpio == "" || cierreLimpio == "" || variacionLimpia == "" {
		return db.CotizacionActual{}, &ErrValidacion{"activo_id, precio, cierre_anterior y variacion_dia_pct son obligatorios"}
	}

	// Validación numérica de los campos DECIMAL
	if pVal, err := strconv.ParseFloat(precioLimpio, 64); err != nil || pVal < 0 {
		return db.CotizacionActual{}, &ErrValidacion{"el precio debe ser un número válido mayor o igual a 0"}
	}
	if cVal, err := strconv.ParseFloat(cierreLimpio, 64); err != nil || cVal < 0 {
		return db.CotizacionActual{}, &ErrValidacion{"el cierre anterior debe ser un número válido mayor o igual a 0"}
	}
	// La variación porcentual del día puede ser negativa
	if _, err := strconv.ParseFloat(variacionLimpia, 64); err != nil {
		return db.CotizacionActual{}, &ErrValidacion{"la variación del día debe ser un número decimal válido"}
	}

	// Si no envían fecha de actualización, usamos la hora actual del servidor
	fechaActualizacion := time.Now()
	if actualizadoEn != nil && !actualizadoEn.IsZero() {
		fechaActualizacion = *actualizadoEn
	}

	// Ejecutamos el Upsert generado por sqlc
	cotizacion, err := s.queries.UpsertCotizacionActual(ctx, db.UpsertCotizacionActualParams{
		ActivoID:        activoID,
		Precio:          precioLimpio,
		CierreAnterior:  cierreLimpio,
		VariacionDiaPct: variacionLimpia,
		ActualizadoEn:   fechaActualizacion,
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // foreign_key_violation
				return db.CotizacionActual{}, &ErrValidacion{"el activo indicado no existe en el sistema"}
			case "22003": // numeric_value_out_of_range
				return db.CotizacionActual{}, &ErrValidacion{"uno de los valores supera el límite numérico permitido"}
			}
		}
		return db.CotizacionActual{}, err
	}

	return cotizacion, nil
}

func (s *CotizacionActualService) ObtenerPorActivoID(ctx context.Context, activoID int64) (db.CotizacionActual, error) {
	cotizacion, err := s.queries.GetCotizacionActual(ctx, activoID)
	if errors.Is(err, sql.ErrNoRows) {
		return db.CotizacionActual{}, ErrCotizacionActualNoEncontrada
	}
	if err != nil {
		return db.CotizacionActual{}, err
	}
	return cotizacion, nil
}

func (s *CotizacionActualService) Listar(ctx context.Context) ([]db.CotizacionActual, error) {
	items, err := s.queries.ListCotizacionesActuales(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []db.CotizacionActual{}, nil
	}
	return items, nil
}

func (s *CotizacionActualService) Eliminar(ctx context.Context, activoID int64) error {
	// 1. Verificamos si existe primero reutilizando ObtenerPorActivoID
	_, err := s.ObtenerPorActivoID(ctx, activoID)
	if err != nil {
		return err
	}

	// 2. Procedemos a borrar
	err = s.queries.DeleteCotizacionActual(ctx, activoID)
	if err != nil {
		return err
	}

	return nil
}
