package Comandos

import (
	"MIA_P1_202110206/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

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
		dks(path, id)
	} else if Comparar(name, "TREE") {
		tree(path, id)
	} else if Comparar(name, "MBR") {
		mbr(path, id)
	} else if Comparar(name, "INODE") {
		inoder(path, id)
	} else if Comparar(name, "BLOCK") {
		blockr(path, id)
	} else if Comparar(name, "JOURNALING") {
		journalr(path, id)
	} else if Comparar(name, "SB") {
		superblockr(path, id)
	} else if Comparar(name, "BM_INODE") {
		bminoder(path, id)
	} else if Comparar(name, "BM_BLOCK") {
		bmblockr(path, id)
	} else if Comparar(name, "FILE") {
		if ruta == "" {
			Error("REP", "Se espera el parámetro ruta.")
		}
		fileR(path, id, ruta)
	} else {
		Error("REP", name+", no es un reporte válido.")
		return
	}
}

// REPORTE DISK
func dks(p string, id string) {
	var pth string
	GetMount("REP", id, &pth)

	//file
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return
	}
	var disk Structs.MBR
	file.Seek(0, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &disk)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}
	file.Close()

	aux := strings.Split(p, ".")
	if len(aux) > 2 {
		Error("REP", "No se admiten nombres de archivos que contengan punto (.")
		return
	}
	pd := aux[0] + ".dot"

	carpeta := ""
	direccion := strings.Split(pd, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(pd, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(pd)
	} else {
		fileaux.Close()
	}

	partitions := GetParticiones(disk)
	var extended Structs.Particion
	ext := false
	for i := 0; i < 4; i++ {
		if partitions[i].Part_status == '1' {
			if partitions[i].Part_type == "E"[0] || partitions[i].Part_type == "e"[0] {
				ext = true
				extended = partitions[i]
			}
		}
	}

	content := "digraph G{\n rankdir=TB;\n forcelabels= true;\n graph [ dpi = \"600\"] ; \n node [shape = plaintext];\n nodo1 [label = <<table>\n <tr>\n <td ROWSPAN='2'> \"MBR\" </td>"
	var positions [5]int64
	var positionsii [5]int64
	positions[0] = disk.Mbr_partition_1.Part_start - (1 + int64(unsafe.Sizeof(Structs.MBR{})))
	positions[1] = disk.Mbr_partition_2.Part_start - disk.Mbr_partition_1.Part_start + disk.Mbr_partition_1.Part_s
	positions[2] = disk.Mbr_partition_3.Part_start - disk.Mbr_partition_2.Part_start + disk.Mbr_partition_2.Part_s
	positions[3] = disk.Mbr_partition_4.Part_start - disk.Mbr_partition_3.Part_start + disk.Mbr_partition_3.Part_s
	positions[4] = disk.Mbr_tamano + 1 - disk.Mbr_partition_4.Part_s + disk.Mbr_partition_4.Part_s
	copy(positionsii[:], positions[:])

	logic := 0
	tmplogic := ""
	if ext {
		tmplogic = "<tr>\n"
		auxEbr := Structs.NewEBR()

		file, err = os.Open(strings.ReplaceAll(pth, "\n", ""))

		if err != nil {
			Error("REP", "No se ha encontrado el disco")
			return
		}

		file.Seek(extended.Part_start, 0)
		data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &auxEbr)
		if err_ != nil {
			Error("REP", "Error al leer el archivo")
			return
		}
		file.Close()
		var tamGen int64 = 0
		for auxEbr.Part_next != -1 {
			tamGen += auxEbr.Part_s
			res := float64(auxEbr.Part_s) / float64(disk.Mbr_tamano)
			res = res * 100
			tmplogic += "<td>\"EBR\"</td>"
			s := fmt.Sprintf("%.2f", res)
			tmplogic += "<td>\"Logica\n " + s + "% de la particion extendida\"</td>\n"

			resta := float64(auxEbr.Part_next) - (float64(auxEbr.Part_start) + float64(auxEbr.Part_s))
			resta = resta / float64(disk.Mbr_tamano)
			resta = resta * 10000.00
			resta = math.Round(resta) / 100.00 //PARA OBTENER LOS PORCENTAJES
			if resta != 0 {
				s = fmt.Sprintf("%f", resta)
				tmplogic += "<td>\"Logica\n " + s + "% libre de la partición extendida\"</td>\n"
				logic++
			}
			logic += 2 //Son los id, para los nodos de graphviz
			file, err = os.Open(strings.ReplaceAll(pth, "\"", ""))

			if err != nil {
				Error("REP", "No se ha encontrado el disco")
				return
			}

			file.Seek(auxEbr.Part_next, 0)
			data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
			buffer = bytes.NewBuffer(data)
			err_ = binary.Read(buffer, binary.BigEndian, &auxEbr)
			if err_ != nil {
				Error("REP", "Error al leer el archivo")
				return
			}
			file.Close()
		}
		resta := float64(extended.Part_s) - float64(tamGen)
		resta = resta / float64(disk.Mbr_tamano)
		resta = math.Round(resta * 100)
		if resta != 0 {
			s := fmt.Sprintf("%.2f", resta)
			tmplogic += "<td>\"Libre \n" + s + "% de la partición extendida. \"</td>\n"
			logic++
		}
		tmplogic += "</tr>\n\n"
		logic += 2
	}
	var tamPrim int64
	for i := 0; i < 4; i++ {
		if partitions[i].Part_type == 'E' {
			tamPrim += partitions[i].Part_s
			res := float64(partitions[i].Part_s) / float64(disk.Mbr_tamano)
			res = math.Round(res*10000.00) / 100.00
			s := fmt.Sprintf("%.2f", res)
			content += "<td COLSPAN='" + strconv.Itoa(logic) + "'> Extendida \n" + s + "% del disco </td>\n"
		} else if partitions[i].Part_start != -1 {
			tamPrim += partitions[i].Part_s
			res := float64(partitions[i].Part_s) / float64(disk.Mbr_tamano)
			res = math.Round(res*10000.00) / 100.00
			s := fmt.Sprintf("%.2f", res)
			content += "<td ROWSPAN='2'> Primaria \n" + s + "% del disco </td>\n"
		}
	}

	if tamPrim != 0 {
		libre := disk.Mbr_tamano - tamPrim
		res := float64(libre) / float64(disk.Mbr_tamano)
		res = math.Round(res * 100)
		s := fmt.Sprintf("%.2f", res)
		content += "<td ROWSPAN='2'> Libre \n" + s + "% del disco </td>"

	}
	content += "</tr>\n\n"
	content += tmplogic
	content += "</table>>];\n}\n"

	//fmt.Println(content)

	//CREAR IMAGEN
	b := []byte(content)
	err_ = ioutil.WriteFile(pd, b, 0644)
	if err_ != nil {
		log.Fatal(err_)
	}

	terminacion := strings.Split(p, ".")

	path, _ := exec.LookPath("dot")
	cmd, _ := exec.Command(path, "-T"+terminacion[1], pd).Output()
	node := int(0777)
	ioutil.WriteFile(p, cmd, os.FileMode(node))
	disco := strings.Split(pth, "/")
	Mensaje("REP", "Reporte tipo DISK del disco "+disco[len(disco)-1]+",creado correctamente")

}

