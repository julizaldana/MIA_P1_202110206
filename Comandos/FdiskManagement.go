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
	driveLetter := "" //MIA/P1, ya tengo que tener la ruta quemada, y tengo que tener el parametro driveletter - IMPORTANTE
	name := ""        //Delete y Add agregar parametros
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
		} else if Comparar(tk[0], "driveletter") {
			driveLetter = tk[1]
		} else if Comparar(tk[0], "type") {
			tipo = tk[1]
		} else if Comparar(tk[0], "fit") {
			fit = tk[1]
		} else if Comparar(tk[0], "name") {
			name = tk[1]
		}
	}

	if size == "" || driveLetter == "" || name == "" {
		Error("FDISK", "El comando FDISK necesita parametros obligatorios")
		return
	} else {
		// Construir la ruta del archivo basado en el driveLetter
		rutaBase := "/home/julio/Escritorio/MIA/P1/"
		nombreDisco := driveLetter + ".dsk"
		path := rutaBase + nombreDisco

		if !ArchivoExiste(path) {
			Error("FDISK", "No se encontró el disco con dicho driveletter")
			return
		}

		// Pasar la ruta construida a la función generarParticion
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
	if Comparar(u, "b") || Comparar(u, "k") || Comparar(u, "m") {
		if Comparar(u, "k") {
			i = i * 1024
		} else if Comparar(u, "m") {
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
		Error("FDISK", "Fit no contiene los valores esperados.")
		return
	}
	mbr := leerDisco(p)
	particiones := GetParticiones(*mbr)
	var between []Transition //arreglo

	//Son contadores que nos van a ayudar
	usado := 0 //3 primarias o 4 primarias
	ext := 0   //1 extendida
	c := 0
	base := int(unsafe.Sizeof(Structs.MBR{})) //tamaño de estructura de mbr
	extended := Structs.NewParticion()        //Nos va a servir para luego ver las particiones logicas

	//Este for es importante para ir moviendonos en la creación de carpetas, etc
	for j := 0; j < len(particiones); j++ {
		prttn := particiones[j]
		if prttn.Part_status == '1' { //PARTICIONES SE ESTÁN UTILIZANDO
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
		Error("FDISK", "Aún no se han creado particiones extendidas, no se puede agregar una lógica.")
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
		Logica(temporal, extended, p) //SE CREA LA PARTICION LOGICA
		return
	}
	//HACER EL AJUSTE DE LA PARTICIÓN
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

		file.Seek(int64(startValue), 0) //5200
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
		Error("FDSIK", "Error al leer el archivo")
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
				Error("FDSIK", "Error al leer el archivo")
				return nil
			}
		} else {
			file.Close()
			break
		}
	}

	return ebrs
}

// Funcion para crear una particion logica
func Logica(particion Structs.Particion, ep Structs.Particion, path string) {
	logic := Structs.NewEBR()
	logic.Part_mount = '1'
	logic.Part_fit = particion.Part_fit
	logic.Part_s = particion.Part_s
	logic.Part_next = -1
	copy(logic.Part_name[:], particion.Part_name[:])

	file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("FDISK", "Error al abrir el archivo del disco.")
		return
	}
	file.Seek(0, 0)

	tmp := Structs.NewEBR()
	tmp.Part_mount = 0
	tmp.Part_s = 0
	tmp.Part_next = -1
	file.Seek(ep.Part_start, 0) //0

	data := leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &tmp)

	if err_ != nil {
		Error("FDSIK", "Error al leer el archivo")
		return
	}
	if err != nil {
		Error("FDISK", "Error al abrir el archivo del disco.")
		return
	}
	var size int64 = 0
	file.Close()
	for {
		size += int64(unsafe.Sizeof(Structs.EBR{})) + tmp.Part_s
		if (tmp.Part_s == 0 && tmp.Part_next == -1) || (tmp.Part_s == 0 && tmp.Part_next == 0) {
			file, err = os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
			logic.Part_start = tmp.Part_start
			logic.Part_next = logic.Part_start + logic.Part_s + int64(unsafe.Sizeof(Structs.EBR{}))
			if (ep.Part_s - size) <= logic.Part_s {
				Error("FDISK", "No queda más espacio para crear más particiones lógicas")
				return
			}
			file.Seek(logic.Part_start, 0)

			var binario2 bytes.Buffer
			binary.Write(&binario2, binary.BigEndian, logic)
			EscribirBytes(file, binario2.Bytes())
			nombre := ""
			for j := 0; j < len(particion.Part_name); j++ {
				nombre += string(particion.Part_name[j])
			}
			file.Seek(logic.Part_next, 0)
			addLogic := Structs.NewEBR()
			addLogic.Part_mount = '0'
			addLogic.Part_next = -1
			addLogic.Part_start = logic.Part_next

			file.Seek(addLogic.Part_start, 0)

			var binarioLogico bytes.Buffer
			binary.Write(&binarioLogico, binary.BigEndian, addLogic)
			EscribirBytes(file, binarioLogico.Bytes())

			Mensaje("FDISK", "Partición Lógica: "+nombre+", creada correctamente.")
			file.Close()
			return
		}
		file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
		if err != nil {
			Error("FDISK", "Error al abrir el archivo del disco.")
			return
		}
		file.Seek(tmp.Part_next, 0)
		data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &tmp)

		if err_ != nil {
			Error("FDSIK", "Error al leer el archivo")
			return
		}
	}
}

