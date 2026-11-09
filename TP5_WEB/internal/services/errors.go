package services

// ErrValidacion representa un error por reglas de negocio no cumplidas.
// Se comparte entre todos los servicios (Sector, Activo, Usuario, etc).
type ErrValidacion struct {
	Mensaje string
}

func (e *ErrValidacion) Error() string {
	return e.Mensaje
}
