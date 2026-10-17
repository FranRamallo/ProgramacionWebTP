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

// DTO para POST /cotizaciones-actuales
type CrearCotizacionActualRequest struct {
	ActivoID        int64      `json:"activo_id"`
	Precio          string     `json:"precio"`
	CierreAnterior  string     `json:"cierre_anterior"`
	VariacionDiaPct string     `json:"variacion_dia_pct"`
	ActualizadoEn   *time.Time `json:"actualizado_en"`
}

// DTO para PUT /cotizaciones-actuales/{activo_id}
type ActualizarCotizacionActualRequest struct {
	Precio          string     `json:"precio"`
	CierreAnterior  string     `json:"cierre_anterior"`
	VariacionDiaPct string     `json:"variacion_dia_pct"`
	ActualizadoEn   *time.Time `json:"actualizado_en"`
}

type CotizacionActualHandler struct {
	svc *services.CotizacionActualService
}

func NewCotizacionActualHandler(svc *services.CotizacionActualService) *CotizacionActualHandler {
	return &CotizacionActualHandler{svc: svc}
}

func (h *CotizacionActualHandler) CotizacionesActualesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/cotizaciones-actuales")
	path = strings.TrimPrefix(path, "/")

	// Colección: GET todas o POST (Upsert con activo_id en el JSON)
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

	// Recurso específico: el ID en la URL es el activo_id (BIGINT -> int64)
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "El ID del activo debe ser un número", http.StatusBadRequest)
		return
	}
	activoID := int64(id)

	switch r.Method {
	case http.MethodGet:
		h.obtener(w, r, activoID)
	case http.MethodPut:
		h.actualizar(w, r, activoID)
	case http.MethodDelete:
		h.eliminar(w, r, activoID)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// --- MÉTODOS PRIVADOS ---

func (h *CotizacionActualHandler) crear(w http.ResponseWriter, r *http.Request) {
	var req CrearCotizacionActualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	cotizacion, err := h.svc.Guardar(
		r.Context(),
		req.ActivoID,
		req.Precio,
		req.CierreAnterior,
		req.VariacionDiaPct,
		req.ActualizadoEn,
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

func (h *CotizacionActualHandler) listar(w http.ResponseWriter, r *http.Request) {
	cotizaciones, err := h.svc.Listar(r.Context())
	if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotizaciones)
}

func (h *CotizacionActualHandler) obtener(w http.ResponseWriter, r *http.Request, activoID int64) {
	cotizacion, err := h.svc.ObtenerPorActivoID(r.Context(), activoID)
	if err != nil {
		if errors.Is(err, services.ErrCotizacionActualNoEncontrada) {
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

func (h *CotizacionActualHandler) actualizar(w http.ResponseWriter, r *http.Request, activoID int64) {
	var req ActualizarCotizacionActualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	cotizacion, err := h.svc.Guardar(
		r.Context(),
		activoID,
		req.Precio,
		req.CierreAnterior,
		req.VariacionDiaPct,
		req.ActualizadoEn,
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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotizacion)
}

func (h *CotizacionActualHandler) eliminar(w http.ResponseWriter, r *http.Request, activoID int64) {
	err := h.svc.Eliminar(r.Context(), activoID)
	if err != nil {
		if errors.Is(err, services.ErrCotizacionActualNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
