package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FranRamallo/ProgramacionWebTP/internal/services"
)

// DTO para crear un registro histórico (Fecha en formato "YYYY-MM-DD")
type CrearCotizacionHistoricaRequest struct {
	ActivoID       int64  `json:"activo_id"`
	Fecha          string `json:"fecha"`
	PrecioApertura string `json:"precio_apertura"`
	PrecioMax      string `json:"precio_max"`
	PrecioMin      string `json:"precio_min"`
	PrecioCierre   string `json:"precio_cierre"`
	Volumen        int64  `json:"volumen"`
}

// DTO para actualizar (sqlc actualiza solo precios y volumen por ID)
type ActualizarCotizacionHistoricaRequest struct {
	PrecioApertura string `json:"precio_apertura"`
	PrecioMax      string `json:"precio_max"`
	PrecioMin      string `json:"precio_min"`
	PrecioCierre   string `json:"precio_cierre"`
	Volumen        int64  `json:"volumen"`
}

type CotizacionHistoricaHandler struct {
	svc *services.CotizacionHistoricaService
}

func NewCotizacionHistoricaHandler(svc *services.CotizacionHistoricaService) *CotizacionHistoricaHandler {
	return &CotizacionHistoricaHandler{svc: svc}
}

func (h *CotizacionHistoricaHandler) CotizacionesHistoricasHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/cotizaciones-historicas")
	path = strings.TrimPrefix(path, "/")

	// Colección: GET /cotizaciones-historicas?activo_id=1 o POST /cotizaciones-historicas
	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.listar(w, r)
		case http.MethodPost:
			h.crear(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
		return
	}

	// Recurso específico: ID numérico (BIGSERIAL -> int64)
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "El ID de la cotización histórica debe ser un número", http.StatusBadRequest)
		return
	}
	id64 := int64(id)

	switch r.Method {
	case http.MethodGet:
		h.obtener(w, r, id64)
	case http.MethodPut:
		h.actualizar(w, r, id64)
	case http.MethodDelete:
		h.eliminar(w, r, id64)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// --- MÉTODOS PRIVADOS ---

func (h *CotizacionHistoricaHandler) crear(w http.ResponseWriter, r *http.Request) {
	var req CrearCotizacionHistoricaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Parseamos la fecha en formato YYYY-MM-DD
	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(req.Fecha))
	if err != nil {
		http.Error(w, "El formato de fecha debe ser YYYY-MM-DD (ej: 2026-03-31)", http.StatusBadRequest)
		return
	}

	cotizacion, err := h.svc.Crear(
		r.Context(),
		req.ActivoID,
		fecha,
		req.PrecioApertura,
		req.PrecioMax,
		req.PrecioMin,
		req.PrecioCierre,
		req.Volumen,
	)
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cotizacion)
}

func (h *CotizacionHistoricaHandler) listar(w http.ResponseWriter, r *http.Request) {
	// Extraemos activo_id de la URL: /cotizaciones-historicas?activo_id=1
	activoIDStr := r.URL.Query().Get("activo_id")
	if activoIDStr == "" {
		http.Error(w, "El parámetro activo_id es obligatorio en la URL", http.StatusBadRequest)
		return
	}

	activoID, err := strconv.Atoi(activoIDStr)
	if err != nil {
		http.Error(w, "El parámetro activo_id debe ser un número válido", http.StatusBadRequest)
		return
	}

	cotizaciones, err := h.svc.ListarPorActivo(r.Context(), int64(activoID))
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotizaciones)
}

func (h *CotizacionHistoricaHandler) obtener(w http.ResponseWriter, r *http.Request, id int64) {
	cotizacion, err := h.svc.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrCotizacionHistoricaNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotizacion)
}

func (h *CotizacionHistoricaHandler) actualizar(w http.ResponseWriter, r *http.Request, id int64) {
	var req ActualizarCotizacionHistoricaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	cotizacion, err := h.svc.Actualizar(
		r.Context(),
		id,
		req.PrecioApertura,
		req.PrecioMax,
		req.PrecioMin,
		req.PrecioCierre,
		req.Volumen,
	)
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errors.Is(err, services.ErrCotizacionHistoricaNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotizacion)
}

func (h *CotizacionHistoricaHandler) eliminar(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.svc.Eliminar(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrCotizacionHistoricaNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