// REPORTE MBR Y EBR
func mbr(p string, id string) {
	var pth string
	GetMount("REP", id, &pth)

	//file
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return
	}
	var disk Structs.MBR
	file.Seek(0, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &disk)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}
	file.Close()

	aux := strings.Split(p, ".")
	if len(aux) > 2 {
		Error("REP", "No se admiten nombres de archivos que contengan punto (.)")
		return
	}
	pd := aux[0] + ".dot"

	carpeta := ""
	direccion := strings.Split(pd, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(pd, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(pd)
	} else {
		fileaux.Close()
	}

	//mbrTamanoStr := strconv.FormatFloat(float64(disk.Mbr_tamano), 'f', -1, 64)
	content := "digraph G {\n  node0 [shape=none label=<\n  <TABLE style=\"rounded\" bgcolor=\"#d5f2e9\">\n  <TR>\n  <TD COLSPAN = '2' bgcolor=\"#34e5b0\">REPORTE MBR</TD>\n  </TR>\n  <TR>\n " +
		"<TD bgcolor=\" #cff5e5 \">mbr_tamano</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(disk.Mbr_tamano)) + "</TD>\n  </TR>" + "<TR>\n  <TD bgcolor=\" #cff5e5 \">mbr_fecha_creacion</TD>\n  <TD bgcolor=\" #cff5e5 \">" + string(disk.Mbr_fecha_creacion[:]) +
		"</TD>\n  </TR>\n  <TR>\n  <TD bgcolor=\" #cff5e5 \">mbr_disk_signature</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(disk.Mbr_dsk_signature)) + "</TD>\n  </TR>"

	contenidoParticiones := graficarParticiones(disk)
	content += contenidoParticiones

	file, err = os.Open(strings.ReplaceAll(pth, "\n", ""))
	if err != nil {
		Error("REP", "No se ha encontrado el disco")
		return
	}

	partitions := GetParticiones(disk)
	ext := false
	var extended Structs.Particion
	for i := 0; i < 4; i++ {
		if partitions[i].Part_status == '1' {
			if partitions[i].Part_type == "E"[0] || partitions[i].Part_type == "e"[0] {
				ext = true
				extended = partitions[i]
			}
		}
	}

	if ext {
		auxEbr := Structs.NewEBR()
		file.Seek(extended.Part_start, 0)
		data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &auxEbr)
		if err_ != nil {
			Error("REP", "Error al leer el archivo")
			return
		}

		for auxEbr.Part_next != -1 {
			partName := strings.TrimRight(string(auxEbr.Part_name[:]), "\x00")
			content += "<TR>\n" +
				"<TD COLSPAN = '2' bgcolor=\"#63bdcf\">Particion Logica (EBR)</TD>\n" +
				"</TR>\n" +
				"<TR>\n" +
				"<TD bgcolor=\"#a9dee9\">part_status</TD>\n" +
				"<TD bgcolor=\"#a9dee9\">" + string(auxEbr.Part_mount) + "</TD>\n" +
				"</TR>\n" +
				"<TR>\n" +
				"<TD bgcolor=\"#a9dee9\">part_fit</TD>\n" +
				"<TD bgcolor=\"#a9dee9\">" + string(auxEbr.Part_fit) + "</TD>\n" +
				"</TR>\n" +
				"<TR>\n" +
				"<TD bgcolor=\"#a9dee9\">part_start</TD>\n" +
				"<TD bgcolor=\"#a9dee9\">" + strconv.Itoa(int(auxEbr.Part_start)) + "</TD>\n" +
				"</TR>\n" +
				"<TR>\n" +
				"<TD bgcolor=\"#a9dee9\">part_size</TD>\n" +
				"<TD bgcolor=\"#a9dee9\">" + strconv.Itoa(int(auxEbr.Part_s)) + "</TD>\n" +
				"</TR>\n" +
				"<TR>\n" +
				"<TD bgcolor=\"#a9dee9\">part_next</TD>\n" +
				"<TD bgcolor=\"#a9dee9\">" + strconv.Itoa(int(auxEbr.Part_next)) + "</TD>\n" +
				"</TR>\n" +
				"<TR>\n" +
				"<TD bgcolor=\"#a9dee9\">part_name</TD>\n" +
				"<TD bgcolor=\"#a9dee9\">" + partName + "</TD>\n" +
				"</TR>\n"

			// Mover el puntero al siguiente registro de arranque extendido
			file.Seek(auxEbr.Part_next, 0)

			// Leer el siguiente EBR
			data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
			buffer = bytes.NewBuffer(data)
			err_ = binary.Read(buffer, binary.BigEndian, &auxEbr)
			if err_ != nil {
				Error("REP", "Error al leer el archivo")
				return
			}
		}

		file.Close() // Cerrar el archivo después de terminar de leer todos los EBRs
	}

	content += "</TABLE>>];\n\n}"

	//fmt.Println(content)

	//CREAR IMAGEN
	b := []byte(content)
	err_ = ioutil.WriteFile(pd, b, 0644)
	if err_ != nil {
		log.Fatal(err_)
	}

	terminacion := strings.Split(p, ".")

	path, _ := exec.LookPath("dot")
	cmd, _ := exec.Command(path, "-T"+terminacion[1], pd).Output()
	node := int(0777)
	ioutil.WriteFile(p, cmd, os.FileMode(node))
	disco := strings.Split(pth, "/")
	Mensaje("REP", "Reporte tipo MBR del disco "+disco[len(disco)-1]+",creado correctamente")

}

