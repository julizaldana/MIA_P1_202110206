package Comandos

import (
	"MIA_P1_202110206/Structs"
	"bytes"
	"encoding/binary"
	"os"
	"strings"
	"time"
	"unsafe"
)

func ValidarDatosMKDIR(context []string, particion Structs.Particion, pth string) {
	path := ""
	p := false
	//fs := "" //Verificar si es ext2 o ext3
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "path") {
			path = comando[1]
		} else if Comparar(comando[0], "p") {
			p = true
		}
	}
	if path == "" {
		Error("MKDIR", "El comando MKDIR requiere el path o ruta para poder crear un directorio.")
		return
	}
	tmp := GetPath(path)
	mkdir(tmp, p, particion, pth)

}

//Yo obtengo /home/usac, tiene que ir de izquierda a derecha, primero verificar si existe home sino crearlo. Y finalmente ir hasta el ultimo, verificar si existe, y sino crearlo.

//Meto Path como parametro, ej: mkdir -r -path=/home
//Crear una carpeta implica crear un objeto inodo y un objeto bloque carpeta.

func crearCarpetaRaiz(path string, r string) {
	var p string
	partition := GetMount("MKDIR", Logged.Id, &p)
	if string(partition.Part_status) == "0" {
		Error("MKDIR", "No se encontró la partición montada con el id: "+Logged.Id)
		return
	}

	file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco.")
		return
	}
	//Se lee el superbloque
	super := Structs.NewSuperBloque()
	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &super)
	if err_ != nil {
		Error("MKDIR", "Error al leer el archivo")
		return
	}
	//Se lee el inodo actual del archivo
	inode := Structs.NewInodos()
	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("MKDIR", "Error al leer el archivo")
		return
	}
	//Se lee el bloque de carpetas actual del archivo
	bloquecarpeta := Structs.NewBloquesCarpetas()
	file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{})), 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &bloquecarpeta)
	if err_ != nil {
		Error("MKDIR", "Error al leer el archivo")
		return
	}
	//Se lee el bitmap inodos, debo de actualizarlo tambien

	//Se lee el bitmap bloques, debo de actualizarlo tambien

	// /usac -> Se debe de crear en la posicion B_content[3]
	//El bloque de carpetas actual, debo de almacenar
	copy(bloquecarpeta.B_content[3].B_name[:], strings.Split(path, "/"))
	bloquecarpeta.B_content[3].B_inodo = 2

	//POR NUEVA CARPETA CREAR UN NUEVO INODO Y UN NUEVO BLOQUE DE CARPETAS

	//Nuevo Inodo
	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_s = 0
	fecha := time.Now().String()
	copy(inode.I_atime[:], fecha)
	copy(inode.I_ctime[:], fecha)
	copy(inode.I_mtime[:], fecha)
	inode.I_type = 0
	inode.I_perm = 664
	inode.I_block[0] = 3

	//Nuevo bloque de carpetas
	copy(bloquecarpeta.B_content[0].B_name[:], ".")
	bloquecarpeta.B_content[0].B_inodo = 0
	copy(bloquecarpeta.B_content[1].B_name[:], "..")
	bloquecarpeta.B_content[1].B_inodo = 0
	copy(bloquecarpeta.B_content[2].B_name[:], "--")
	bloquecarpeta.B_content[2].B_inodo = -1
	copy(bloquecarpeta.B_content[3].B_name[:], "--")
	bloquecarpeta.B_content[3].B_inodo = -1

	//Se vuelve a reescribir los cambios de inodos
	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)

	var bin3 bytes.Buffer
	binary.Write(&bin3, binary.BigEndian, inode)
	EscribirBytes(file, bin3.Bytes())

	//Se vuelven a reescribir los cambios para bloques de carpetas
	file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{})), 0)

	var bin5 bytes.Buffer
	binary.Write(&bin5, binary.BigEndian, bloquecarpeta)
	EscribirBytes(file, bin5.Bytes())

}
