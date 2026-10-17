package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/FranRamallo/ProgramacionWebTP/internal/services"
)

type SectorRequest struct {
	Nombre string `json:"nombre"`
}

type SectorHandler struct {
	svc *services.SectorService
}

func NewSectorHandler(svc *services.SectorService) *SectorHandler {
	return &SectorHandler{svc: svc}
}

// SectoresHandler unifica todas las operaciones CRUD en un solo punto de entrada.
func (h *SectorHandler) SectoresHandler(w http.ResponseWriter, r *http.Request) {
	// Limpiamos la URL base para quedarnos con el ID (si existe)
	path := strings.TrimPrefix(r.URL.Path, "/sectores")
	path = strings.TrimPrefix(path, "/")

	// Si no hay ID en la URL, manejamos la colección (GET todos o POST)
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

	// Si llegamos acá, es porque hay un ID en la URL (/sectores/123)
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "El ID del sector debe ser un número", http.StatusBadRequest)
		return
	}

	id32 := int32(id)

	// Manejamos las operaciones sobre un recurso específico
	switch r.Method {
	case http.MethodGet:
		h.obtener(w, r, id32)
	case http.MethodPut:
		h.actualizar(w, r, id32)
	case http.MethodDelete:
		h.eliminar(w, r, id32)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// --- MÉTODOS PRIVADOS DE OPERACIÓN ---

func (h *SectorHandler) crear(w http.ResponseWriter, r *http.Request) {
	var req SectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	sector, err := h.svc.Crear(r.Context(), req.Nombre)
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
	json.NewEncoder(w).Encode(sector)
}

func (h *SectorHandler) listar(w http.ResponseWriter, r *http.Request) {
	sectores, err := h.svc.Listar(r.Context())
	if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sectores)
}

func (h *SectorHandler) obtener(w http.ResponseWriter, r *http.Request, id int32) {
	sector, err := h.svc.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSectorNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sector)
}

func (h *SectorHandler) actualizar(w http.ResponseWriter, r *http.Request, id int32) {
	var req SectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	sector, err := h.svc.Actualizar(r.Context(), id, req.Nombre)
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errors.Is(err, services.ErrSectorNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sector)
}

func (h *SectorHandler) eliminar(w http.ResponseWriter, r *http.Request, id int32) {
	err := h.svc.Eliminar(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrSectorNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
