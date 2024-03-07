package main

import (
	"MIA_P1_202110206/Comandos"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

//PROYECTO 1 - MANEJO E IMPLEMENTACIÓN DE ARCHIVOS
//NOMBRE: JULIO ALEJANDRO ZALDAÑA RÍOS 		CARNET: 202110206

var logued = false //variable booleana para verificar si un usuario estará logueado en su sesión

func main() {
	for {
		fmt.Println("*=*=*=*=*=*= INGRESE UN COMANDO =*=*=*=*=*=")
		fmt.Println("*=*=*=*= Para terminar con la aplicación ingresar el comando exit")
		fmt.Print("\t")

		reader := bufio.NewReader(os.Stdin)
		entrada, _ := reader.ReadString('\n')
		eleccion := strings.TrimRight(entrada, "\r\n")
		if eleccion == "exit" {
			break
		}
		comando := Comando(eleccion)
		eleccion = strings.TrimSpace(eleccion)
		eleccion = strings.TrimLeft(eleccion, comando)
		tokens := SepararTokens(eleccion)
		funciones(comando, tokens)
		fmt.Println("\tPresione Enter para continuar...")
		fmt.Scanln()
	}
}

func Comando(text string) string {
	var tkn string
	terminar := false
	for i := 0; i < len(text); i++ {
		if terminar {
			if string(text[i]) == " " || string(text[i]) == "-" {
				break
			}
			tkn += string(text[i])
		} else if string(text[i]) != " " && !terminar {
			if string(text[i]) == "#" {
				tkn = text
			} else {
				tkn += string(text[i])
				terminar = true
			}
		}
	}
	return tkn
}

func SepararTokens(texto string) []string {
	var tokens []string
	if texto == "" {
		return tokens
	}
	texto += " "
	var token string
	estado := 0
	for i := 0; i < len(texto); i++ {
		c := string(texto[i])
		if estado == 0 && c == "-" {
			estado = 1
		} else if estado == 0 && c == "#" {
			continue
		} else if estado != 0 {
			if estado == 1 {
				if c == "=" {
					estado = 2
				} else if c == " " {
					continue
				} else if (c == "P" || c == "p") && string(texto[i+1]) == " " && string(texto[i-1]) == "-" {
					estado = 0
					tokens = append(tokens, c)
					token = ""
					continue
				} else if (c == "R" || c == "r") && string(texto[i+1]) == " " && string(texto[i-1]) == "-" {
					estado = 0
					tokens = append(tokens, c)
					token = ""
					continue
				}
			} else if estado == 2 {
				if c == " " {
					continue
				}
				if c == "\"" {
					estado = 3
					continue
				} else {
					estado = 4
				}
			} else if estado == 3 {
				if c == "\"" {
					estado = 4
					continue
				}
			} else if estado == 4 && c == "\"" {
				tokens = []string{}
				continue
			} else if estado == 4 && c == " " {
				estado = 0
				tokens = append(tokens, token)
				token = ""
				continue
			}
			token += c
		}
	}
	return tokens
}

func funciones(token string, tks []string) {
	if token != "" {
		if Comandos.Comparar(token, "EXECUTE") {
			fmt.Println("=*=*=*=*=*=*= FUNCION EXECUTE =*=*=*=*=*=*=")
			FuncionExec(tks)
		} else if Comandos.Comparar(token, "MKDISK") {
			fmt.Println("=*=*=*=*=*=*= FUNCION MKDISK =*=*=*=*=*=*=")
			Comandos.ValidarDatosMKDISK(tks)
		} else if Comandos.Comparar(token, "RMDISK") {
			fmt.Println("=*=*=*=*=*=*= FUNCION RMDISK =*=*=*=*=*=*=*=")
			Comandos.RMDISK(tks)
		} else if Comandos.Comparar(token, "FDISK") {
			fmt.Println("=*=*=*=*=*=*= FUNCION FDISK =*=*=*=*=*=*=*=")
			Comandos.ValidarDatosFDISK(tks)
		} else if Comandos.Comparar(token, "REP") {
			fmt.Println("=*=*=*=*=*=*= FUNCION REP =*=*=*=*=*=*=*=")
			Comandos.ValidarDatosREP(tks)
		} else if Comandos.Comparar(token, "MOUNT") {
			fmt.Println("=*=*=*=*=*=*= FUNCION MOUNT =*=*=*=*=*=*=*=")
			Comandos.ValidarDatosMOUNT(tks)
		} else if Comandos.Comparar(token, "UNMOUNT") {
			fmt.Println("=*=*=*=*=*=*= FUNCION UNMOUNT =*=*=*=*=*=*=*=")
			Comandos.ValidarDatosUNMOUNT(tks)
		} else if Comandos.Comparar(token, "MKFS") {
			fmt.Println("=*=*=*=*=*=*= FUNCION MKFS =*=*=*=*=*=*=*=")
			Comandos.ValidarDatosMKFS(tks)
		} else if Comandos.Comparar(token, "LOGIN") {
			fmt.Println("=*=*=*=*=*=*= FUNCION LOGIN =*=*=*=*=*=*=*=")
			if logued {
				Comandos.Error("LOGIN", "Ya hay un usuario en línea.")
				return
			} else {
				logued = Comandos.ValidarDatosLOGIN(tks)
			}
		} else if Comandos.Comparar(token, "LOGOUT") {
			fmt.Println("=*=*=*=*=*=*= FUNCION LOGOUT =*=*=*=*=*=*=*=")
			if !logued {
				Comandos.Error("LOGOUT", "Aún no se ha iniciado sesión.")
				return
			} else {
				logued = Comandos.CerrarSesion()
			}
		} else if Comandos.Comparar(token, "MKGRP") {
			fmt.Println("=*=*=*=*=*=*= FUNCION MKGRP =*=*=*=*=*=*=*=")
			if !logued {
				Comandos.Error("MKGRP", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosGrupos(tks, "MK")
			}
		} else if Comandos.Comparar(token, "RMGRP") {
			fmt.Println("=*=*=*=*=*=*= FUNCION RMGRP =*=*=*=*=*=*=*=")
			if !logued {
				Comandos.Error("RMGRP", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosGrupos(tks, "RM")
			}
		} else if Comandos.Comparar(token, "MKUSR") {
			fmt.Println("=*=*=*=*=*=*= FUNCION MKUSR =*=*=*=*=*=*=*=")
			if !logued {
				Comandos.Error("MKUSR", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosUsers(tks, "MK")
			}
		} else if Comandos.Comparar(token, "RMUSR") {
			fmt.Println("=*=*=*=*=*=*= FUNCION RMUSR =*=*=*=*=*=*=*=")
			if !logued {
				Comandos.Error("RMUSR", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosUsers(tks, "RM")
			}
		} else if Comandos.Comparar(token, "RMUSR") {
			fmt.Println("=*=*=*=*=*=*= FUNCION RMUSR =*=*=*=*=*=*=*=")
			if !logued {
				Comandos.Error("RMUSR", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosUsers(tks, "RM")
			}
		} else if Comandos.Comparar(token, "MKDIR") {
			fmt.Println("=*=*=*=*=*=*= FUNCION MKDIR =*=*=*=*=*=*=*=")
			if !logued {
				Comandos.Error("MKDIR", "Aún no se ha iniciado sesión con ningún usuario.")
				return
			} else {
				Comandos.ValidarDatosMKDIR(tks)
			}
		} else if Comandos.Comparar(token, "MKFILE") {
			fmt.Println("=*=*=*=*=*=*= FUNCION MKFILE =*=*=*=*=*=*=*=")
			if !logued {
				Comandos.Error("MKFILE", "Aún no se ha iniciado sesión con ningún usuario.")
				return
			} else {
				Comandos.ValidarDatosMKFILE(tks)
			}
		} else {
			Comandos.Error("ANALIZADOR", "No se reconoce el comando \""+token+"\"")
		}
	}
}

func FuncionExec(tokens []string) {
	path := ""
	for i := 0; i < len(tokens); i++ {
		datos := strings.Split(tokens[i], "=")
		if Comandos.Comparar(datos[0], "path") {
			path = datos[1]
		}
	}
	if path == "" {
		Comandos.Error("EXECUTE", "Se requiere el parámetro \"path\" para este comando")
		return
	}
	Exec(path)
}

func Exec(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error al abrir el archivo: %s", err)
	}
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		texto := fileScanner.Text()
		texto = strings.TrimSpace(texto)
		tk := Comando(texto)
		if texto != "" {
			if Comandos.Comparar(tk, "pause") {
				fmt.Println("************************************** FUNCIÓN PAUSE **************************************")
				var pause string
				Comandos.Mensaje("PAUSE", "Presione \"enter\" para continuar...")
				fmt.Scanln(&pause)
				continue
			} else if string(texto[0]) == "#" {
				fmt.Println("************************************** COMENTARIO **************************************")
				Comandos.Mensaje("COMENTARIO", texto)
				continue
			}
			texto = strings.TrimLeft(texto, tk)
			tokens := SepararTokens(texto)
			funciones(tk, tokens)
		}
	}
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error al leer el archivo: %s", err)
	}
	file.Close()
}