func graficarParticiones(disk Structs.MBR) string {
	contenido := ""
	particiones := GetParticiones(disk)
	for i := 0; i < len(particiones); i++ {
		// Solo graficar la partición si su part_start es diferente de -1
		if particiones[i].Part_start != -1 {
			contenido += generarTablaParticion(particiones[i])
		}
	}
	return contenido
}

func generarTablaParticion(particion Structs.Particion) string {
	partName := strings.TrimRight(string(particion.Part_name[:]), "\x00")

	tabla := "<TR>\n"
	tabla += "<TD COLSPAN = '2' bgcolor=\"#34e5dd\">Particion</TD>\n"
	tabla += "</TR>\n"
	tabla += "<TR>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">part_status</TD>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">" + string(particion.Part_status) + "</TD>\n"
	tabla += "</TR>\n"
	tabla += "<TR>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">part_type</TD>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">" + string(particion.Part_type) + "</TD>\n"
	tabla += "</TR>\n"
	tabla += "<TR>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">part_fit</TD>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">" + string(particion.Part_fit) + "</TD>\n"
	tabla += "</TR>\n"
	tabla += "<TR>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">part_start</TD>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">" + strconv.Itoa(int(particion.Part_start)) + "</TD>\n"
	tabla += "</TR>\n"
	tabla += "<TR>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">part_size</TD>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">" + strconv.Itoa(int(particion.Part_s)) + "</TD>\n"
	tabla += "</TR>\n"
	tabla += "<TR>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">part_name</TD>\n"
	tabla += "<TD bgcolor=\"#b0ebe8\">" + partName + "</TD>\n"
	tabla += "</TR>\n"
	return tabla
}

