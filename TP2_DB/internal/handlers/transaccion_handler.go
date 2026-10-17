package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/FranRamallo/ProgramacionWebTP/internal/services"
	"github.com/google/uuid"
)

// DTO para crear una nueva transacción
type CrearTransaccionRequest struct {
	UsuarioID      uuid.UUID  `json:"usuario_id"`
	ActivoID       int64      `json:"activo_id"`
	TipoOperacion  string     `json:"tipo_operacion"`
	Cantidad       string     `json:"cantidad"`
	PrecioUnitario string     `json:"precio_unitario"`
	Comision       string     `json:"comision"`
	FechaOperacion *time.Time `json:"fecha_operacion"`
	Notas          *string    `json:"notas"`
}

// DTO para actualizar una transacción (no permite cambiar de activo)
type ActualizarTransaccionRequest struct {
	UsuarioID      uuid.UUID  `json:"usuario_id"`
	TipoOperacion  string     `json:"tipo_operacion"`
	Cantidad       string     `json:"cantidad"`
	PrecioUnitario string     `json:"precio_unitario"`
	Comision       string     `json:"comision"`
	FechaOperacion *time.Time `json:"fecha_operacion"`
	Notas          *string    `json:"notas"`
}

type TransaccionHandler struct {
	svc *services.TransaccionService
}

func NewTransaccionHandler(svc *services.TransaccionService) *TransaccionHandler {
	return &TransaccionHandler{svc: svc}
}

func (h *TransaccionHandler) TransaccionesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/transacciones")
	path = strings.TrimPrefix(path, "/")

	// Colección: GET /transacciones?usuario_id=... o POST /transacciones
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

	// Recurso específico: validamos el UUID de la transacción en la URL
	id, err := uuid.Parse(path)
	if err != nil {
		http.Error(w, "El ID de la transacción debe ser un UUID válido", http.StatusBadRequest)
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

func (h *TransaccionHandler) crear(w http.ResponseWriter, r *http.Request) {
	var req CrearTransaccionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	transaccion, err := h.svc.Crear(
		r.Context(),
		req.UsuarioID,
		req.ActivoID,
		req.TipoOperacion,
		req.Cantidad,
		req.PrecioUnitario,
		req.Comision,
		req.FechaOperacion,
		req.Notas,
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
	json.NewEncoder(w).Encode(transaccion)
}

func (h *TransaccionHandler) listar(w http.ResponseWriter, r *http.Request) {
	// Extraemos usuario_id de los query params: /transacciones?usuario_id=UUID
	usuarioIDStr := r.URL.Query().Get("usuario_id")
	if usuarioIDStr == "" {
		http.Error(w, "El parámetro usuario_id es obligatorio en la URL", http.StatusBadRequest)
		return
	}

	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		http.Error(w, "El usuario_id debe ser un UUID válido", http.StatusBadRequest)
		return
	}

	transacciones, err := h.svc.ListarPorUsuario(r.Context(), usuarioID)
	if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transacciones)
}

func (h *TransaccionHandler) obtener(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	// Extraemos usuario_id de la URL: /transacciones/{id}?usuario_id=UUID
	usuarioIDStr := r.URL.Query().Get("usuario_id")
	if usuarioIDStr == "" {
		http.Error(w, "El parámetro usuario_id es obligatorio en la URL", http.StatusBadRequest)
		return
	}

	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		http.Error(w, "El usuario_id debe ser un UUID válido", http.StatusBadRequest)
		return
	}

	transaccion, err := h.svc.ObtenerPorID(r.Context(), id, usuarioID)
	if err != nil {
		if errors.Is(err, services.ErrTransaccionNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transaccion)
}

func (h *TransaccionHandler) actualizar(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req ActualizarTransaccionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	transaccion, err := h.svc.Actualizar(
		r.Context(),
		id,
		req.UsuarioID,
		req.TipoOperacion,
		req.Cantidad,
		req.PrecioUnitario,
		req.Comision,
		req.FechaOperacion,
		req.Notas,
	)
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errors.Is(err, services.ErrTransaccionNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transaccion)
}

func (h *TransaccionHandler) eliminar(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	// Extraemos usuario_id de la URL: DELETE /transacciones/{id}?usuario_id=UUID
	usuarioIDStr := r.URL.Query().Get("usuario_id")
	if usuarioIDStr == "" {
		http.Error(w, "El parámetro usuario_id es obligatorio en la URL", http.StatusBadRequest)
		return
	}

	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		http.Error(w, "El usuario_id debe ser un UUID válido", http.StatusBadRequest)
		return
	}

	err = h.svc.Eliminar(r.Context(), id, usuarioID)
	if err != nil {
		if errors.Is(err, services.ErrTransaccionNoEncontrada) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