// FUNCION PARA REALIZAR EL AJUSTE DE PARTICIONES
func ajustar(mbr Structs.MBR, p Structs.Particion, t []Transition, ps []Structs.Particion, u int) *Structs.MBR {
	if u == 0 {
		p.Part_start = int64(unsafe.Sizeof(mbr))
		startValue = int(p.Part_start)
		mbr.Mbr_partition_1 = p
		return &mbr
	} else {
		var usar Transition
		c := 0
		for i := 0; i < len(t); i++ {
			tr := t[i]
			if c == 0 {
				usar = tr
				c++
				continue
			}

			if Comparar(string(mbr.Dsk_fit[0]), "F") {
				if int64(usar.before) >= p.Part_s || int64(usar.after) >= p.Part_s {
					break
				}
				usar = tr
			} else if Comparar(string(mbr.Dsk_fit[0]), "B") {
				if int64(tr.before) >= p.Part_s || int64(usar.after) < p.Part_s {
					usar = tr
				} else {
					if int64(tr.before) >= p.Part_s || int64(tr.after) >= p.Part_s {
						b1 := usar.before - int(p.Part_s)
						a1 := usar.after - int(p.Part_s)
						b2 := tr.before - int(p.Part_s)
						a2 := tr.after - int(p.Part_s)

						if (b1 < b2 && b1 < a2) || (a1 < b2 && a1 < a2) {
							c++
							continue
						}
						usar = tr
					}
				}
			} else if Comparar(string(mbr.Dsk_fit[0]), "W") {
				if int64(usar.before) >= p.Part_s || int64(usar.after) < p.Part_s {
					usar = tr
				} else {
					if int64(tr.before) >= p.Part_s || int64(tr.after) >= p.Part_s {
						b1 := usar.before - int(p.Part_s)
						a1 := usar.after - int(p.Part_s)
						b2 := tr.before - int(p.Part_s)
						a2 := tr.after - int(p.Part_s)

						if (b1 > b2 && b1 > a2) || (a1 > b2 && a1 > a2) {
							c++
							continue
						}
						usar = tr
					}
				}
			}
			c++
		}
		if usar.before >= int(p.Part_s) || usar.after >= int(p.Part_s) {
			if Comparar(string(mbr.Dsk_fit[0]), "F") {
				if usar.before >= int(p.Part_s) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			} else if Comparar(string(mbr.Dsk_fit[0]), "B") {
				b1 := usar.before - int(p.Part_s)
				a1 := usar.after - int(p.Part_s)

				if (usar.before >= int(p.Part_s) && b1 < a1) || usar.after < int(p.Part_start) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			} else if Comparar(string(mbr.Dsk_fit[0]), "W") {
				b1 := usar.before - int(p.Part_s)
				a1 := usar.after - int(p.Part_s)

				if (usar.before >= int(p.Part_s) && b1 > a1) || usar.after < int(p.Part_start) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			}
			var partitions [4]Structs.Particion
			for i := 0; i < len(ps); i++ {
				partitions[i] = ps[i]
			}

			for i := 0; i < len(partitions); i++ {
				partition := partitions[i]
				if partition.Part_status != '1' {
					partitions[i] = p
					break
				}
			}
			mbr.Mbr_partition_1 = partitions[0]
			mbr.Mbr_partition_2 = partitions[1]
			mbr.Mbr_partition_3 = partitions[2]
			mbr.Mbr_partition_4 = partitions[3]
			return &mbr
		} else {
			Error("FDISK", "No hay espacio suficiente.")
			return nil
		}
	}
}

func VerificarNombreParticion(name string) {

}
