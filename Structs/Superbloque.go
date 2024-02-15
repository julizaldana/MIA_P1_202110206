package Structs

import "unsafe"

//Contiene la informacion principal del sistema de archivos, ya sea EXT2/EXT3.

type SuperBloque struct {
	S_filesystem_type   int64
	S_inodes_count      int64
	S_blocks_count      int64
	S_free_blocks_count int64
	S_free_inodes_count int64
	S_mtime             [16]byte
	S_umtime            [16]byte
	S_mnt_count         int64
	S_magic             int64
	S_inode_s           int64
	S_block_s           int64
	S_first_ino         int64
	S_first_blo         int64
	S_bm_inode_start    int64
	S_bm_block_start    int64
	S_inode_start       int64
	S_block_start       int64
}

func NewSuperBloque() SuperBloque {
	var spr SuperBloque
	spr.S_magic = 0xEF53
	spr.S_inode_s = int64(unsafe.Sizeof(Inodos{}))
	spr.S_block_s = int64(unsafe.Sizeof(BloquesCarpetas{}))
	spr.S_first_ino = 0
	spr.S_first_blo = 0
	return spr
}
