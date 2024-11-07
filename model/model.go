package model

type Node struct {
	Id            string  `json:"id"`            //Unique identifier of the node. This ID is referenced by edge in its source and target field.
	Title         string  `json:"title"`         //Name of the node visible in just under the node.
	MainStat      float32 `json:"mainStat"`      //First stat shown inside the node itself
	SecondaryStat float32 `json:"secondaryStat"` //Same as mainStat, but shown under it inside the node
	Arc__Failed   float32 `json:"arc__failed"`   //to create the color circle around the node. All values in these fields should add up to 1.
	Arc__Passed   float32 `json:"arc__passed"`   //
	Detail__Role  string  `json:"detail__role"`  //shown in the header of context menu when clicked on the node
	Color         string  `json:"color"`         //Can be used to specify a single color instead of using the arc__ fields to specify color sections
	Icon          string  `json:"icon"`          //
	NodeRadius    int     `json:"nodeRadius"`    //Radius value in pixels. Used to manage node size.
	Highlighted   bool    `json:"highlighted"`   //Sets whether the node should be highlighted.
}

type Edge struct {
	Id            string  `json:"id"`            //Unique identifier of the edge.
	Source        string  `json:"source"`        //Id of the source node.
	Target        string  `json:"target"`        //Id of the target.
	MainStat      float32 `json:"mainStat"`      //First stat shown in the overlay when hovering over the edge.
	SecondaryStar float32 `json:"secondarystat"` //Same as mainStat, but shown right under it.
	Detail__Info  string  `json:"detail__info"`  //will be shown in the header of context menu when clicked on the edge
	Thickness     float32 `json:"thickness"`     //The thickness of the edge. Default: 1
	Highlighted   bool    `json:"highlighted"`   //boolean	Sets whether the edge should be highlighted.
	Color         string  `json:"color"`         //string	Sets the default color of the edge. It can be an acceptable HTML color string. Default: #999
}

type ProcessFd struct {
	Id           string
	Name         string
	Path         string
	Size         int64
	Dm           string
	DeviceNumber uint64
	DevicePath   string
	MountPoint   string
}

type ProcessIO struct {
	ReadBytes  int64
	WriteBytes int64
	ReadIos    int64
	WriteIos   int64
}

type FileSystem struct {
	MountDevice  string
	MountPoint   string
	Type         string
	Option       string
	DeviceNumber uint64
	Major        uint64
	Minor        uint64
	DevicePath   string
}

type BolckDeviceStat struct {
	Dev     uint64
	Ino     uint64
	Nlink   uint64
	Mode    uint32
	Uid     uint32
	Gid     uint32
	Rdev    uint64
	Size    int64
	Blksize int64
	Blocks  int64
}

var FsTypeMap map[int64]string

func SetCode() {
	// 파일 시스템 타입 매핑
	FsTypeMap = map[int64]string{
		0xADFF:     "affs",       // Amiga Fast File System
		0x5346414F: "afs",        // Andrew File System
		0x09041934: "anon-inode", // Anonymous Inodes
		0x61756673: "aufs",       // Advanced Multi-Layered Unification Filesystem
		0x1BADFACE: "bpf_fs",     // Berkeley Packet Filter Filesystem
		0x42465331: "bfs",        // Boot File System (BFS)
		0x9123683E: "btrfs",      // B-tree File System
		0x73757245: "ecryptfs",   // Encrypted Filesystem
		0x51494C46: "qnx4",       // QNX4 File System
		0x52654973: "reiserfs",   // Reiser File System
		0xEF53:     "ext2/3/4",   // Extended Filesystem (EXT2/3/4)
		0xF15F:     "ecryptfs",   // eCryptFS (Encrypted)
		0x6969:     "nfs",        // Network File System (NFS)
		0xFF534D42: "cifs",       // Common Internet File System (CIFS)
		0x01021994: "tmpfs",      // Temporary Filesystem (TMPFS)
		0x58465342: "xfs",        // XFS Filesystem
		0x4d44:     "msdos",      // MS-DOS Filesystem
		0x137F:     "minix",      // Minix Filesystem
		0x3153464A: "jfs",        // Journaled File System (JFS)
		0x858458F6: "ramfs",      // RAM Filesystem
		//0x1021994:  "hugetlbfs",  // HugeTLB Filesystem
		0x28cd3d45: "cramfs",    // Compressed ROM File System
		0xE0F5E1E2: "nssockfs",  // Network Sockets Filesystem
		0x53464846: "fuse",      // FUSE Filesystem
		0x00011954: "overlayfs", // Overlay Filesystem
		0x19830326: "hfs",       // Hierarchical File System
		0x19540119: "hfsplus",   // Hierarchical File System Plus
		0x15013346: "efs",       // SGI Extended File System
		// 추가적인 파일 시스템 타입들을 여기에 추가할 수 있습니다.
	}
}
