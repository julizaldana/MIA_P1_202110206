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
	Tipo   [20]byte
	Estado byte
	Id     [4]byte
	Nombre [20]byte
}

// CARNET -> 202110206 (ULTIMOS DOS DIGITOS -> 06)

func ValidarDatosMOUNT(context []string) {
	name := ""
	driveLetter := "" //SE QUITA Y LE COLOCO EL DRIVELETTER -> Para ir a buscar el archivo binario
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "name") {
			name = comando[1]
		}
		if Comparar(comando[0], "driveletter") {
			driveLetter = comando[1]
		}

	}
	if driveLetter == "" || name == "" {
		Error("MOUNT", "El comando MOUNT requiere parametros obligatorios")
		return
	} else {
		// Construir la ruta del archivo basado en el driveLetter
		rutaBase := "/home/julio/Escritorio/MIA/P1/"
		nombreDisco := driveLetter + ".dsk"
		path := rutaBase + nombreDisco

		if !ArchivoExiste(path) {
			Error("MOUNT", "No se encontró el disco con dicho driveletter")
			return
		}

		mount(path, name, driveLetter)
		listaMount()
	}

}

func mount(p string, n string, d string) {
	file, error_ := os.Open(p)
	if error_ != nil {
		Error("MOUNT", "No se ha podido abrir el archivo.")
		return
	}
	// Obtener el número de la partición del nombre, donde se corta el nombre de la partición. EJ: Part2, me queda solo (2), Part3, me queda solo (3)
	numParticion, err := strconv.Atoi(strings.TrimPrefix(n, "Part"))
	if err != nil {
		Error("MOUNT", "No se pudo obtener el número de la partición del nombre.")
		return
	}

	disk := Structs.NewMBR()
	file.Seek(0, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &disk)
	if err_ != nil {
		Error("MOUNT", "Error al leer el archivo")
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
						Error("MOUNT", "No se puede montar una partición Lógica eliminada.")
						return
					}
				}
				if !encontrada {
					Error("MOUNT", "No se encontró la partición Lógica.")
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
					copy(DiscMont[i].Particiones[j].Nombre[:], n)
					re := d + strconv.Itoa(numParticion) + "06"
					copy(DiscMont[i].Particiones[j].Id[:], re)
					Mensaje("MOUNT", "Se ha realizado correctamente el mount -id = "+re)
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
					copy(DiscMont[i].Particiones[j].Nombre[:], n)
					re := d + strconv.Itoa(numParticion) + "06"
					copy(DiscMont[i].Particiones[j].Id[:], re)
					Mensaje("MOUNT", "Se ha realizado correctamente el mount -id = "+re)
					return
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
				id := ""
				for k := 0; k < len(DiscMont[i].Particiones[j].Nombre); k++ {
					if DiscMont[i].Particiones[j].Nombre[k] != 0 {
						nombre += string(DiscMont[i].Particiones[j].Nombre[k])
					}
				}
				for k := 0; k < len(DiscMont[i].Particiones[j].Id); k++ {
					if DiscMont[i].Particiones[j].Id[k] != 0 {
						id += string(DiscMont[i].Particiones[j].Id[k])
					}
				}
				fmt.Println("\t id: " + id + " || " + "nombre: " + nombre)
			}
		}
	}
}

func ValidarDatosUNMOUNT(context []string) {
	id := ""
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "id") {
			id = comando[1]
		}
	}
	if id == "" {
		Error("UNMOUNT", "El comando UNMOUNT requiere el id de forma obligatoria")
		return
	} else {
		unmount(id)
		listaMount()
	}

}

// SE HACE UNA FUNCION UNMOUNT PARA PODER BUSCAR EL ID DENTRO DE LA ESTRUCTURA DE PARTICIONES Y SE DESMONTA SI EXISTE
func unmount(id string) {
	for i := 0; i < 99; i++ {
		for j := 0; j < 26; j++ {
			if DiscMont[i].Particiones[j].Estado == 1 {
				currentID := ""
				for k := 0; k < len(DiscMont[i].Particiones[j].Id); k++ {
					if DiscMont[i].Particiones[j].Id[k] != 0 {
						currentID += string(DiscMont[i].Particiones[j].Id[k])
					}
				}
				if currentID == id {
					// Desmontar la partición, y se coloca estado 0, de que no está montada la partición
					DiscMont[i].Particiones[j].Estado = 0
					DiscMont[i].Particiones[j].Tipo = [20]byte{}
					DiscMont[i].Particiones[j].Id = [4]byte{}
					DiscMont[i].Particiones[j].Nombre = [20]byte{}
					fmt.Println("Partición desmontada exitosamente.")
					return
				}
			}
		}
	} //SI NO SE ENCUENTRA LA PARTICION EN LA ESTRUCTURA, se manda un mensaje de que no se encontró
	Error("UNMOUNT", "No se encontró la partición con el ID proporcionado.")
}
