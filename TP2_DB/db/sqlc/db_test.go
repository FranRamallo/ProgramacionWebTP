package db

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ========================================================================
// 1. HELPERS
// ========================================================================

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:7777/apirest?sslmode=disable")
	if err != nil {
		t.Fatalf("No se pudo abrir la conexión: %v", err)
	}
	if err := conn.Ping(); err != nil {
		t.Fatalf("No se pudo conectar a la base: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// ========================================================================
// 2. TESTS INDEPENDIENTES
// ========================================================================

func TestSector(t *testing.T) {
	db := openDB(t)
	queries := New(db)
	ctx := context.Background()
	var createdSector Sector

	t.Run("Create", func(t *testing.T) {
		var err error
		createdSector, err = queries.CreateSector(ctx, "Sector "+randomString(6))
		if err != nil {
			t.Fatalf("CreateSector falló: %v", err)
		}
	})

	t.Cleanup(func() { _ = queries.DeleteSector(ctx, createdSector.ID) })

	t.Run("Read", func(t *testing.T) {
		sector, err := queries.GetSector(ctx, createdSector.ID)
		if err != nil {
			t.Fatalf("GetSector falló: %v", err)
		}
		if createdSector.Nombre != sector.Nombre {
			t.Errorf("Nombres no coinciden")
		}
	})

	t.Run("Update", func(t *testing.T) {
		nuevoNombre := "Actualizado " + randomString(4)
		err := queries.UpdateSector(ctx, UpdateSectorParams{
			ID:     createdSector.ID,
			Nombre: nuevoNombre,
		})
		if err != nil {
			t.Fatalf("UpdateSector falló: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		sectores, err := queries.ListSectores(ctx)
		if err != nil || len(sectores) == 0 {
			t.Fatalf("ListSectores falló o devolvió 0 elementos")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := queries.DeleteSector(ctx, createdSector.ID)
		if err != nil {
			t.Fatalf("DeleteSector falló: %v", err)
		}
		_, err = queries.GetSector(ctx, createdSector.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Esperaba sql.ErrNoRows tras borrar, obtuve: %v", err)
		}
	})
}

func TestUsuario(t *testing.T) {
	db := openDB(t)
	queries := New(db)
	ctx := context.Background()
	var createdUser Usuario

	t.Run("Create", func(t *testing.T) {
		var err error
		arg := CreateUsuarioParams{
			NombreCompleto: "Usuario Test",
			Email:          randomString(8) + "@test.com",
			PasswordHash:   "hash123",
		}
		createdUser, err = queries.CreateUsuario(ctx, arg)
		if err != nil {
			t.Fatalf("CreateUsuario falló: %v", err)
		}
	})

	t.Cleanup(func() { _ = queries.DeleteUsuario(ctx, createdUser.ID) })

	t.Run("Read", func(t *testing.T) {
		user, err := queries.GetUsuario(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("GetUsuario falló: %v", err)
		}
		if createdUser.Email != user.Email {
			t.Errorf("Emails no coinciden")
		}
	})

	t.Run("Update", func(t *testing.T) {
		arg := UpdateUsuarioParams{
			ID:             createdUser.ID,
			NombreCompleto: "Usuario Modificado",
			Email:          createdUser.Email, // Mantenemos email para no violar UNIQUE
		}
		if err := queries.UpdateUsuario(ctx, arg); err != nil {
			t.Fatalf("UpdateUsuario falló: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		usuarios, err := queries.ListUsuarios(ctx)
		if err != nil || len(usuarios) == 0 {
			t.Fatalf("ListUsuarios falló")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := queries.DeleteUsuario(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("DeleteUsuario falló: %v", err)
		}
		_, err = queries.GetUsuario(ctx, createdUser.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Esperaba ErrNoRows tras borrar")
		}
	})
}

// ========================================================================
// 3. TESTS CON DEPENDENCIAS
// ========================================================================

func TestActivo(t *testing.T) {
	db := openDB(t)
	queries := New(db)
	ctx := context.Background()

	sector, err := queries.CreateSector(ctx, "Sector Activos "+randomString(4))
	if err != nil {
		t.Fatalf("setup: CreateSector falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteSector(ctx, sector.ID) })

	var createdActivo Activo

	t.Run("Create", func(t *testing.T) {
		var err error
		arg := CreateActivoParams{
			Ticker:        randomString(4),
			NombreEmpresa: "Empresa SA",
			SectorID:      sql.NullInt32{Int32: sector.ID, Valid: true},
			Bolsa:         "NASDAQ",
			TipoActivo:    "Accion",
		}
		createdActivo, err = queries.CreateActivo(ctx, arg)
		if err != nil {
			t.Fatalf("CreateActivo falló: %v", err)
		}
	})

	t.Cleanup(func() { _ = queries.DeleteActivo(ctx, createdActivo.ID) })

	t.Run("Read", func(t *testing.T) {
		activo, err := queries.GetActivo(ctx, createdActivo.ID)
		if err != nil {
			t.Fatalf("GetActivo falló: %v", err)
		}
		if activo.Ticker != createdActivo.Ticker {
			t.Errorf("Tickers no coinciden")
		}
	})

	t.Run("Update", func(t *testing.T) {
		arg := UpdateActivoParams{
			ID:            createdActivo.ID,
			Ticker:        createdActivo.Ticker,
			NombreEmpresa: "Empresa Modificada",
			SectorID:      createdActivo.SectorID,
			Bolsa:         "NYSE",
			TipoActivo:    "Accion",
		}
		if err := queries.UpdateActivo(ctx, arg); err != nil {
			t.Fatalf("UpdateActivo falló: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		activos, err := queries.ListActivos(ctx)
		if err != nil || len(activos) == 0 {
			t.Fatalf("ListActivos falló")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := queries.DeleteActivo(ctx, createdActivo.ID)
		if err != nil {
			t.Fatalf("DeleteActivo falló: %v", err)
		}
		_, err = queries.GetActivo(ctx, createdActivo.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Esperaba ErrNoRows")
		}
	})
}

func TestTransaccion(t *testing.T) {
	db := openDB(t)
	queries := New(db)
	ctx := context.Background()

	usuario, err := queries.CreateUsuario(ctx, CreateUsuarioParams{
		NombreCompleto: "Inversor Test",
		Email:          randomString(8) + "@test.com",
		PasswordHash:   "1234",
	})
	if err != nil {
		t.Fatalf("setup: CreateUsuario falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteUsuario(ctx, usuario.ID) })

	sector, err := queries.CreateSector(ctx, "Sector Fin "+randomString(4))
	if err != nil {
		t.Fatalf("setup: CreateSector falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteSector(ctx, sector.ID) })

	activo, err := queries.CreateActivo(ctx, CreateActivoParams{
		Ticker:        randomString(4),
		NombreEmpresa: "Empresa SA",
		SectorID:      sql.NullInt32{Int32: sector.ID, Valid: true},
		Bolsa:         "NYSE",
		TipoActivo:    "Accion",
	})
	if err != nil {
		t.Fatalf("setup: CreateActivo falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteActivo(ctx, activo.ID) })

	var tx Transaccion

	t.Run("Create", func(t *testing.T) {
		var err error
		arg := CreateTransaccionParams{
			UsuarioID:      usuario.ID,
			ActivoID:       activo.ID,
			TipoOperacion:  "compra",
			Cantidad:       "10.5",
			PrecioUnitario: "150.0",
			Comision:       "1.5",
			FechaOperacion: sql.NullTime{Time: time.Now(), Valid: true},
			Notas:          sql.NullString{String: "Inicial", Valid: true},
		}
		tx, err = queries.CreateTransaccion(ctx, arg)
		if err != nil {
			t.Fatalf("CreateTransaccion falló: %v", err)
		}
	})

	t.Cleanup(func() {
		_ = queries.DeleteTransaccion(ctx, DeleteTransaccionParams{ID: tx.ID, UsuarioID: usuario.ID})
	})

	t.Run("Read", func(t *testing.T) {
		res, err := queries.GetTransaccion(ctx, GetTransaccionParams{ID: tx.ID, UsuarioID: usuario.ID})
		if err != nil {
			t.Fatalf("GetTransaccion falló: %v", err)
		}
		if res.Cantidad != tx.Cantidad {
			t.Errorf("Cantidades no coinciden")
		}
	})

	t.Run("Update", func(t *testing.T) {
		arg := UpdateTransaccionParams{
			ID:             tx.ID,
			UsuarioID:      tx.UsuarioID,
			TipoOperacion:  "venta",
			Cantidad:       "5.0",
			PrecioUnitario: "160.0",
			Comision:       "1.5",
			FechaOperacion: tx.FechaOperacion,
			Notas:          sql.NullString{String: "Venta parcial", Valid: true},
		}
		if err := queries.UpdateTransaccion(ctx, arg); err != nil {
			t.Fatalf("UpdateTransaccion falló: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		txs, err := queries.ListTransaccionesPorUsuario(ctx, usuario.ID)
		if err != nil || len(txs) == 0 {
			t.Fatalf("ListTransacciones falló o vino vacío")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := queries.DeleteTransaccion(ctx, DeleteTransaccionParams{ID: tx.ID, UsuarioID: usuario.ID})
		if err != nil {
			t.Fatalf("DeleteTransaccion falló: %v", err)
		}
		_, err = queries.GetTransaccion(ctx, GetTransaccionParams{ID: tx.ID, UsuarioID: usuario.ID})
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Esperaba ErrNoRows tras borrar")
		}
	})
}

// ========================================================================
// 4. TESTS DE DATOS FINANCIEROS
// ========================================================================

func TestCotizacionActual(t *testing.T) {
	db := openDB(t)
	queries := New(db)
	ctx := context.Background()

	sector, err := queries.CreateSector(ctx, "Sector Cotiz "+randomString(4))
	if err != nil {
		t.Fatalf("setup: CreateSector falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteSector(ctx, sector.ID) })

	activo, err := queries.CreateActivo(ctx, CreateActivoParams{
		Ticker:        randomString(4),
		NombreEmpresa: "Empresa SA",
		SectorID:      sql.NullInt32{Int32: sector.ID, Valid: true},
		Bolsa:         "NASDAQ",
		TipoActivo:    "Accion",
	})
	if err != nil {
		t.Fatalf("setup: CreateActivo falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteActivo(ctx, activo.ID) })

	t.Run("Upsert - Create", func(t *testing.T) {
		arg := UpsertCotizacionActualParams{
			ActivoID:        activo.ID,
			Precio:          "150.50",
			CierreAnterior:  "149.00",
			VariacionDiaPct: "1.01",
			ActualizadoEn:   time.Now(),
		}
		if _, err := queries.UpsertCotizacionActual(ctx, arg); err != nil {
			t.Fatalf("Upsert (Create) falló: %v", err)
		}
	})

	t.Cleanup(func() { _ = queries.DeleteCotizacionActual(ctx, activo.ID) })

	t.Run("Read", func(t *testing.T) {
		if _, err := queries.GetCotizacionActual(ctx, activo.ID); err != nil {
			t.Fatalf("GetCotizacionActual falló: %v", err)
		}
	})

	t.Run("Upsert - Update", func(t *testing.T) {
		arg := UpsertCotizacionActualParams{
			ActivoID:        activo.ID,
			Precio:          "155.00",
			CierreAnterior:  "150.50",
			VariacionDiaPct: "2.99",
			ActualizadoEn:   time.Now(),
		}
		if _, err := queries.UpsertCotizacionActual(ctx, arg); err != nil {
			t.Fatalf("Upsert (Update) falló: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		list, err := queries.ListCotizacionesActuales(ctx)
		if err != nil || len(list) == 0 {
			t.Fatalf("ListCotizacionesActuales falló")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := queries.DeleteCotizacionActual(ctx, activo.ID); err != nil {
			t.Fatalf("Delete falló: %v", err)
		}
		if _, err := queries.GetCotizacionActual(ctx, activo.ID); !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Esperaba ErrNoRows")
		}
	})
}

func TestCotizacionHistorica(t *testing.T) {
	db := openDB(t)
	queries := New(db)
	ctx := context.Background()

	sector, err := queries.CreateSector(ctx, "Sector Hist "+randomString(4))
	if err != nil {
		t.Fatalf("setup: CreateSector falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteSector(ctx, sector.ID) })

	activo, err := queries.CreateActivo(ctx, CreateActivoParams{
		Ticker:        randomString(4),
		NombreEmpresa: "Empresa SA",
		SectorID:      sql.NullInt32{Int32: sector.ID, Valid: true},
		Bolsa:         "NYSE",
		TipoActivo:    "Accion",
	})
	if err != nil {
		t.Fatalf("setup: CreateActivo falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteActivo(ctx, activo.ID) })

	var historica CotizacionHistorica

	t.Run("Create", func(t *testing.T) {
		var err error
		arg := CreateCotizacionHistoricaParams{
			ActivoID:       activo.ID,
			Fecha:          time.Now(),
			PrecioApertura: "100.00",
			PrecioMax:      "105.00",
			PrecioMin:      "99.00",
			PrecioCierre:   "104.50",
			Volumen:        1500000,
		}
		historica, err = queries.CreateCotizacionHistorica(ctx, arg)
		if err != nil {
			t.Fatalf("Create falló: %v", err)
		}
	})

	t.Cleanup(func() { _ = queries.DeleteCotizacionHistorica(ctx, historica.ID) })

	t.Run("Read", func(t *testing.T) {
		if _, err := queries.GetCotizacionHistorica(ctx, historica.ID); err != nil {
			t.Fatalf("Get falló: %v", err)
		}
	})

	t.Run("Update", func(t *testing.T) {
		arg := UpdateCotizacionHistoricaParams{
			ID:             historica.ID,
			PrecioApertura: "100.00",
			PrecioMax:      "106.00",
			PrecioMin:      "99.00",
			PrecioCierre:   "105.50",
			Volumen:        2000000,
		}
		if err := queries.UpdateCotizacionHistorica(ctx, arg); err != nil {
			t.Fatalf("Update falló: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		list, err := queries.ListCotizacionesHistoricas(ctx, activo.ID)
		if err != nil || len(list) == 0 {
			t.Fatalf("List falló")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := queries.DeleteCotizacionHistorica(ctx, historica.ID); err != nil {
			t.Fatalf("Delete falló: %v", err)
		}
		if _, err := queries.GetCotizacionHistorica(ctx, historica.ID); !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Esperaba ErrNoRows")
		}
	})
}

func TestMetricaFundamental(t *testing.T) {
	db := openDB(t)
	queries := New(db)
	ctx := context.Background()

	sector, err := queries.CreateSector(ctx, "Sector Metrica "+randomString(4))
	if err != nil {
		t.Fatalf("setup: CreateSector falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteSector(ctx, sector.ID) })

	activo, err := queries.CreateActivo(ctx, CreateActivoParams{
		Ticker:        randomString(4),
		NombreEmpresa: "Empresa SA",
		SectorID:      sql.NullInt32{Int32: sector.ID, Valid: true},
		Bolsa:         "BCBA",
		TipoActivo:    "CEDEAR",
	})
	if err != nil {
		t.Fatalf("setup: CreateActivo falló: %v", err)
	}
	t.Cleanup(func() { _ = queries.DeleteActivo(ctx, activo.ID) })

	var metrica MetricaFundamental

	t.Run("Create", func(t *testing.T) {
		var err error
		arg := CreateMetricaFundamentalParams{
			ActivoID:      activo.ID,
			FechaReporte:  time.Now(),
			PeRatio:       sql.NullString{String: "15.5", Valid: true},
			DividendYield: sql.NullString{String: "2.1", Valid: true},
			DeudaCapital:  sql.NullString{String: "0.8", Valid: true},
			Roe:           sql.NullString{String: "12.5", Valid: true},
			Eps:           sql.NullString{String: "3.45", Valid: true},
		}
		metrica, err = queries.CreateMetricaFundamental(ctx, arg)
		if err != nil {
			t.Fatalf("Create falló: %v", err)
		}
	})

	t.Cleanup(func() { _ = queries.DeleteMetricaFundamental(ctx, metrica.ID) })

	t.Run("Read", func(t *testing.T) {
		if _, err := queries.GetMetricaFundamental(ctx, metrica.ID); err != nil {
			t.Fatalf("Get falló: %v", err)
		}
	})

	t.Run("Update", func(t *testing.T) {
		arg := UpdateMetricaFundamentalParams{
			ID:            metrica.ID,
			PeRatio:       sql.NullString{String: "16.0", Valid: true},
			DividendYield: sql.NullString{String: "2.1", Valid: true},
			DeudaCapital:  sql.NullString{String: "0.8", Valid: true},
			Roe:           sql.NullString{String: "12.5", Valid: true},
			Eps:           sql.NullString{String: "3.50", Valid: true},
		}
		if err := queries.UpdateMetricaFundamental(ctx, arg); err != nil {
			t.Fatalf("Update falló: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		list, err := queries.ListMetricasFundamentales(ctx, activo.ID)
		if err != nil || len(list) == 0 {
			t.Fatalf("List falló")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := queries.DeleteMetricaFundamental(ctx, metrica.ID); err != nil {
			t.Fatalf("Delete falló: %v", err)
		}
		if _, err := queries.GetMetricaFundamental(ctx, metrica.ID); !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Esperaba ErrNoRows")
		}
	})
}
