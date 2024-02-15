package structs

//Parte de la partición lógica, será una lista enlazada, donde conectará con los siguientes EBR.

type EBR struct {
	Part_mount byte
	Part_fit   byte
	Part_start int64
	Part_size  int64
	Part_next  int64
	Part_name  [16]byte
}

func newEBR() EBR {
	var eb EBR
	eb.Part_mount = '0'
	eb.Part_size = 0
	eb.Part_next = -1
	return eb
}
