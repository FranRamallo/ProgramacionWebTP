package services

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	db "github.com/FranRamallo/ProgramacionWebTP/db/sqlc"
)

var ErrTransaccionNoEncontrada = errors.New("transacción no encontrada")

type TransaccionService struct {
	queries *db.Queries
}

func NewTransaccionService(queries *db.Queries) *TransaccionService {
	return &TransaccionService{queries: queries}
}

func (s *TransaccionService) Crear(ctx context.Context, usuarioID uuid.UUID, activoID int64, tipoOperacion, cantidad, precioUnitario, comision string, fechaOperacion *time.Time, notas *string) (db.Transaccion, error) {
	tipoLimpio := strings.ToLower(strings.TrimSpace(tipoOperacion))
	cantLimpia := strings.TrimSpace(cantidad)
	precioLimpio := strings.TrimSpace(precioUnitario)
	comisionLimpia := strings.TrimSpace(comision)

	// Si no envían comisión, asumimos "0" por defecto
	if comisionLimpia == "" {
		comisionLimpia = "0"
	}

	// Validación de campos obligatorios y CHECK de tipo_operacion ('compra' o 'venta')
	if activoID <= 0 || tipoLimpio == "" || cantLimpia == "" || precioLimpio == "" {
		return db.Transaccion{}, &ErrValidacion{"activo_id, tipo_operacion, cantidad y precio_unitario son obligatorios"}
	}
	if tipoLimpio != "compra" && tipoLimpio != "venta" {
		return db.Transaccion{}, &ErrValidacion{"el tipo de operación debe ser 'compra' o 'venta'"}
	}

	// Validación numérica de los campos DECIMAL
	if cantVal, err := strconv.ParseFloat(cantLimpia, 64); err != nil || cantVal <= 0 {
		return db.Transaccion{}, &ErrValidacion{"la cantidad debe ser un número válido mayor a 0"}
	}
	if precioVal, err := strconv.ParseFloat(precioLimpio, 64); err != nil || precioVal < 0 {
		return db.Transaccion{}, &ErrValidacion{"el precio unitario debe ser un número válido mayor o igual a 0"}
	}
	if comVal, err := strconv.ParseFloat(comisionLimpia, 64); err != nil || comVal < 0 {
		return db.Transaccion{}, &ErrValidacion{"la comisión debe ser un número válido mayor o igual a 0"}
	}

	// Traducción de punteros opcionales a estructuras sql.Null*
	// Si no envían fecha, asignamos time.Now() para evitar que se inserte NULL
	fechaNull := sql.NullTime{Time: time.Now(), Valid: true}
	if fechaOperacion != nil {
		fechaNull = sql.NullTime{Time: *fechaOperacion, Valid: true}
	}

	var notasNull sql.NullString
	if notas != nil {
		notasTrim := strings.TrimSpace(*notas)
		if notasTrim != "" {
			notasNull = sql.NullString{String: notasTrim, Valid: true}
		}
	}

	// Inserción en base de datos
	transaccion, err := s.queries.CreateTransaccion(ctx, db.CreateTransaccionParams{
		UsuarioID:      usuarioID,
		ActivoID:       activoID,
		TipoOperacion:  tipoLimpio,
		Cantidad:       cantLimpia,
		PrecioUnitario: precioLimpio,
		Comision:       comisionLimpia,
		FechaOperacion: fechaNull,
		Notas:          notasNull,
	})

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // foreign_key_violation (usuario_id o activo_id inexistentes)
				return db.Transaccion{}, &ErrValidacion{"el usuario o el activo indicado no existe en el sistema"}
			case "22003": // numeric_value_out_of_range (supera los dígitos del DECIMAL)
				return db.Transaccion{}, &ErrValidacion{"uno de los montos supera el límite numérico permitido"}
			}
		}
		return db.Transaccion{}, err
	}

	return transaccion, nil
}

