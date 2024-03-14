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

	content := ""
	content = "digraph G{\n rankdir=TB;\n forcelabels= true;\n graph { dpi = \"600\"] ; \n node [shape = plaintext];\n nodo 1 {label = <<table>\n <tr>\n"
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
				tmplogic += "<td>\"Libre\n " + s + "% libre de la partición extendida\"</td>\n"
				logic++
			}
			logic += 2 //Son los id, para los nodos de graphviz

		}
		var tamPrim int64
		for i := 0; i < 4; i++ {
			if partitions[i].Part_type == 'E' {
				tamPrim += partitions[i].Part_s
				res := float64(partitions[i].Part_s) / float64(disk.Mbr_tamano)
				res = math.Round(res*10000.00) / 100.00
				s := fmt.Sprintf("%.2f", res)
				content += "<td COLSPAN=" + strconv.Itoa(logic) + "> Extendida \n" + s + "% del disco </td>\n"
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

		fmt.Println(content)

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

}

func tree(p string, id string) {
	var pth string
	spr := Structs.NewSuperBloque()
	inode := Structs.NewInodos()
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
	err_ = binary.Read(buffer, binary.BigEndian, &spr)
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

	content := "digraph G{\n rankdir=LR;\n graph [ dpi = \'608\' ]; \n  forcelabels=true; \n node { shape = plaintext};\n "
	for i := 0; i < int(freeI); i++ {
		atime := arregloString(inode.I_atime)
		ctime := arregloString(inode.I_ctime)
		mtime := arregloString(inode.I_mtime)
		content += " inode" + strconv.Itoa(i) + " [label = <<table>\n" +
			"<tr><td COLSPAN = '2'>" +
			"<font color=\"white\"> INODO " + strconv.Itoa(i) + "</font>" +
			"</td></tr>\n" +
			"<tr><td>NOMBRE</td><td>VALOR</td></tr>\n" +
			"<tr><td>i_uid</td><td>" + strconv.Itoa(int(inode.I_uid)) + "</td></tr>\n" +
			"<tr><td>i_gid</td><td>" + strconv.Itoa(int(inode.I_gid)) + "</td></tr>\n" +
			"<tr><td>i_size</td><td>" + strconv.Itoa(int(inode.I_s)) + "</td></tr>\n" +
			"<tr><td>i_atime</td>" + atime + "<td>VALOR</td></tr>\n" +
			"<tr><td>i_ctime</td>" + ctime + "<td>VALOR</td></tr>\n" +
			"<tr><td>i_mtime</td>" + mtime + "<td>VALOR</td></tr>\n"
		for j := 0; j < 16; j++ {
			content += "<tr=\n<td>i_block " + strconv.Itoa(j+1) + "</td><td port=\"" + strconv.Itoa(j) + "\">" + strconv.Itoa(int(inode.I_block[j])) + "</td></tr>\n"
		}
		content += "<tr><td>i_type</td><td>" + strconv.Itoa(int(inode.I_type)) + "</td></tr>\n" +
			"<tr><td>i_perm</td><td>" + strconv.Itoa(int(inode.I_perm)) + "</td></tr>"
	}

	if inode.I_type == 0 {
		for j := 0;
	}

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

func getDataFile(path []string, particion Structs.Particion, pth string) {
	spr := Structs.NewSuperBloque()
	inode := Structs.NewInodos()
	folder := Structs.NewBloquesCarpetas()
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("REP", "No se ha encontrado el disco.")
		return
	}
	file.Seek(particion.Part_start, 0)
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
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("REP", "Error al leer el archivo")
		return
	}

	if (path) == 0 {
		Error("REP", "No se ha ")
	}

	var dua []string
	for i := 0; i < len[path]; i++ {
		aup
	}

}
