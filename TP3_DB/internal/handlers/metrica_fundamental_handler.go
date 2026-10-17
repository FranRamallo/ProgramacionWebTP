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

// DTO para crear una métrica fundamental (FechaReporte en formato "YYYY-MM-DD")
type CrearMetricaFundamentalRequest struct {
	ActivoID      int64   `json:"activo_id"`
	FechaReporte  string  `json:"fecha_reporte"`
	PeRatio       *string `json:"pe_ratio"`
	DividendYield *string `json:"dividend_yield"`
	DeudaCapital  *string `json:"deuda_capital"`
	Roe           *string `json:"roe"`
	Eps           *string `json:"eps"`
}

// DTO para actualizar (sqlc solo actualiza los indicadores, no el activo_id ni la fecha_reporte)
type ActualizarMetricaFundamentalRequest struct {
	PeRatio       *string `json:"pe_ratio"`
	DividendYield *string `json:"dividend_yield"`
	DeudaCapital  *string `json:"deuda_capital"`
	Roe           *string `json:"roe"`
	Eps           *string `json:"eps"`
}

type MetricaFundamentalHandler struct {
	svc *services.MetricaFundamentalService
}

func NewMetricaFundamentalHandler(svc *services.MetricaFundamentalService) *MetricaFundamentalHandler {
	return &MetricaFundamentalHandler{svc: svc}
}

func (h *MetricaFundamentalHandler) MetricasFundamentalesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/metricas-fundamentales")
	path = strings.TrimPrefix(path, "/")

	// Colección: GET /metricas-fundamentales?activo_id=1 o POST /metricas-fundamentales
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
		http.Error(w, "El ID de la métrica debe ser un número", http.StatusBadRequest)
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

func (h *MetricaFundamentalHandler) crear(w http.ResponseWriter, r *http.Request) {
	var req CrearMetricaFundamentalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Parseamos la fecha en formato YYYY-MM-DD (el layout estándar de Go es "2006-01-02")
	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(req.FechaReporte))
	if err != nil {
		http.Error(w, "El formato de fecha_reporte debe ser YYYY-MM-DD (ej: 2026-03-31)", http.StatusBadRequest)
		return
	}

	metrica, err := h.svc.Crear(
		r.Context(),
		req.ActivoID,
		fecha,
		req.PeRatio,
		req.DividendYield,
		req.DeudaCapital,
		req.Roe,
		req.Eps,
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
	json.NewEncoder(w).Encode(metrica)
}

func (h *MetricaFundamentalHandler) listar(w http.ResponseWriter, r *http.Request) {
	// Extraemos activo_id de la URL: /metricas-fundamentales?activo_id=1
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

	metricas, err := h.svc.ListarPorActivo(r.Context(), int64(activoID))
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
	json.NewEncoder(w).Encode(metricas)
}

func (h *MetricaFundamentalHandler) obtener(w http.ResponseWriter, r *http.Request, id int64) {
	metrica, err := h.svc.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrMetricaFundamentalNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrica)
}

func (h *MetricaFundamentalHandler) actualizar(w http.ResponseWriter, r *http.Request, id int64) {
	var req ActualizarMetricaFundamentalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	metrica, err := h.svc.Actualizar(
		r.Context(),
		id,
		req.PeRatio,
		req.DividendYield,
		req.DeudaCapital,
		req.Roe,
		req.Eps,
	)
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errors.Is(err, services.ErrMetricaFundamentalNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrica)
}

func (h *MetricaFundamentalHandler) eliminar(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.svc.Eliminar(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrMetricaFundamentalNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