// REPORTE TREE
func tree(p string, id string) {
	var pth string
	spr := Structs.NewSuperBloque()
	inodo := Structs.NewInodos()
	partition := GetMount("REP", id, &pth)

	if partition.Part_start == -1 {
		return
	}

	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco")
		return
	}

	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &spr)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}

	file.Seek(spr.S_inode_start, 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inodo)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}

	freeI := GetFree(spr, pth, "BI")
	aux := strings.Split(strings.ReplaceAll(p, "\"", ""), ".")
	pd := aux[0] + ".dot"

	carpeta := ""
	direccion := strings.Split(pd, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(pd, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(pd)
	} else {
		fileaux.Close()
	}

	content := "digraph G{\n rankdir=LR;\n graph [ dpi = \"600\" ]; \n  forcelabels=true; \n node [shape = plaintext];\n "
	for i := 0; i < int(freeI); i++ {
		atime := arregloString(inodo.I_atime)
		ctime := arregloString(inodo.I_ctime)
		mtime := arregloString(inodo.I_mtime)
		content += " inode" + strconv.Itoa(i) + " [label = <<table>\n" +
			"<tr><td COLSPAN = '2' BGCOLOR=\"#000060\" >" +
			"<font color=\"white\"> INODO " + strconv.Itoa(i) + "</font>" +
			"</td></tr>\n" +
			"<tr><td BGCOLOR=\"#87CEFA\">NOMBRE</td><td BGCOLOR=\"#87CEFA\">VALOR</td></tr>\n" +
			"<tr><td>i_uid</td><td>" + strconv.Itoa(int(inodo.I_uid)) + "</td></tr>\n" +
			"<tr><td>i_gid</td><td>" + strconv.Itoa(int(inodo.I_gid)) + "</td></tr>\n" +
			"<tr><td>i_size</td><td>" + strconv.Itoa(int(inodo.I_s)) + "</td></tr>\n" +
			"<tr><td>i_atime</td><td>" + atime + "</td></tr>\n" +
			"<tr><td>i_ctime</td><td>" + ctime + "</td></tr>\n" +
			"<tr><td>i_mtime</td><td>" + mtime + "</td></tr>\n"
		for j := 0; j < 16; j++ {
			content += "<tr>\n<td>i_block" + strconv.Itoa(j+1) + "</td><td port=\"" + strconv.Itoa(j) + "\">" + strconv.Itoa(int(inodo.I_block[j])) + "</td></tr>\n"
		}
		content += "<tr><td>i_type</td><td>" + strconv.Itoa(int(inodo.I_type)) + "</td></tr>\n" +
			"<tr><td>i_perm</td><td>" + strconv.Itoa(int(inodo.I_perm)) + "</td></tr></table>>];\n"

		if inodo.I_type == 0 {
			for j := 0; j < 16; j++ {
				if inodo.I_block[j] != -1 {
					bloquesUsados = append(bloquesUsados, inodo.I_block[j])
					contadorBloques++
					if existeEnArreglo(bloquesUsados, inodo.I_block[j]) == 1 {
						foldertmp := Structs.BloquesCarpetas{}

						file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodo.I_block[j]*int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodo.I_block[j], 0)
						data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
						buffer = bytes.NewBuffer(data)
						err_ = binary.Read(buffer, binary.BigEndian, &foldertmp)

						if err_ != nil {
							Error("MKDIR", "Error al leer el archivo")
							return
						}

						if foldertmp.B_content[j].B_inodo == -1 {
							continue
						}
						content += "inode" + strconv.Itoa(j) + ":" + strconv.Itoa(j) + "-> BLOCK" + strconv.Itoa(contadorBloques) + "_" + strconv.Itoa(int(inodo.I_block[j])) + "\n"

						content += "BLOCK" + strconv.Itoa(contadorBloques) + "_" + strconv.Itoa(int(inodo.I_block[j])) +
							" [label = <<table><tr><td COLSPAN = '2' BGCOLOR=\"#145A32\"><font color=\"white\">BLOCK" + strconv.Itoa(int(inodo.I_block[j])) + "</font>" +
							"</td></tr>\n" + "<tr><td BGCOLOR=\"#90EE90\"> B_NAME </td><td BGCOLOR=\"#90EE90\"> B_INODE </td></tr>\n"

						for k := 0; k < 4; k++ {
							ctmp := ""
							for name := 0; name < len(foldertmp.B_content[k].B_name); name++ {
								if foldertmp.B_content[k].B_name[name] != 0 {
									ctmp += string(foldertmp.B_content[k].B_name[name])
								}
							}
							content += "<tr><td>" + ctmp + "</td><td port=\"" + strconv.Itoa(k) + "\">" + strconv.Itoa(int(foldertmp.B_content[k].B_inodo)) + "</td></tr>\n"
						}

						content += "</table>>];\n"

						for b := 0; b < 4; b++ {
							if foldertmp.B_content[b].B_inodo != -1 {
								ctmp := ""
								for name := 0; name < len(foldertmp.B_content[b].B_name); name++ {
									if foldertmp.B_content[b].B_name[name] != 0 {
										ctmp += string(foldertmp.B_content[b].B_name[name])
									}
								}
								if ctmp != "." && ctmp != ".." {
									content += "BLOCK" + strconv.Itoa(contadorBloques) + "_" + strconv.Itoa(int(inodo.I_block[j])) + ":" + strconv.Itoa(b) + "-> inode" + strconv.Itoa(b-1) + ";"
								}
							}
						}
					}
				}
			}
		} else {
			for j := 0; j < 16; j++ {
				if inodo.I_block[j] != -1 {
					if j < 16 {
						var contador int64 = 0
						var posicion int
						for {
							foldertmp := Structs.NewBloquesCarpetas()
							file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*contador*int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*contador, 0)
							data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &foldertmp)

							if err_ != nil {
								Error("REP", "No se pudo leer el archivo")
								return
							}
							salir := false
							for l := 0; l < 4; l++ {
								if foldertmp.B_content[l].B_inodo == inodo.I_block[l] {
									posicion = l
									salir = true
									break
								}
							}
							if salir {
								break
							}
							contador++
						}
						if posicion == 2 || posicion == 0 {
							file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*contador+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*contador+int64(unsafe.Sizeof(Structs.Inodos{}))*contador, 0)
							for k := 0; k < 16; k++ {
								contadorArchivos++
								if inodo.I_block[k] == -1 {
									break
								}
								content += "inodo" + strconv.Itoa(i) + ":" + strconv.Itoa(k) + " -> FILE" + strconv.Itoa(contadorArchivos) + "_" + strconv.Itoa(int(inodo.I_block[k])) + "\n"
								filetmp := Structs.BloquesArchivos{}
								data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
								buffer = bytes.NewBuffer(data)
								err_ = binary.Read(buffer, binary.BigEndian, &filetmp)
								if err_ != nil {
									Error("REP", "Error al leer el archivo")
									return
								}

								contenido := ""
								for arch := 0; arch < len(filetmp.B_content); arch++ {
									if filetmp.B_content[arch] != 0 {
										contenido += string(filetmp.B_content[arch])
									}
								}
								content += "FILE" + strconv.Itoa(contadorArchivos) + "_" + strconv.Itoa(int(inodo.I_block[k])) + " [label = <<table> <tr><td COLSPAN = '2' BGCOLOR=\"#CCCC00\"> FILE" + strconv.Itoa(int(inodo.I_block[k])) +
									"</td></tr><tr><td COLSPAN ='2'>" + contenido + "</td></tr>\n</table>>];\n"
							}
						} else if posicion == 1 {
							file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*contador*int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*contador*int64(unsafe.Sizeof(inodo.I_block[j])), 0)
							for k := 0; k < 16; k++ {
								contadorArchivos++
								if inodo.I_block[k] == -1 {
									break
								}
								content += "inodo" + strconv.Itoa(k) + ":" + strconv.Itoa(k) + "-> FILE" + strconv.Itoa(contadorArchivos) + "_" + strconv.Itoa(int(inodo.I_block[k])) + "\n"
								filetmp := Structs.BloquesArchivos{}
								datas := leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
								buffer = bytes.NewBuffer(datas)
								err_ = binary.Read(buffer, binary.BigEndian, &filetmp)
								if err_ != nil {
									Error("REP", "No se pudo abrir el archivo")
									return
								}

								contenido := ""
								for arch := 0; arch < len(filetmp.B_content); arch++ {
									if filetmp.B_content[arch] != 0 {
										contenido += string(filetmp.B_content[arch])
									}
								}
								content += "FILE" + strconv.Itoa(contadorArchivos) + "_" + strconv.Itoa(int(inodo.I_block[k])) + " [ label = <<table> <tr><td COLSPAN = '2' BGCOLOR=\"#CCCC00\">" +
									"FILE " + strconv.Itoa(int(inodo.I_block[k])) + "</td></tr>\n" + "<tr><td COLSPAN = '2'>" + contenido + "</td></tr>\n</table>>];\n"

							}
						}
						break
					}
				} else {
					break
				}
			}
		}
		file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*int64(i+1), 0)
		data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &inodo)
		if err_ != nil {
			Error("REP", "Error al leer el archivo")
			return
		}
	}
	file.Close()
	content += "\n\n}\n"

	//fmt.Println(content)

	b := []byte(content)
	err_ = ioutil.WriteFile(pd, b, 0644)
	if err_ != nil {
		log.Fatal(err_)
	}

	terminacion := strings.Split(p, ".")

	path, _ := exec.LookPath("dot")
	cmd, _ := exec.Command(path, "-T"+terminacion[1], pd).Output()
	node := int(0777)
	ioutil.WriteFile(p, cmd, os.FileMode(node))
	Mensaje("REP", "Reporte tipo TREE para el filesystem de la particion "+id+",creado correctamente")

}

