package Comandos

import "strings"

func ValidarDatosMKFILE(context []string) {
	path := ""
	r := ""
	size := ""
	cont := ""
	//fs := "" //Verificar si es ext2 o ext3
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "path") {
			path = comando[1]
		} else if Comparar(comando[0], "r") {
			r = comando[1]
		} else if Comparar(comando[0], "size") {
			size = comando[1]
		} else if Comparar(comando[0], "cont") {
			cont = comando[1]
		}
	}
	if path == "" {
		Error("MKDIR", "El comando MKDIR requiere el path o ruta para poder crear un directorio.")
		return
	} else {
		crearArchivo(path, r, size, cont)
	}

}

func crearArchivo(path string, r string, size string, cont string) {

}
