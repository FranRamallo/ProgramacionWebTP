package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/FranRamallo/ProgramacionWebTP/internal/services"
	"github.com/FranRamallo/ProgramacionWebTP/views"
)

// WebHandler es la capa de presentación server-side: recibe peticiones HTTP, le
// pide los datos a los servicios y responde con HTML renderizado por las vistas templ.
type WebHandler struct {
	transacciones *services.TransaccionService
	activos       *services.ActivoService
	// usuarioID es el usuario con el que opera la app. Todavía no hay login, asi
	// que main.go le pasa el usuario demo que crea db/seed/seed.sql.
	usuarioID uuid.UUID
}

func NewWebHandler(t *services.TransaccionService, a *services.ActivoService, usuarioID uuid.UUID) *WebHandler {
	return &WebHandler{transacciones: t, activos: a, usuarioID: usuarioID}
}

// Index atiende GET /: lista las transacciones y muestra el formulario vacío.
func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	formVacio := views.FormValues{Tipo: "compra", Comision: "0"}
	h.renderIndex(w, r, http.StatusOK, formVacio, "")
}

// Crear atiende POST /transacciones (Post/Redirect/Get):
//
//	OK       -> 303 See Other hacia "/" (el navegador hace un GET y ve la lista nueva).
//	Inválido -> vuelve a dibujar la página con 400, el mensaje y lo que el usuario escribio.
func (h *WebHandler) Crear(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario inválido", http.StatusBadRequest)
		return
	}

	form := views.FormValues{
		ActivoID: r.PostFormValue("activo_id"),
		Tipo:     r.PostFormValue("tipo_operacion"),
		Cantidad: r.PostFormValue("cantidad"),
		Precio:   r.PostFormValue("precio_unitario"),
		Comision: r.PostFormValue("comision"),
		Notas:    r.PostFormValue("notas"),
	}

	activoID, err := strconv.ParseInt(form.ActivoID, 10, 64)
	if err != nil {
		h.renderIndex(w, r, http.StatusBadRequest, form, "Elegí un activo de la lista")
		return
	}

	// fecha en nil: el servicio usa la fecha actual. Las notas vacías el servicio las guarda como NULL.
	_, err = h.transacciones.Crear(r.Context(), h.usuarioID, activoID,
		form.Tipo, form.Cantidad, form.Precio, form.Comision, nil, &form.Notas)
	if err != nil {
		var errVal *services.ErrValidacion
		if errors.As(err, &errVal) {
			h.renderIndex(w, r, http.StatusBadRequest, form, errVal.Mensaje)
			return
		}
		log.Printf("[WEB] error creando transacción: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Eliminar atiende POST /transacciones/eliminar.
func (h *WebHandler) Eliminar(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario inválido", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(r.PostFormValue("id"))
	if err != nil {
		http.Error(w, "El ID de la transacción debe ser un UUID válido", http.StatusBadRequest)
		return
	}

	err = h.transacciones.Eliminar(r.Context(), id, h.usuarioID)
	// Si ya no existía (por ejemplo, doble clic), igual volvemos a la lista.
	if err != nil && !errors.Is(err, services.ErrTransaccionNoEncontrada) {
		log.Printf("[WEB] error eliminando transacción: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// renderIndex junta los datos que necesita la página principal y la dibuja con el
// status indicado. Lo usan Index (200) y Crear cuando hay un error (400).
func (h *WebHandler) renderIndex(w http.ResponseWriter, r *http.Request, status int, form views.FormValues, mensajeError string) {
	activos, err := h.activos.Listar(r.Context())
	if err != nil {
		log.Printf("[WEB] error listando activos: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	transacciones, err := h.transacciones.ListarPorUsuario(r.Context(), h.usuarioID)
	if err != nil {
		log.Printf("[WEB] error listando transacciones: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	err = views.IndexPage(views.IndexData{
		Activos:       activos,
		Transacciones: transacciones,
		Form:          form,
		Error:         mensajeError,
	}).Render(r.Context(), w)
	if err != nil {
		// Ya empezamos a escribir la respuesta, así que no se puede cambiar el status: solo lo registramos.
		log.Printf("[WEB] error renderizando la vista: %v", err)
	}
}