func inoder(p string, id string) {

}

func blockr(p string, id string) {

}

// REPORTE JOURNALING
func journalr(p string, id string) {
	var pth string
	partition := GetMount("REP", id, &pth)

	//file
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return
	}
	journal := Structs.NewJournaling()
	file.Seek(partition.Part_start+int64(unsafe.Sizeof(Structs.SuperBloque{})), 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.Journaling{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &journal)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}
	file.Close()

	aux := strings.Split(p, ".")
	if len(aux) > 2 {
		Error("REP", "No se admiten nombres de archivos que contengan punto (.)")
		return
	}
	pd := aux[0] + ".dot"

	carpeta := ""
	direccion := strings.Split(pd, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(pd, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(pd)
	} else {
		fileaux.Close()
	}

	content := "digraph G {\n  node0 [shape=none label=<\n  <TABLE style=\"rounded\" bgcolor=\" #d5f2e9 \">\n  <TR>\n  <TD COLSPAN = '4' bgcolor=\"  #badbb8 \">REPORTE DE JOURNALING</TD>\n  </TR>\n" +
		"<TR>\n<TD bgcolor=\"  #dceddb  \">operacion</TD>\n  <TD bgcolor=\"  #dceddb \">path</TD>\n    <TD bgcolor=\"  #dceddb \">contenido</TD>\n      <TD bgcolor=\"  #dceddb \">fecha</TD>\n  </TR>\n " +
		" <TR>\n<TD bgcolor=\"  #dceddb  \"> " + string(journal.J_operacion[:]) + "</TD>\n  <TD bgcolor=\"  #dceddb \">" + string(journal.J_path[:]) + "</TD>\n    <TD bgcolor=\"  #dceddb \">" + string(journal.J_content[:]) + "</TD>\n      <TD bgcolor=\"  #dceddb \">" + string(journal.J_fecha[:]) + "</TD>\n  </TR>\n  \n    </TABLE>>];\n\n}\n"

	//fmt.Println(content)

	//CREAR IMAGEN
	b := []byte(content)
	err_ = ioutil.WriteFile(pd, b, 0644)
	if err_ != nil {
		log.Fatal(err_)
	}

	terminacion := strings.Split(p, ".")

	path, _ := exec.LookPath("dot")
	cmd, _ := exec.Command(path, "-T"+terminacion[1], pd).Output()
	node := int(0777)
	ioutil.WriteFile(p, cmd, os.FileMode(node))
	Mensaje("REP", "Reporte tipo JOURNALING para la particion "+id+", creado correctamente")

}

