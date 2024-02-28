package Comandos

import "strings"

//SE DEBE DE CREAR UN SISTEMA DE ARCHIVOS EXT2. EN LA RAIZ, SE TENDRÁ UN ARCHIVO TXT, CON LOS USUARIOS, GRUPOS, INFORMACION DE LOS MISMOS.

func ValidarDatosMKFS(context []string) {
	typ := ""
	id := ""
	fs := ""
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "id") {
			id = comando[1]
		} else if Comparar(comando[0], "type") {
			typ = comando[1]
		} else if Comparar(comando[0], "fs") {
			fs = comando[1]
		}

	}
	if id == "" {
		Error("MKFS", "El comando MKFS requiere un id para poder ser ejecutado.")
		return
	} else {
		crearFileSystem()
	}

}

func crearFileSystem() {

}
