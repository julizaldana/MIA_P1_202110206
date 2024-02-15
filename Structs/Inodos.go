package Structs

type Inodos struct {
	I_uid   int64
	I_gid   int64
	I_s     int64 //SIZE
	I_atime [16]byte
	I_ctime [16]byte
	I_mtime [16]byte
	I_block [16]int64
	I_type  int64
	I_perm  int64
}

func NewInodos() Inodos {
	var inode Inodos
	inode.I_uid = -1
	inode.I_gid = -1
	inode.I_s = -1 //SIZE
	for i := 0; i < 16; i++ {
		inode.I_block[i] = -1
	}
	inode.I_type = -1
	inode.I_perm = -1
	return inode
}
