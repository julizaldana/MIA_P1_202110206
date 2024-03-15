package Structs

//El objeto Journaling servirá para tener un registro de las operaciones realizadas para el sistema EXT3.

type Journaling struct {
	J_operacion [16]byte
	J_path      [32]byte
	J_content   [64]byte
	J_fecha     [16]byte
}

func NewJournaling() Journaling {
	var journal Journaling
	journal.J_operacion = [16]byte{}
	journal.J_path = [32]byte{}
	journal.J_content = [64]byte{}
	journal.J_fecha = [16]byte{}
	return journal
}
