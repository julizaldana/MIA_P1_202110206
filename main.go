package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		fmt.Println("*=*=*=*=*=*= INGRESE UN COMANDO =*=*=*=*=*=")
		fmt.Println("*=*=*=*= Para terminar con la aplicación ingresar -exit")
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
				}
			}
		}

	}
}

func funciones(token string, tks []string) {
	if token != "" {
		if Comandos.Comparar(token, "EXEC") {

		}
	}
}
