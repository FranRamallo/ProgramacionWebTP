package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/FranRamallo/ProgramacionWebTP/internal/services"
)

// ActivoRequest es el DTO para recibir el JSON.
// Usamos un puntero (*int32) para SectorID para que soporte valores null desde el cliente.
type ActivoRequest struct {
	Ticker        string `json:"ticker"`
	NombreEmpresa string `json:"nombre_empresa"`
	SectorID      *int32 `json:"sector_id"`
	Bolsa         string `json:"bolsa"`
	TipoActivo    string `json:"tipo_activo"`
}

type ActivoHandler struct {
	svc *services.ActivoService
}

func NewActivoHandler(svc *services.ActivoService) *ActivoHandler {
	return &ActivoHandler{svc: svc}
}

// ActivosHandler unifica todas las operaciones CRUD en un solo punto de entrada.
func (h *ActivoHandler) ActivosHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/activos")
	path = strings.TrimPrefix(path, "/")

	// Colección: GET todos o POST
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

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "El ID del activo debe ser un número", http.StatusBadRequest)
		return
	}
	id64 := int64(id) // Casteo explícito a 64 bits para que sqlc lo acepte

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

// --- MÉTODOS PRIVADOS DE OPERACIÓN ---

func (h *ActivoHandler) crear(w http.ResponseWriter, r *http.Request) {
	var req ActivoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	activo, err := h.svc.Crear(r.Context(), req.Ticker, req.NombreEmpresa, req.SectorID, req.Bolsa, req.TipoActivo)
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
	json.NewEncoder(w).Encode(activo)
}

func (h *ActivoHandler) listar(w http.ResponseWriter, r *http.Request) {
	activos, err := h.svc.Listar(r.Context())
	if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(activos)
}

func (h *ActivoHandler) obtener(w http.ResponseWriter, r *http.Request, id int64) {
	activo, err := h.svc.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrActivoNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(activo)
}

func (h *ActivoHandler) actualizar(w http.ResponseWriter, r *http.Request, id int64) {
	var req ActivoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	activo, err := h.svc.Actualizar(r.Context(), id, req.Ticker, req.NombreEmpresa, req.SectorID, req.Bolsa, req.TipoActivo)
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errors.Is(err, services.ErrActivoNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(activo)
}

func (h *ActivoHandler) eliminar(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.svc.Eliminar(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrActivoNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
