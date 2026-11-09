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

var ErrMetricaFundamentalNoEncontrada = errors.New("métrica fundamental no encontrada")

type MetricaFundamentalService struct {
	queries *db.Queries
}

func NewMetricaFundamentalService(queries *db.Queries) *MetricaFundamentalService {
	return &MetricaFundamentalService{queries: queries}
}

// procesarMetricaOpcional valida que el string (si existe) sea un decimal válido y lo convierte a sql.NullString
func procesarMetricaOpcional(valor *string, nombreCampo string, permitirNegativos bool) (sql.NullString, error) {
	if valor == nil {
		return sql.NullString{Valid: false}, nil
	}
	limpio := strings.TrimSpace(*valor)
	if limpio == "" {
		return sql.NullString{Valid: false}, nil
	}

	num, err := strconv.ParseFloat(limpio, 64)
	if err != nil {
		return sql.NullString{}, &ErrValidacion{"el campo " + nombreCampo + " debe ser un número decimal válido"}
	}
	if !permitirNegativos && num < 0 {
		return sql.NullString{}, &ErrValidacion{"el campo " + nombreCampo + " no puede ser negativo"}
	}

	return sql.NullString{String: limpio, Valid: true}, nil
}

func (s *MetricaFundamentalService) Crear(ctx context.Context, activoID int64, fechaReporte time.Time, peRatio, dividendYield, deudaCapital, roe, eps *string) (db.MetricaFundamental, error) {
	// 1. Validación de campos obligatorios
	if activoID <= 0 {
		return db.MetricaFundamental{}, &ErrValidacion{"el activo_id es obligatorio y debe ser mayor a 0"}
	}
	if fechaReporte.IsZero() {
		return db.MetricaFundamental{}, &ErrValidacion{"la fecha_reporte es obligatoria"}
	}

	// 2. Validación y conversión de métricas opcionales
	peNull, err := procesarMetricaOpcional(peRatio, "pe_ratio", true)
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	divNull, err := procesarMetricaOpcional(dividendYield, "dividend_yield", false) // El dividendo no puede ser negativo
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	deudaNull, err := procesarMetricaOpcional(deudaCapital, "deuda_capital", true)
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	roeNull, err := procesarMetricaOpcional(roe, "roe", true)
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	epsNull, err := procesarMetricaOpcional(eps, "eps", true)
	if err != nil {
		return db.MetricaFundamental{}, err
	}

	// 3. Inserción en la base de datos
	metrica, err := s.queries.CreateMetricaFundamental(ctx, db.CreateMetricaFundamentalParams{
		ActivoID:      activoID,
		FechaReporte:  fechaReporte,
		PeRatio:       peNull,
		DividendYield: divNull,
		DeudaCapital:  deudaNull,
		Roe:           roeNull,
		Eps:           epsNull,
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // foreign_key_violation
				return db.MetricaFundamental{}, &ErrValidacion{"el activo indicado no existe en el sistema"}
			case "23505": // unique_violation (activo_id + fecha_reporte)
				return db.MetricaFundamental{}, &ErrValidacion{"ya existen métricas registradas para este activo en esa fecha de reporte"}
			case "22003": // numeric_value_out_of_range
				return db.MetricaFundamental{}, &ErrValidacion{"uno de los indicadores supera el límite de dígitos permitido"}
			}
		}
		return db.MetricaFundamental{}, err
	}

	return metrica, nil
}

func (s *MetricaFundamentalService) ObtenerPorID(ctx context.Context, id int64) (db.MetricaFundamental, error) {
	metrica, err := s.queries.GetMetricaFundamental(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.MetricaFundamental{}, ErrMetricaFundamentalNoEncontrada
	}
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	return metrica, nil
}

func (s *MetricaFundamentalService) ListarPorActivo(ctx context.Context, activoID int64) ([]db.MetricaFundamental, error) {
	if activoID <= 0 {
		return nil, &ErrValidacion{"el activo_id debe ser mayor a 0"}
	}
	items, err := s.queries.ListMetricasFundamentales(ctx, activoID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []db.MetricaFundamental{}, nil
	}
	return items, nil
}

func (s *MetricaFundamentalService) Actualizar(ctx context.Context, id int64, peRatio, dividendYield, deudaCapital, roe, eps *string) (db.MetricaFundamental, error) {
	// Validamos y convertimos los campos opcionales
	peNull, err := procesarMetricaOpcional(peRatio, "pe_ratio", true)
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	divNull, err := procesarMetricaOpcional(dividendYield, "dividend_yield", false)
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	deudaNull, err := procesarMetricaOpcional(deudaCapital, "deuda_capital", true)
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	roeNull, err := procesarMetricaOpcional(roe, "roe", true)
	if err != nil {
		return db.MetricaFundamental{}, err
	}
	epsNull, err := procesarMetricaOpcional(eps, "eps", true)
	if err != nil {
		return db.MetricaFundamental{}, err
	}

	err = s.queries.UpdateMetricaFundamental(ctx, db.UpdateMetricaFundamentalParams{
		ID:            id,
		PeRatio:       peNull,
		DividendYield: divNull,
		DeudaCapital:  deudaNull,
		Roe:           roeNull,
		Eps:           epsNull,
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22003" {
				return db.MetricaFundamental{}, &ErrValidacion{"uno de los indicadores supera el límite de dígitos permitido"}
			}
		}
		return db.MetricaFundamental{}, err
	}

	return s.ObtenerPorID(ctx, id)
}

func (s *MetricaFundamentalService) Eliminar(ctx context.Context, id int64) error {
	// 1. Verificamos si existe primero reutilizando ObtenerPorID
	_, err := s.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Procedemos a borrar
	err = s.queries.DeleteMetricaFundamental(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
