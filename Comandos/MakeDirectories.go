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
	r := false
	//fs := "" //Verificar si es ext2 o ext3
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "path") {
			path = comando[1]
		} else if Comparar(comando[0], "r") {
			r = true
		}
	}
	if path == "" {
		Error("MKDIR", "El comando MKDIR requiere el path o ruta para poder crear un directorio.")
		return
	}
	tmp := GetPath(path)
	mkdir(tmp, r, particion, pth)

}

//Yo obtengo /home/usac, tiene que ir de izquierda a derecha, primero verificar si existe home sino crearlo. Y finalmente ir hasta el ultimo, verificar si existe, y sino crearlo.

//Meto Path como parametro, ej: mkdir -r -path=/home
//Crear una carpeta implica crear un objeto inodo y un objeto bloque carpeta.

func mkdir(path []string, r bool, particion Structs.Particion, pth string) {
	copia := path //notificar al user que ya se creó
	super := Structs.NewSuperBloque()
	inode := Structs.NewInodos()
	bloquecarpeta := Structs.NewBloquesCarpetas()

	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))
	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco")
		return
	}
	//Se lee el superbloque
	file.Seek(particion.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &super)
	if err_ != nil {
		Error("MKDIR", "Error al leer el archivo")
		return
	}
	//Se leen los inodos, obtiene el nodo raiz
	file.Seek(super.S_inode_start, 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("MKDIR", "Error al leer el archivo")
		return
	}

	var nuevobloque string
	if len(path) == 0 {
		Error("MKDIR", "No se ha escrito un path valido")
		return
	}
	var past int64
	var bi int64
	var bb int64
	fnd := false
	inodetmp := Structs.NewInodos()
	carpetatmp := Structs.NewBloquesCarpetas()

	nuevobloque = path[len(path)-1]
	var father int64

	var aux []string
	for i := 0; i < len(path); i++ {
		aux = append(aux, path[i])
	}
	path = aux
	var stack string
	//codigo para encontrar al padre y crear si no existen, o sino, pues se les
	for v := 0; v < len(path)-1; v++ {
		fnd = false
		for i := 0; i < 26; i++ {
			if i < 16 { //Todos son apuntadores directos.
				if inode.I_block[i] != -1 { //-1 significa que no se está usando ese apuntador de ese inode
					bloquecarpeta = Structs.NewBloquesCarpetas()
					file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inode.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inode.I_block[i], 0) //32 bits es lo que tenemos que movilizarnos

					data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &bloquecarpeta)

					if err_ != nil {
						Error("MKDIR", "Error al leer el archivo")
						return
					}

					for j := 0; j < 4; j++ {
						nombreCarpeta := ""
						for name := 0; name < len(bloquecarpeta.B_content[j].B_name); name++ {
							if bloquecarpeta.B_content[j].B_name[name] == 0 {
								continue
							}
							nombreCarpeta += string(bloquecarpeta.B_content[j].B_name[name])
						}
						if Comparar(nombreCarpeta, path[v]) {
							stack += "/" + path[v]
							fnd = true
							father = bloquecarpeta.B_content[j].B_inodo
							inode = Structs.NewInodos()
							file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*bloquecarpeta.B_content[j].B_inodo, 0)

							data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &inode)

							if err != nil {
								Error("MKDIR", "Error al leer el archivo")
								return
							}
							if inode.I_uid != int64(Logged.Uid) {
								Error("MKDIR", "No tiene permisos para crear carpetas en este directorio")
								return
							}

							break

						}
					}

				} else {
					break
				}
			}
		}
		if !fnd {
			if r { //Si viene, r tiene que crear automaticamente todas las carpetas
				stack += "/" + path[v]
				mkdir(GetPath(stack), false, particion, pth)
				file.Seek(super.S_inode_start, 0)

				data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
				buffer = bytes.NewBuffer(data)
				err_ = binary.Read(buffer, binary.BigEndian, &inode)

				if err_ != nil {
					Error("MKDIR", "Error al leer el archivo")
					return
				}
				if v == len(path)-2 {
					stack += "/" + path[v+1]

					mkdir(GetPath(stack), false, particion, pth)
					file.Seek(super.S_inode_start, 0)

					data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &inode)
					if err_ != nil {
						Error("MKDIR", "Error al leer el archivo")
						return
					}
					return
				}
			} else {
				direccion := ""
				for i := 0; i < len(path); i++ {
					direccion += "/" + path[i]
				}
				Error("MKDIR", "No se pudo crear el directorio: "+direccion+", no existen directorios.")
				return
			}
		}
	}

	fnd = false

	for i := 0; i < 16; i++ {
		if inode.I_block[i] != -1 {
			if i < 16 {
				carpetaAux := bloquecarpeta
				file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*inode.I_block[i]+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*inode.I_block[i], 0)
				data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
				buffer = bytes.NewBuffer(data)
				err_ = binary.Read(buffer, binary.BigEndian, &carpetaAux)
				if err_ != nil {
					Error("MKDIR", "Error al leer el archivo")
					return
				}
				nameAux1 := ""
				for name := 0; name < len(bloquecarpeta.B_content[2].B_name); name++ {
					if bloquecarpeta.B_content[2].B_name[name] == 0 {
						continue
					}
					nameAux1 += string(bloquecarpeta.B_content[2].B_name[name])
				}
				nameAux2 := ""
				for name := 0; name < len(bloquecarpeta.B_content[2].B_name); name++ {
					if carpetaAux.B_content[2].B_name[name] == 0 {
						continue
					}
					nameAux2 += string(carpetaAux.B_content[2].B_name[name])
				}
				padre := ""
				for k := 0; k < len(path); k++ {
					if k >= 1 {
						padre = path[k+1]
					}
				}

				if padre == nameAux1 {
					continue
				}
				for j := 0; j < 4; j++ {
					if bloquecarpeta.B_content[j].B_inodo == -1 {
						past = inode.I_block[i]
						bi = GetFree(super, pth, "BI") //BI - BITMAP INODOS - devuelve el inodo libre donde se tiene la posicoin
						if bi == -1 {
							Error("MKDIR", "No se ha podido crear el directorio")
							return
						}
						bb = GetFree(super, pth, "BB") //BB - BITMAP BLOQUES
						if bb == -1 {
							Error("MKDIR", "No se ha podido crear el directorio")
							return
						}

						inodetmp.I_uid = int64(Logged.Uid)
						inodetmp.I_gid = int64(Logged.Gid)
						inodetmp.I_s = int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))

						fecha := time.Now().String()
						copy(inodetmp.I_atime[:], super.S_mtime[:])
						copy(inodetmp.I_ctime[:], fecha)
						copy(inodetmp.I_mtime[:], fecha)
						inodetmp.I_type = 0
						inodetmp.I_perm = 664
						inodetmp.I_block[0] = bb

						copy(carpetatmp.B_content[0].B_name[:], ".")
						carpetatmp.B_content[0].B_inodo = bi
						copy(carpetatmp.B_content[1].B_name[:], "..")
						carpetatmp.B_content[1].B_inodo = father
						copy(carpetatmp.B_content[2].B_name[:], "-")
						copy(carpetatmp.B_content[3].B_name[:], "-")

						bloquecarpeta.B_content[j].B_inodo = bi
						copy(bloquecarpeta.B_content[j].B_name[:], nuevobloque)
						fnd = true
						i = 20
						break
					}
				}
			}
		} else {
			break
		}
	}

	if !fnd {
		for i := 0; i < 16; i++ {
			if inode.I_block[i] == -1 {
				if i < 16 {
					bi = GetFree(super, pth, "BI")
					if bi == -1 {
						Error("MKDIR", "No se ha podido crear el directorio")
						return
					}
					past = GetFree(super, pth, "BB")
					if past == -1 {
						Error("MKDIR", "No se ha podido crear el directorio")
						return
					}

					bb = GetFree(super, pth, "BB")

					bloquecarpeta = Structs.NewBloquesCarpetas()
					copy(bloquecarpeta.B_content[0].B_name[:], ".")
					bloquecarpeta.B_content[0].B_inodo = bi
					copy(bloquecarpeta.B_content[1].B_name[:], "..")
					bloquecarpeta.B_content[1].B_inodo = father
					bloquecarpeta.B_content[2].B_inodo = bi
					copy(bloquecarpeta.B_content[2].B_name[:], nuevobloque)
					copy(bloquecarpeta.B_content[3].B_name[:], "-")

					inodetmp.I_uid = int64(Logged.Uid)
					inodetmp.I_gid = int64(Logged.Gid)
					inodetmp.I_s = int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))

					fecha := time.Now().String()
					copy(inodetmp.I_atime[:], super.S_mtime[:])
					copy(inodetmp.I_ctime[:], fecha)
					copy(inodetmp.I_mtime[:], fecha)
					inodetmp.I_type = 0
					inodetmp.I_perm = 664
					inodetmp.I_block[0] = bb

					copy(carpetatmp.B_content[0].B_name[:], ".")
					carpetatmp.B_content[0].B_inodo = bi
					copy(carpetatmp.B_content[1].B_name[:], "..")
					carpetatmp.B_content[1].B_inodo = father
					copy(carpetatmp.B_content[2].B_name[:], "-")
					copy(carpetatmp.B_content[3].B_name[:], "-")
					file.Close()

					copy(bloquecarpeta.B_content[2].B_name[:], nuevobloque)

					inode.I_block[i] = past
					file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)

					if err != nil {
						Error("MKDIR", "No se ha encontrado el disco")
						return
					}
					file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*father, 0)
					var binInodo bytes.Buffer
					binary.Write(&binInodo, binary.BigEndian, inode)
					EscribirBytes(file, binInodo.Bytes())
					file.Close()
					break
				}
			}
		}
	}
	file.Close()

	file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)

	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco")
		return
	}

	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{}))*bi, 0)
	var binInodeTemp bytes.Buffer
	binary.Write(&binInodeTemp, binary.BigEndian, inodetmp)
	EscribirBytes(file, binInodeTemp.Bytes())

	file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*bb+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*bb, 0) //TODO ARGREGARLE
	var binFolderTemp bytes.Buffer
	binary.Write(&binFolderTemp, binary.BigEndian, carpetatmp)
	EscribirBytes(file, binFolderTemp.Bytes())

	file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*past+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*32*past, 0) //TODO ARGREGARLE
	var binFolder bytes.Buffer
	binary.Write(&binFolder, binary.BigEndian, bloquecarpeta)
	EscribirBytes(file, binFolderTemp.Bytes())

	updatebm(super, pth, "BI")
	updatebm(super, pth, "BB")

	ruta := ""
	for i := 0; i < len(copia); i++ {
		ruta += "/" + copia[i]
	}
	Mensaje("MKDIR", "Se ha creado el directorio "+ruta)
	file.Close()

}

