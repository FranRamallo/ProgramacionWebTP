package views

import (
	"database/sql"
	"strconv"
	"strings"

	db "github.com/FranRamallo/ProgramacionWebTP/db/sqlc"
)

// FormValues son los valores del formulario de alta tal como los escribió el
// usuario. Son strings a propósito: así podemos volver a mostrarlos tal cual si
// el servidor rechaza la operación (por ejemplo, cantidad = 0).
type FormValues struct {
	ActivoID string
	Tipo     string
	Cantidad string
	Precio   string
	Comision string
	Notas    string
}

// IndexData agrupa todo lo que necesita IndexPage para dibujarse.
type IndexData struct {
	Activos       []db.Activo
	Transacciones []db.Transaccion
	Form          FormValues
	Error         string
}

// idStr convierte un ID numérico a texto (para value="..." de los <option>).
func idStr(id int64) string {
	return strconv.FormatInt(id, 10)
}

// tickerDe busca el ticker del activo con ese ID. Si no lo encuentra devuelve "#ID".
func tickerDe(activos []db.Activo, id int64) string {
	for _, a := range activos {
		if a.ID == id {
			return a.Ticker
		}
	}
	return "#" + idStr(id)
}

// formatDecimal le saca los ceros sobrantes a un DECIMAL que viene de Postgres
func formatDecimal(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// formatPrecio es formatDecimal con el símbolo de dólar adelante.
func formatPrecio(s string) string {
	return "$" + formatDecimal(s)
}

// formatFecha muestra la fecha de la operación en hora local; "—" si es NULL.
func formatFecha(t sql.NullTime) string {
	if !t.Valid {
		return "—"
	}
	return t.Time.Local().Format("02/01/2006 15:04")
}