// REPORTE SUPERBLOQUE
func superblockr(p string, id string) {
	var pth string
	partition := GetMount("REP", id, &pth)

	//file
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return
	}
	super := Structs.NewSuperBloque()
	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &super)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}
	file.Close()

	aux := strings.Split(p, ".")
	if len(aux) > 2 {
		Error("REP", "No se admiten nombres de archivos que contengan punto (.)")
		return
	}
	pd := aux[0] + ".dot"

	carpeta := ""
	direccion := strings.Split(pd, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(pd, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(pd)
	} else {
		fileaux.Close()
	}
	ultfecha := strings.TrimRight(string(super.S_umtime[:]), "\x00")
	//mbrTamanoStr := strconv.FormatFloat(float64(disk.Mbr_tamano), 'f', -1, 64)
	content := "digraph G {\n  node0 [shape=none label=<\n  <TABLE style=\"rounded\" bgcolor=\" #d5f2e9 \">\n  <TR>\n  <TD COLSPAN = '2' bgcolor=\"  #9bf597 \">REPORTE DE SUPERBLOQUE</TD>\n  </TR>\n  <TR>\n" +
		"<TD bgcolor=\" #cff5e5 \">spr_filesystem_type</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_filesystem_type)) + "</TD>\n  </TR>\n  <TR>\n " +
		"<TD bgcolor=\" #cff5e5 \">spr_inodes_count</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_inodes_count)) + "</TD>\n  </TR>\n  <TR>\n" +
		"<TD bgcolor=\" #cff5e5 \">spr_blocks_count</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_blocks_count)) + "</TD>\n  </TR>\n  <TR>\n  " +
		"<TD bgcolor=\" #cff5e5 \">spr_free_inodes_count</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_free_blocks_count)) + "</TD>\n  </TR>\n" +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_free_blocks_count</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_free_blocks_count)) + "</TD>\n  </TR>\n " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_mtime</TD>\n  <TD bgcolor=\" #cff5e5 \">" + string(super.S_mtime[:]) + "</TD>\n  </TR>\n" +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">sptr_umtime</TD>\n  <TD bgcolor=\" #cff5e5 \">" + ultfecha + "</TD>\n  </TR>\n " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_mnt_count</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_mnt_count)) + "</TD>\n  </TR>\n " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_magic</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_magic)) + "</TD>\n  </TR>\n" +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_inode_s</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_inode_s)) + "</TD>\n  </TR>\n " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_block_s</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_block_s)) + "</TD>\n  </TR>\n  " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_first_ino</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_first_ino)) + "</TD>\n  </TR>\n   " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_first_blo</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_first_blo)) + "</TD>\n  </TR>\n    " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_bm_inode_start</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_bm_inode_start)) + "</TD>\n  </TR>\n  " +
		" <TR>\n  <TD bgcolor=\" #cff5e5 \">spr_bm_block_start</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_bm_block_start)) + "</TD>\n  </TR>\n   " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_inode_start</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_inode_start)) + "</TD>\n  </TR>\n  " +
		"<TR>\n  <TD bgcolor=\" #cff5e5 \">spr_block_start</TD>\n  <TD bgcolor=\" #cff5e5 \">" + strconv.Itoa(int(super.S_block_start)) + "</TD>\n  </TR>"

	content += "</TABLE>>];\n\n}"

	//fmt.Println(content)

	//CREAR IMAGEN
	b := []byte(content)
	err_ = ioutil.WriteFile(pd, b, 0644)
	if err_ != nil {
		log.Fatal(err_)
	}

	terminacion := strings.Split(p, ".")

	path, _ := exec.LookPath("dot")
	cmd, _ := exec.Command(path, "-T"+terminacion[1], pd).Output()
	node := int(0777)
	ioutil.WriteFile(p, cmd, os.FileMode(node))
	Mensaje("REP", "Reporte tipo SUPERBLOQUE para la particion "+id+", creado correctamente")

}

