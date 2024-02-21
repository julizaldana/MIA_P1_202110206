package Comandos

import (
	"MIA_P1_202110206/Structs"
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

// Estructura Transicion sirve para ir manejando e ir moviendonos dentro del archivo binario
type Transition struct {
	partition int
	start     int
	end       int
	before    int
	after     int
}

var startValue int

func ValidarDatosFDISK(tokens []string) {
	size := ""
	path := "" //MIA/P1, ya tengo que tener la ruta quemada, y tengo que tener el parametro driveletter - IMPORTANTE
	name := "" //Delete y Add agregar parametros
	unit := "k"
	tipo := "P"
	fit := "WF"
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "size") {
			size = tk[1]
		} else if Comparar(tk[0], "unit") {
			unit = tk[1]
		} else if Comparar(tk[0], "path") {
			path = strings.ReplaceAll(tk[1], "\"", "")
		} else if Comparar(tk[0], "type") {
			tipo = tk[1]
		} else if Comparar(tk[0], "fit") {
			fit = tk[1]
		} else if Comparar(tk[0], "name") {
			name = tk[1]
		}
	}
	if size == "" || path == "" || name == "" {
		Error("FDISK", "El comando FDISK necesita parametros obligatorios")
		return
	} else {
		generarParticion(size, unit, path, tipo, fit, name)
	}
}

func generarParticion(s string, u string, p string, t string, f string, n string) {
	startValue = 0
	i, error_ := strconv.Atoi(s) //string a entero, i es un entero
	if error_ != nil {
		Error("FDISK", "Size debe ser un número entero")
		return
	}
	if i <= 0 {
		Error("FDISK", "Size debe ser mayor que 0")
		return
	}
	if Comparar(u, "b") || Comparar(u, "k") || Comparar(u, "n") {
		if Comparar(u, "k") {
			i = i * 1024
		} else if Comparar(u, "n") {
			i = i * 1024 * 1024
		}
	} else {
		Error("FDISK", "Unit no contiene los valores esperados.")
		return
	}
	if !(Comparar(t, "p") || Comparar(t, "e") || Comparar(t, "l")) {
		Error("FDISK", "Type no contiene los valores esperados.")
		return
	}
	if !(Comparar(f, "bf") || Comparar(f, "ff") || Comparar(f, "wf")) {
		Error("FDISK", "Fit no contiene los valores esperados. ")
		return
	}
	mbr := leerDisco(p)
	particiones := GetParticiones(*mbr)
	var between []Transition //arreglo

	//Son contadores que nos van a ayudar
	usado := 0 //3 prim o 4 prim
	ext := 0   //1 ext
	c := 0
	base := int(unsafe.Sizeof(Structs.MBR{})) //tamaño de estructura de MBR
	extended := Structs.NewParticion()        //NOs va a servir para luego ver las particiones logicas

	//Este for es importante para ir moviendonos en la creacion de carpetas, etc
	for j := 0; j < len(particiones); j++ {
		prttn := particiones[j]
		if prttn.Part_status == '1' { //PARTICION SE ESTÁ UTILIZANDO
			var trn Transition //Datos de la particion como tal, nuevo objeto
			trn.partition = c
			trn.start = int(prttn.Part_start)
			trn.end = int(prttn.Part_start + prttn.Part_s)
			trn.before = trn.start - base //Base es el mbr
			base = trn.end
			if usado != 0 {
				between[usado-1].after = trn.start - (between[usado-1].end)
			}
			between = append(between, trn)
			usado++

			if prttn.Part_type == "e"[0] || prttn.Part_type == "E"[0] {
				ext++
				extended = prttn
			}
		}
		if usado == 4 && !Comparar(t, "l") {
			Error("FDISK", "Limite de particiones alcanzado")
			return
		} else if ext == 1 && Comparar(t, "e") {
			Error("FDISK", "Solo se puede crear una partición extendida")
			return
		}
		c++
	}
	if ext == 0 && Comparar(t, "l") {
		Error("FDISK", "Aún no se han creado particiones extendidas, no se puede agregar")
		return
	}
	if usado != 0 { //que ya existe una particion
		between[len(between)-1].after = int(mbr.Mbr_tamano) - between[len(between)-1].end
	}
	regresa := BuscarParticiones(*mbr, n, p)
	if regresa != nil {
		Error("FDISK", "El nombre: "+n+", ya está en uso.")
		return
	}
	temporal := Structs.NewParticion()
	temporal.Part_status = '1'
	temporal.Part_s = int64(i)
	temporal.Part_type = strings.ToUpper(t)[0]
	temporal.Part_fit = strings.ToUpper(f)[0]
	copy(temporal.Part_name[:], n)

	if Comparar(t, "l") {
		//CREAR PARTICION LOGICA
		//Logica(temporal, extended, p)
		return
	}
	//HACER EL AJUSTE DE LA PARTICION.
	mbr = ajustar(*mbr, temporal, between, particiones, usado)
	if mbr == nil {
		return
	}
	file, err := os.OpenFile(strings.ReplaceAll(p, "\"", ""), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		Error("FDISK", "Error al abrir el archivo")
	}
	file.Seek(0, 0)
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, mbr)
	EscribirBytes(file, binario2.Bytes())
	if Comparar(t, "E") {
		ebr := Structs.NewEBR()
		ebr.Part_mount = '0'
		ebr.Part_start = int64(startValue)
		ebr.Part_s = 0
		ebr.Part_next = -1

		file.Seek(int64(startValue), 0)
		var binario3 bytes.Buffer
		binary.Write(&binario3, binary.BigEndian, ebr)
		EscribirBytes(file, binario3.Bytes())
		Mensaje("FDISK", "Partición Extendida: "+n+", creada correctamente.")
		return
	}
	file.Close()
	Mensaje("FDISK", "Partición Primaria: "+n+", creada correctamente.")
}

