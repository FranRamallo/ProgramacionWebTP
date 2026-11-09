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

var ErrCotizacionHistoricaNoEncontrada = errors.New("cotización histórica no encontrada")

type CotizacionHistoricaService struct {
	queries *db.Queries
}

func NewCotizacionHistoricaService(queries *db.Queries) *CotizacionHistoricaService {
	return &CotizacionHistoricaService{queries: queries}
}

// validarPreciosYVolumen centraliza las reglas de negocio de una vela diaria (OHLC + Volumen)
func validarPreciosYVolumen(aperturaStr, maxStr, minStr, cierreStr string, volumen int64) (string, string, string, string, error) {
	apLimpio := strings.TrimSpace(aperturaStr)
	maxLimpio := strings.TrimSpace(maxStr)
	minLimpio := strings.TrimSpace(minStr)
	cieLimpio := strings.TrimSpace(cierreStr)

	if apLimpio == "" || maxLimpio == "" || minLimpio == "" || cieLimpio == "" {
		return "", "", "", "", &ErrValidacion{"precio_apertura, precio_max, precio_min y precio_cierre son obligatorios"}
	}

	ap, err := strconv.ParseFloat(apLimpio, 64)
	if err != nil || ap < 0 {
		return "", "", "", "", &ErrValidacion{"el precio_apertura debe ser un número válido mayor o igual a 0"}
	}
	max, err := strconv.ParseFloat(maxLimpio, 64)
	if err != nil || max < 0 {
		return "", "", "", "", &ErrValidacion{"el precio_max debe ser un número válido mayor o igual a 0"}
	}
	min, err := strconv.ParseFloat(minLimpio, 64)
	if err != nil || min < 0 {
		return "", "", "", "", &ErrValidacion{"el precio_min debe ser un número válido mayor o igual a 0"}
	}
	cie, err := strconv.ParseFloat(cieLimpio, 64)
	if err != nil || cie < 0 {
		return "", "", "", "", &ErrValidacion{"el precio_cierre debe ser un número válido mayor o igual a 0"}
	}

	// Coherencia financiera: el mínimo no puede superar al máximo, y apertura/cierre deben estar dentro del rango [min, max]
	if min > max {
		return "", "", "", "", &ErrValidacion{"el precio_min no puede ser mayor que el precio_max"}
	}
	if ap < min || ap > max || cie < min || cie > max {
		return "", "", "", "", &ErrValidacion{"los precios de apertura y cierre deben encontrarse entre precio_min y precio_max"}
	}

	if volumen < 0 {
		return "", "", "", "", &ErrValidacion{"el volumen no puede ser negativo"}
	}

	return apLimpio, maxLimpio, minLimpio, cieLimpio, nil
}

func (s *CotizacionHistoricaService) Crear(ctx context.Context, activoID int64, fecha time.Time, precioApertura, precioMax, precioMin, precioCierre string, volumen int64) (db.CotizacionHistorica, error) {
	if activoID <= 0 {
		return db.CotizacionHistorica{}, &ErrValidacion{"el activo_id es obligatorio y debe ser mayor a 0"}
	}
	if fecha.IsZero() {
		return db.CotizacionHistorica{}, &ErrValidacion{"la fecha es obligatoria"}
	}

	apLimpio, maxLimpio, minLimpio, cieLimpio, err := validarPreciosYVolumen(precioApertura, precioMax, precioMin, precioCierre, volumen)
	if err != nil {
		return db.CotizacionHistorica{}, err
	}

	cotizacion, err := s.queries.CreateCotizacionHistorica(ctx, db.CreateCotizacionHistoricaParams{
		ActivoID:       activoID,
		Fecha:          fecha,
		PrecioApertura: apLimpio,
		PrecioMax:      maxLimpio,
		PrecioMin:      minLimpio,
		PrecioCierre:   cieLimpio,
		Volumen:        volumen,
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // foreign_key_violation
				return db.CotizacionHistorica{}, &ErrValidacion{"el activo indicado no existe en el sistema"}
			case "23505": // unique_violation (activo_id + fecha)
				return db.CotizacionHistorica{}, &ErrValidacion{"ya existe una cotización histórica para este activo en esa fecha"}
			case "22003": // numeric_value_out_of_range
				return db.CotizacionHistorica{}, &ErrValidacion{"uno de los precios supera el límite de dígitos permitido"}
			}
		}
		return db.CotizacionHistorica{}, err
	}

	return cotizacion, nil
}

func (s *CotizacionHistoricaService) ObtenerPorID(ctx context.Context, id int64) (db.CotizacionHistorica, error) {
	cotizacion, err := s.queries.GetCotizacionHistorica(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.CotizacionHistorica{}, ErrCotizacionHistoricaNoEncontrada
	}
	if err != nil {
		return db.CotizacionHistorica{}, err
	}
	return cotizacion, nil
}

func (s *CotizacionHistoricaService) ListarPorActivo(ctx context.Context, activoID int64) ([]db.CotizacionHistorica, error) {
	if activoID <= 0 {
		return nil, &ErrValidacion{"el activo_id debe ser mayor a 0"}
	}
	items, err := s.queries.ListCotizacionesHistoricas(ctx, activoID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []db.CotizacionHistorica{}, nil
	}
	return items, nil
}

func (s *CotizacionHistoricaService) Actualizar(ctx context.Context, id int64, precioApertura, precioMax, precioMin, precioCierre string, volumen int64) (db.CotizacionHistorica, error) {
	if id <= 0 {
		return db.CotizacionHistorica{}, &ErrValidacion{"el id de la cotización debe ser mayor a 0"}
	}

	apLimpio, maxLimpio, minLimpio, cieLimpio, err := validarPreciosYVolumen(precioApertura, precioMax, precioMin, precioCierre, volumen)
	if err != nil {
		return db.CotizacionHistorica{}, err
	}

	// Verificamos que el registro exista antes de actualizar
	_, err = s.ObtenerPorID(ctx, id)
	if err != nil {
		return db.CotizacionHistorica{}, err
	}

	err = s.queries.UpdateCotizacionHistorica(ctx, db.UpdateCotizacionHistoricaParams{
		ID:             id,
		PrecioApertura: apLimpio,
		PrecioMax:      maxLimpio,
		PrecioMin:      minLimpio,
		PrecioCierre:   cieLimpio,
		Volumen:        volumen,
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22003" {
				return db.CotizacionHistorica{}, &ErrValidacion{"uno de los precios supera el límite de dígitos permitido"}
			}
		}
		return db.CotizacionHistorica{}, err
	}

	return s.ObtenerPorID(ctx, id)
}

func (s *CotizacionHistoricaService) Eliminar(ctx context.Context, id int64) error {
	// Verificamos si existe primero reutilizando ObtenerPorID
	_, err := s.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Procedemos a borrar
	err = s.queries.DeleteCotizacionHistorica(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