func bminoder(p string, id string) {

}

func bmblockr(p string, id string) {

}

func arregloString(arreglo [16]byte) string {
	reg := ""
	for i := 0; i < 16; i++ {
		if arreglo[i] != 0 {
			reg += string(arreglo[i])
		}
	}
	return reg
}

func existeEnArreglo(arreglo []int64, busqueda int64) int {
	regresa := 0
	for _, numero := range arreglo {
		if numero == busqueda {
			regresa++
		}
	}
	return regresa
}

func fileR(p string, id string, ruta string) {
	carpeta := ""
	direccion := strings.Split(p, "/")

	fileaux, _ := os.Open(strings.ReplaceAll(p, "\"", ""))
	if fileaux == nil {
		for i := 0; i < len(direccion); i++ {
			carpeta += "/" + direccion[i]
			if _, err_2 := os.Stat(carpeta); os.IsNotExist(err_2) {
				os.Mkdir(carpeta, 0777)
			}
		}
		os.Remove(p)
	} else {
		fileaux.Close()
	}

	var path string
	particion := GetMount("MKDIR", id, &path)
	tmp := GetPath(ruta)
	data := getDataFile(tmp, particion, path)
	b := []byte(data)
	err_ := ioutil.WriteFile(p, b, 0644)
	if err_ != nil {
		log.Fatal(err_)
	}

	archivo := strings.Split(ruta, "/")
	Mensaje("REP", "Reporte tipo FILE del archivo"+archivo[len(archivo)-1]+"creado correctamente!")

}