// LEE EL MBR, Y RETORNA INFORMACION DE PARTICION 1, PARTICION 2, PARTICION 3, PARTICION 4, COMO SI FUESE UN ARREGLO
func GetParticiones(disco Structs.MBR) []Structs.Particion {
	var v []Structs.Particion
	v = append(v, disco.Mbr_partition_1)
	v = append(v, disco.Mbr_partition_2)
	v = append(v, disco.Mbr_partition_3)
	v = append(v, disco.Mbr_partition_4)
	return v
}

func BuscarParticiones(mbr Structs.MBR, name string, path string) *Structs.Particion {
	var particiones [4]Structs.Particion
	particiones[0] = mbr.Mbr_partition_1
	particiones[1] = mbr.Mbr_partition_2
	particiones[2] = mbr.Mbr_partition_3
	particiones[3] = mbr.Mbr_partition_4

	ext := false
	extended := Structs.NewParticion()
	for i := 0; i < len(particiones); i++ {
		particion := particiones[i]
		if particion.Part_status == "1"[0] {
			nombre := ""
			for j := 0; j < len(particion.Part_name); j++ {
				if particion.Part_name[j] != 0 {
					nombre += string(particion.Part_name[j])
				}
			}
			if Comparar(nombre, name) {
				return &particion
			} else if particion.Part_type == "E"[0] || particion.Part_type == "e"[0] {
				ext = true
				extended = particion
			}
		}
	}
	if ext {
		ebrs := GetLogicas(extended, path)
		for i := 0; i < len(ebrs); i++ {
			ebr := ebrs[i]
			if ebr.Part_mount == '1' {
				nombre := ""
				for j := 0; j < len(ebr.Part_name); j++ {
					if ebr.Part_name[j] != 0 {
						nombre += string(ebr.Part_name[j])
					}
				}
				if Comparar(nombre, name) {
					tmp := Structs.NewParticion()
					tmp.Part_status = '1'
					tmp.Part_type = 'L'
					tmp.Part_fit = ebr.Part_fit
					tmp.Part_start = ebr.Part_start
					tmp.Part_s = ebr.Part_s
					copy(tmp.Part_name[:], ebr.Part_name[:])
					return &tmp
				}
			}
		}
	}
	return nil
}

//Get Logicas nos retornará un arreglo de ebrs, ahi tenemos toda la informacion de las particiones logicas

func GetLogicas(particion Structs.Particion, path string) []Structs.EBR {
	var ebrs []Structs.EBR
	file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("FDISK", "Error al abrir el archivo")
		return nil
	}
	file.Seek(0, 0)
	tmp := Structs.NewEBR()
	file.Seek(particion.Part_start, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &tmp)
	if err_ != nil {
		Error("FDISK", "Error al leer el archivo")
		return nil
	}
	for {
		if int(tmp.Part_next) != -1 && int(tmp.Part_mount) != 0 {
			ebrs = append(ebrs, tmp)
			file.Seek(tmp.Part_next, 0)

			data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
			buffer = bytes.NewBuffer(data)
			err_ = binary.Read(buffer, binary.BigEndian, &tmp)
			if err_ != nil {
				Error("FDISK", "Error al leer el archivo")
				return nil
			}
		} else {
			file.Close()
			break
		}
	}
	return ebrs
}

func Logica(particion Structs.Particion, ep Structs.Particion, path string) {

}

// Transition funciona para ir moviendose en el ajuste.
func ajustar(mbr Structs.MBR, p Structs.Particion, t []Transition, ps []Structs.Particion, u int) {

}
