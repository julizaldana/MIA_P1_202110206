package Comandos

import (
	"MIA_P1_202110206/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

var DiscMont [99]DiscoMontado

type DiscoMontado struct {
	Path        [150]byte
	Estado      byte
	Particiones [26]ParticionMontada
}

type ParticionMontada struct {
	Letra  byte
	Estado byte
	Nombre [20]byte
}

var alfabeto = []byte{'a', 'b', 'c', 'd', 'e', 'f'}

func ValidarDatosMOUNT(context []string) {
	name := ""
	path := "" //SE QUITA Y LE COLOCO EL DRIVELETTER -> Para ir a buscar el archivo binario
	for i := 0; i < len(context); i++ {
		current := context[i]

		comando := strings.Split(current, "=")

		if Comparar(comando[0], "name") {
			name = comando[1]
		} else if Comparar(comando[0], "path") {
			path = strings.ReplaceAll(comando[1], "\"", "")
		}

	}
	if path == "" || name == "" {
		Error("MOUNT", "El comando MOUNT requiere parametros obligatorios")
		return
	}
	mount(path, name)
	listaMount()
}

func mount(p string, n string) {
	file, error_ := os.Open(p)
	if error_ != nil {
		Error("MOUNT", "No se ha podido abrir el archivo.")
		return
	}

	disk := Structs.NewMBR()
	file.Seek(0, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &disk)
	if err_ != nil {
		Error("FDISK", "Error al leer el archivo.")
		return
	}
	file.Close()

	particion := BuscarParticiones(disk, n, p)
	if particion.Part_type == 'E' || particion.Part_type == 'L' {
		var nombre [16]byte
		copy(nombre[:], n)
		if particion.Part_name == nombre && particion.Part_type == 'E' {
			Error("MOUNT", "No se puede montar una partición extendida.")
			return
		} else {
			ebrs := GetLogicas(*particion, p)
			encontrada := false
			if len(ebrs) != 0 {
				for i := 0; i < len(ebrs); i++ {
					ebr := ebrs[i]
					nombreebr := ""
					for j := 0; j < len(ebr.Part_name); j++ {
						if ebr.Part_name[j] != 0 {
							nombreebr += string(ebr.Part_name[j])
						}
					}
					if Comparar(nombreebr, n) && ebr.Part_mount == '1' {
						encontrada = true
						n = nombreebr
						break
					} else if nombreebr == n && ebr.Part_mount == '0' {
						Error("MOUNT", "No se puede montar una particion logica eliminada")
						return
					}
				}
				if !encontrada {
					Error("MOUNT", "No se encontró la partición lógica.")
					return
				}
			}
		}
	}
	for i := 0; i < 99; i++ {
		var ruta [150]byte
		copy(ruta[:], p)
		if DiscMont[i].Path == ruta {
			for j := 0; j < 26; j++ {
				var nombre [20]byte
				copy(nombre[:], n)
				if DiscMont[i].Particiones[j].Nombre == nombre {
					Error("MOUNT", "Ya se ha montado la partición "+n)
					return
				}
				if DiscMont[i].Particiones[j].Estado == 0 {
					DiscMont[i].Particiones[j].Estado = 1
					DiscMont[i].Particiones[j].Letra = alfabeto[j]
					copy(DiscMont[i].Particiones[j].Nombre[:], n)
					//re := strconv.Itoa(i+1) + string(alfabeto[j])
					Mensaje("MOUNT", "Se ha realizado correctamente el mount -id = ")
					return
				}
			}
		}
	}
	for i := 0; i < 99; i++ {
		if DiscMont[i].Estado == 0 {
			DiscMont[i].Estado = 1
			copy(DiscMont[i].Path[:], p)
			for j := 0; j < 26; j++ {
				if DiscMont[i].Particiones[j].Estado == 0 {
					DiscMont[i].Particiones[j].Estado = 1
				}
			}
		}
	}

}
func listaMount() {
	fmt.Println("\n<-*-*-*-*-*-*-*-*-* LISTADO DE PARTICIONES MONTADAS -*-*-*-*-*-*-*-*-*-*>")
	for i := 0; i < 99; i++ {
		for j := 0; j < 26; j++ {
			if DiscMont[i].Particiones[j].Estado == 1 {
				nombre := ""
				for k := 0; k < len(DiscMont[i].Particiones[j].Nombre); k++ {
					if DiscMont[i].Particiones[j].Nombre[k] != 0 {
						nombre += string(DiscMont[i].Particiones[j].Nombre[k])
					}
				}
				fmt.Println("\t id: 06" + strconv.Itoa(i+1) + string(alfabeto[j]) + ", Nombre: " + nombre)
			} //HAY QUE CAMBIARLO PORQUE
		}
	}
}