func GetFree(spr Structs.SuperBloque, pth string, t string) int64 {
	ch := '2'
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco")
		return -1
	}
	if t == "BI" {
		file.Seek(spr.S_bm_inode_start, 0)
		for i := 0; i < int(spr.S_inodes_count); i++ {
			data := leerBytes(file, int(unsafe.Sizeof(ch)))
			buffer := bytes.NewBuffer(data)
			err_ := binary.Read(buffer, binary.BigEndian, &ch)
			if err_ != nil {
				Error("MKDIR", "Error al leer el archivo")
				return -1
			}
			if ch == '0' {
				file.Close()
				return int64(i)
			}
		}
	} else {
		file.Seek(spr.S_bm_block_start, 0)
		for i := 0; i < int(spr.S_blocks_count); i++ {
			data := leerBytes(file, int(unsafe.Sizeof(ch)))
			buffer := bytes.NewBuffer(data)
			err_ := binary.Read(buffer, binary.BigEndian, &ch)
			if err_ != nil {
				Error("MKDIR", "Error al leer el archivo")
				return -1
			}
			if ch == '0' {
				file.Close()
				return int64(i)
			}
		}
	}
	file.Close()
	return -1
}

func GetPath(path string) []string {
	var result []string
	if path == "" {
		return result
	}
	aux := strings.Split(path, "/")
	for i := 1; i < len(aux); i++ {
		result = append(result, aux[i])
	}
	return result
}

func getDataFile(path []string, particion Structs.Particion, pth string) string {
	spr := Structs.NewSuperBloque()
	inodo := Structs.NewInodos()
	folder := Structs.NewBloquesCarpetas()
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return ""
	}
	file.Seek(particion.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &spr)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return ""
	}

	file.Seek(spr.S_inode_start, 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inodo)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return ""
	}

	if len(path) == 0 {
		Error("REP", "No se ha encontrar una ruta o path valido ")
		return ""
	}

	var aux []string
	for i := 0; i < len(path); i++ {
		aux = append(aux, path[i])
	}
	path = aux

	for v := 0; v < len(path); v++ {
		for i := 0; i < 16; i++ {
			if i < 16 {
				if inodo.I_block[i] != -1 {
					folderc := Structs.BloquesCarpetas{}
					file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodo.I_block[i]*int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodo.I_block[i], 0)

					data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &folderc)
					if err_ != nil {
						Error("REP", "Error al abrir el archivo")
					}

					for j := 0; j < 4; j++ {
						nombreFIle := ""
						for name := 0; name < len(folder.B_content[j].B_name); name++ {
							if folder.B_content[j].B_name[name] == 0 {
								continue
							}
							nombreFIle += string(folder.B_content[j].B_name[name])
						}
						if Comparar(nombreFIle, path[v]) {
							inodeAux := inodo
							inodo = Structs.NewInodos()
							file.Seek(spr.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*folder.B_content[j].B_inodo, 0)

							data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &inodo)
							if err_ != nil {
								Error("REP", "Error al leer el archivo")
								return ""
							}

							if inodo.I_type == 1 && nombreFIle == path[len(path)-1] {
								if j == 2 {
									archivo := Structs.BloquesArchivos{}
									contenido := ""
									for k := 0; k < 16; k++ {
										file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodo.I_block[i], 0)
										if inodo.I_block[k] == -1 {
											break
										}
										data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
										buffer = bytes.NewBuffer(data)
										err_ = binary.Read(buffer, binary.BigEndian, &archivo)

										for l := 0; l < len(archivo.B_content[:]); l++ {
											if archivo.B_content[l] != 0 {
												contenido += string(archivo.B_content[l])
											}
										}
									}

									if nombreFIle == path[len(path)-1] {
										return contenido
									}

								} else if j == 3 {
									archivo := Structs.BloquesArchivos{}
									contenido := ""
									for k := 0; k < 16; k++ {
										file.Seek(spr.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inodeAux.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inodo.I_block[i], 0)
										if inodo.I_block[k] == -1 {
											break
										}
										data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
										buffer = bytes.NewBuffer(data)
										err_ = binary.Read(buffer, binary.BigEndian, &archivo)

										for l := 0; l < len(archivo.B_content); l++ {
											if archivo.B_content[l] != 0 {
												contenido += string(archivo.B_content[l])
											}
										}
									}
									if nombreFIle == path[len(path)-1] {
										return contenido
									}
								}
							}
							break
						}
					}
				} else {
					break
				}
			}
		}
	}
	return ""
}