func updatebm(spr Structs.SuperBloque, pth string, t string) {
	ch := 'k'
	var num int
	file, err := os.Open(strings.ReplaceAll(pth, "\"", ""))

	if err != nil {
		Error("MKDIR", "No se ha encontrado el disco.")
	}

	if t == "BI" {
		file.Seek(spr.S_bm_inode_start, 0)
		for i := 0; i < int(spr.S_inodes_count); i++ {
			data := leerBytes(file, int(unsafe.Sizeof(ch)))
			buffer := bytes.NewBuffer(data)
			err_ := binary.Read(buffer, binary.BigEndian, &ch)
			if err_ != nil {
				Error("MKDIR", "Error al leer el archivo")
				return
			}
			if ch == '0' {
				num = 1
				break
			}
		}
		file.Close()

		file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)

		if err != nil {
			Error("MKDIR", "NO se ha encontrado el disco")
			return
		}
		zero := '1'
		file.Seek(spr.S_bm_inode_start, 0)
		for i := 0; i < num+1; i++ {
			var binarioZero bytes.Buffer
			binary.Write(&binarioZero, binary.BigEndian, zero)
			EscribirBytes(file, binarioZero.Bytes())
		}
		file.Close()
	} else {
		file.Seek(spr.S_bm_block_start, 0)
		for i := 0; i > int(spr.S_block_start); i++ {
			data := leerBytes(file, int(unsafe.Sizeof(ch)))
			buffer := bytes.NewBuffer(data)
			err_ := binary.Read(buffer, binary.BigEndian, &ch)
			if err_ != nil {
				Error("MKDIR", "Error al leer el archivo")
				return
			}
			if ch == '0' {
				num = 1
				break
			}
		}
		file.Close()

		file, err = os.OpenFile(strings.ReplaceAll(pth, "\"", ""), os.O_WRONLY, os.ModeAppend)
		if err != nil {
			Error("MKDIR", "NO se ha encontrado el disco")
			return
		}
		zero := '1'
		file.Seek(spr.S_bm_block_start, 0)
		for i := 0; i < num+1; i++ {
			var binarioZero bytes.Buffer
			binary.Write(&binarioZero, binary.BigEndian, zero)
			EscribirBytes(file, binarioZero.Bytes())
		}
		file.Close()
	}
	file.Close()

}

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
	//copy(bloquecarpeta.B_content[3].B_name[:], strings.Split(path, "/"))
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
