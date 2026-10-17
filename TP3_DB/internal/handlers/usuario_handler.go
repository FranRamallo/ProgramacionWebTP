package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/FranRamallo/ProgramacionWebTP/internal/services"
	"github.com/google/uuid"
)

// DTO para registrar un nuevo usuario
type CrearUsuarioRequest struct {
	NombreCompleto string `json:"nombre_completo"`
	Email          string `json:"email"`
	Password       string `json:"password"`
}

// DTO para editar el perfil (sin contraseña)
type ActualizarUsuarioRequest struct {
	NombreCompleto string `json:"nombre_completo"`
	Email          string `json:"email"`
}

type UsuarioHandler struct {
	svc *services.UsuarioService
}

func NewUsuarioHandler(svc *services.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{svc: svc}
}

func (h *UsuarioHandler) UsuariosHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/usuarios")
	path = strings.TrimPrefix(path, "/")

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

	// Intentamos parsear el ID de la URL como un UUID válido
	id, err := uuid.Parse(path) // convierte el uuid a string
	if err != nil {
		http.Error(w, "El ID del usuario debe ser un UUID válido", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.obtener(w, r, id)
	case http.MethodPut:
		h.actualizar(w, r, id)
	case http.MethodDelete:
		h.eliminar(w, r, id)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// --- MÉTODOS PRIVADOS ---

func (h *UsuarioHandler) crear(w http.ResponseWriter, r *http.Request) {
	var req CrearUsuarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	usuario, err := h.svc.Crear(r.Context(), req.NombreCompleto, req.Email, req.Password)
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
	json.NewEncoder(w).Encode(usuario)
}

func (h *UsuarioHandler) listar(w http.ResponseWriter, r *http.Request) {
	usuarios, err := h.svc.Listar(r.Context())
	if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(usuarios)
}

func (h *UsuarioHandler) obtener(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	usuario, err := h.svc.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrUsuarioNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(usuario)
}

func (h *UsuarioHandler) actualizar(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req ActualizarUsuarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	usuario, err := h.svc.Actualizar(r.Context(), id, req.NombreCompleto, req.Email)
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errors.Is(err, services.ErrUsuarioNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(usuario)
}

func (h *UsuarioHandler) eliminar(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	err := h.svc.Eliminar(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrUsuarioNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
