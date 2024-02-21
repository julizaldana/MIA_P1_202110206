package Comandos

import "strings"

//REP -> SIRVE PARA REALIZAR LOS REPORTES ESPECIFICOS - CON GRAPHVIZ

var contadorBloques int
var contadorArchivos int
var bloquesUsados []int64

func ValidarDatosREP(context []string) {
	contadorBloques = 0
	contadorArchivos = 0
	bloquesUsados = []int64{}
	name := ""
	path := ""
	id := ""
	ruta := ""
	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "path") {
			path = strings.ReplaceAll(tk[1], "\"", "")
		} else if Comparar(tk[0], "name") {
			name = tk[1]
		} else if Comparar(tk[0], "id") {
			id = tk[1]
		} else if Comparar(tk[0], "ruta") {
			ruta = tk[1]
		}
	}
	if name == "" || path == "" || id == "" {
		Error("REP", "Se esperaban parámetros obligatorios")
		return
	}
	if Comparar(name, "DISK") {
		//dks(path, id)
	} else if Comparar(name, "TREE") {
		//tree(path, id)
	} else if Comparar(name, "FILE") {
		if ruta == "" {
			Error("REP", "Se espera el parámetro ruta.")
		}
		//fileR(path, id, ruta)
	} else {
		Error("REP", name+", no es un reporte válido.")
		return
	}
}