func (s *TransaccionService) ObtenerPorID(ctx context.Context, id, usuarioID uuid.UUID) (db.Transaccion, error) {
	transaccion, err := s.queries.GetTransaccion(ctx, db.GetTransaccionParams{
		ID:        id,
		UsuarioID: usuarioID, //Por seguridad se pide el usuario para el get generado por sqlc automaticamente
	})
	if errors.Is(err, sql.ErrNoRows) {
		return db.Transaccion{}, ErrTransaccionNoEncontrada
	}
	if err != nil {
		return db.Transaccion{}, err
	}
	return transaccion, nil
}

func (s *TransaccionService) ListarPorUsuario(ctx context.Context, usuarioID uuid.UUID) ([]db.Transaccion, error) {
	items, err := s.queries.ListTransaccionesPorUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []db.Transaccion{}, nil
	}
	return items, nil
}

func (s *TransaccionService) Actualizar(
	ctx context.Context,
	id, usuarioID uuid.UUID,
	tipoOperacion, cantidad, precioUnitario, comision string,
	fechaOperacion *time.Time,
	notas *string,
) (db.Transaccion, error) {
	tipoLimpio := strings.ToLower(strings.TrimSpace(tipoOperacion))
	cantLimpia := strings.TrimSpace(cantidad)
	precioLimpio := strings.TrimSpace(precioUnitario)
	comisionLimpia := strings.TrimSpace(comision)

	if comisionLimpia == "" {
		comisionLimpia = "0"
	}

	if tipoLimpio == "" || cantLimpia == "" || precioLimpio == "" {
		return db.Transaccion{}, &ErrValidacion{"tipo_operacion, cantidad y precio_unitario son obligatorios"}
	}
	if tipoLimpio != "compra" && tipoLimpio != "venta" {
		return db.Transaccion{}, &ErrValidacion{"el tipo de operación debe ser 'compra' o 'venta'"}
	}

	if cantVal, err := strconv.ParseFloat(cantLimpia, 64); err != nil || cantVal <= 0 {
		return db.Transaccion{}, &ErrValidacion{"la cantidad debe ser un número válido mayor a 0"}
	}
	if precioVal, err := strconv.ParseFloat(precioLimpio, 64); err != nil || precioVal < 0 {
		return db.Transaccion{}, &ErrValidacion{"el precio unitario debe ser un número válido mayor o igual a 0"}
	}
	if comVal, err := strconv.ParseFloat(comisionLimpia, 64); err != nil || comVal < 0 {
		return db.Transaccion{}, &ErrValidacion{"la comisión debe ser un número válido mayor o igual a 0"}
	}

	// Primero verificamos que exista (y nos sirve para conservar la fecha original si no envían una nueva)
	actual, err := s.ObtenerPorID(ctx, id, usuarioID)
	if err != nil {
		return db.Transaccion{}, err
	}

	fechaNull := actual.FechaOperacion
	if fechaOperacion != nil {
		fechaNull = sql.NullTime{Time: *fechaOperacion, Valid: true}
	}

	var notasNull sql.NullString
	if notas != nil {
		notasTrim := strings.TrimSpace(*notas)
		if notasTrim != "" {
			notasNull = sql.NullString{String: notasTrim, Valid: true}
		}
	}

	err = s.queries.UpdateTransaccion(ctx, db.UpdateTransaccionParams{
		ID:             id,
		TipoOperacion:  tipoLimpio,
		Cantidad:       cantLimpia,
		PrecioUnitario: precioLimpio,
		Comision:       comisionLimpia,
		FechaOperacion: fechaNull,
		Notas:          notasNull,
		UsuarioID:      usuarioID,
	})
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22003" {
				return db.Transaccion{}, &ErrValidacion{"uno de los montos supera el límite numérico permitido"}
			}
		}
		return db.Transaccion{}, err
	}

	return s.ObtenerPorID(ctx, id, usuarioID)
}

func (s *TransaccionService) Eliminar(ctx context.Context, id, usuarioID uuid.UUID) error {
	// Verificamos si existe (y si pertenece al usuario) reutilizando ObtenerPorID
	_, err := s.ObtenerPorID(ctx, id, usuarioID)
	if err != nil {
		return err
	}

	// Procedemos a borrar
	err = s.queries.DeleteTransaccion(ctx, db.DeleteTransaccionParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		return err
	}
	return nil
}
