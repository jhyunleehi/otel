





```c
// This program demonstrates attaching an eBPF program to a network interface
// with XDP (eXpress Data Path). The program parses the IPv4 source address
// from packets and writes the packet count by IP to an LRU hash map.
// The userspace program (Go code in this file) prints the contents
// of the map to stdout every second.
// It is possible to modify the XDP program to drop or redirect packets
// as well -- give it a try!
// This example depends on bpf_link, available in Linux kernel version 5.7 or newer.
package main

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// //go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf xdp.c -- -I../headers

func main() {
	...
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s, err := formatMapContents(objs.XdpStatsMap)
		if err != nil {
			log.Printf("Error reading map: %s", err)
			continue
		}
		log.Printf("Map contents:\n%s", s)
	}
}

func formatMapContents(m *ebpf.Map) (string, error) {
	var (
		sb  strings.Builder
		key netip.Addr
		val uint32
	)
	iter := m.Iterate()
	for iter.Next(&key, &val) {
		sourceIP := key // IPv4 source address in network byte order.
		packetCount := val
		sb.WriteString(fmt.Sprintf("\t%s => %d\n", sourceIP, packetCount))
	}
	return sb.String(), iter.Err()
}

```

### how to get full path

```c
#include <linux/fs.h>
#include <linux/path.h>
#include <linux/mount.h>
#include <linux/string.h>

// Function to retrieve filesystem information
void get_filesystem_info(struct file *filp) {
    struct path path;
    char fs_type[32]; // Adjust size according to your needs

    // Get the struct path associated with the file
    if (filp && filp->f_path.dentry) {
        path = filp->f_path;
        
        // Access vfsmount structure from struct path
        struct vfsmount *mnt = path.mnt;

        // Access filesystem type and store it
        if (mnt && mnt->mnt_sb && mnt->mnt_sb->s_type && mnt->mnt_sb->s_type->name) {
            strncpy(fs_type, mnt->mnt_sb->s_type->name, sizeof(fs_type));
            fs_type[sizeof(fs_type) - 1] = '\0'; // Ensure null-termination
        } else {
            strcpy(fs_type, "Unknown");
        }

        // Now fs_type contains the filesystem type associated with the file
        printk(KERN_INFO "Filesystem type: %s\n", fs_type);
    }
}
```


### how to get Filesystem type from inode 

```c
#include <linux/fs.h>

// Function to get filesystem info based on s_magic
void get_filesystem_info(struct super_block *sb) {
    if (sb) {
        unsigned long magic = sb->s_magic;
        
        // Compare s_magic with known filesystem magic numbers
        switch (magic) {
            case EXT4_SUPER_MAGIC:
                printk(KERN_INFO "Filesystem: ext4\n");
                // Access ext4 specific information if needed
                break;
            case XFS_SUPER_MAGIC:
                printk(KERN_INFO "Filesystem: xfs\n");
                // Access xfs specific information if needed
                break;
            // Add more cases for other filesystem types as needed
            default:
                printk(KERN_INFO "Unknown filesystem\n");
                break;
        }
    }
}
```


```c
/* SPDX-License-Identifier: GPL-2.0 WITH Linux-syscall-note */
#ifndef __LINUX_MAGIC_H__
#define __LINUX_MAGIC_H__

#define ADFS_SUPER_MAGIC	0xadf5
#define AFFS_SUPER_MAGIC	0xadff
#define AFS_SUPER_MAGIC                0x5346414F
#define AUTOFS_SUPER_MAGIC	0x0187
#define CEPH_SUPER_MAGIC	0x00c36400
#define CODA_SUPER_MAGIC	0x73757245
#define CRAMFS_MAGIC		0x28cd3d45	/* some random number */
#define CRAMFS_MAGIC_WEND	0x453dcd28	/* magic number with the wrong endianess */
#define DEBUGFS_MAGIC          0x64626720
#define SECURITYFS_MAGIC	0x73636673
#define SELINUX_MAGIC		0xf97cff8c
#define SMACK_MAGIC		0x43415d53	/* "SMAC" */
#define RAMFS_MAGIC		0x858458f6	/* some random number */
#define TMPFS_MAGIC		0x01021994
#define HUGETLBFS_MAGIC 	0x958458f6	/* some random number */
#define SQUASHFS_MAGIC		0x73717368
#define ECRYPTFS_SUPER_MAGIC	0xf15f
#define EFS_SUPER_MAGIC		0x414A53
#define EROFS_SUPER_MAGIC_V1	0xE0F5E1E2
#define EXT2_SUPER_MAGIC	0xEF53
#define EXT3_SUPER_MAGIC	0xEF53
#define XENFS_SUPER_MAGIC	0xabba1974
#define EXT4_SUPER_MAGIC	0xEF53
#define BTRFS_SUPER_MAGIC	0x9123683E
#define NILFS_SUPER_MAGIC	0x3434
#define F2FS_SUPER_MAGIC	0xF2F52010
#define HPFS_SUPER_MAGIC	0xf995e849
#define ISOFS_SUPER_MAGIC	0x9660
#define JFFS2_SUPER_MAGIC	0x72b6
#define XFS_SUPER_MAGIC		0x58465342	/* "XFSB" */
#define PSTOREFS_MAGIC		0x6165676C
#define EFIVARFS_MAGIC		0xde5e81e4
#define HOSTFS_SUPER_MAGIC	0x00c0ffee
#define OVERLAYFS_SUPER_MAGIC	0x794c7630
#define FUSE_SUPER_MAGIC	0x65735546

#define MINIX_SUPER_MAGIC	0x137F		/* minix v1 fs, 14 char names */
#define MINIX_SUPER_MAGIC2	0x138F		/* minix v1 fs, 30 char names */
#define MINIX2_SUPER_MAGIC	0x2468		/* minix v2 fs, 14 char names */
#define MINIX2_SUPER_MAGIC2	0x2478		/* minix v2 fs, 30 char names */
#define MINIX3_SUPER_MAGIC	0x4d5a		/* minix v3 fs, 60 char names */

#define MSDOS_SUPER_MAGIC	0x4d44		/* MD */
#define EXFAT_SUPER_MAGIC	0x2011BAB0
#define NCP_SUPER_MAGIC		0x564c		/* Guess, what 0x564c is :-) */
#define NFS_SUPER_MAGIC		0x6969
#define OCFS2_SUPER_MAGIC	0x7461636f
#define OPENPROM_SUPER_MAGIC	0x9fa1
#define QNX4_SUPER_MAGIC	0x002f		/* qnx4 fs detection */
#define QNX6_SUPER_MAGIC	0x68191122	/* qnx6 fs detection */
#define AFS_FS_MAGIC		0x6B414653


#define REISERFS_SUPER_MAGIC	0x52654973	/* used by gcc */
					/* used by file system utilities that
	                                   look at the superblock, etc.  */
#define REISERFS_SUPER_MAGIC_STRING	"ReIsErFs"
#define REISER2FS_SUPER_MAGIC_STRING	"ReIsEr2Fs"
#define REISER2FS_JR_SUPER_MAGIC_STRING	"ReIsEr3Fs"

#define SMB_SUPER_MAGIC		0x517B
#define CIFS_SUPER_MAGIC	0xFF534D42      /* the first four bytes of SMB PDUs */
#define SMB2_SUPER_MAGIC	0xFE534D42

#define CGROUP_SUPER_MAGIC	0x27e0eb
#define CGROUP2_SUPER_MAGIC	0x63677270

#define RDTGROUP_SUPER_MAGIC	0x7655821

#define STACK_END_MAGIC		0x57AC6E9D

#define TRACEFS_MAGIC          0x74726163

#define V9FS_MAGIC		0x01021997

#define BDEVFS_MAGIC            0x62646576
#define DAXFS_MAGIC             0x64646178
#define BINFMTFS_MAGIC          0x42494e4d
#define DEVPTS_SUPER_MAGIC	0x1cd1
#define BINDERFS_SUPER_MAGIC	0x6c6f6f70
#define FUTEXFS_SUPER_MAGIC	0xBAD1DEA
#define PIPEFS_MAGIC            0x50495045
#define PROC_SUPER_MAGIC	0x9fa0
#define SOCKFS_MAGIC		0x534F434B
#define SYSFS_MAGIC		0x62656572
#define USBDEVICE_SUPER_MAGIC	0x9fa2
#define MTD_INODE_FS_MAGIC      0x11307854
#define ANON_INODE_FS_MAGIC	0x09041934
#define BTRFS_TEST_MAGIC	0x73727279
#define NSFS_MAGIC		0x6e736673
#define BPF_FS_MAGIC		0xcafe4a11
#define AAFS_MAGIC		0x5a3c69f0
#define ZONEFS_MAGIC		0x5a4f4653

/* Since UDF 2.01 is ISO 13346 based... */
#define UDF_SUPER_MAGIC		0x15013346
#define DMA_BUF_MAGIC		0x444d4142	/* "DMAB" */
#define DEVMEM_MAGIC		0x454d444d	/* "DMEM" */
#define SECRETMEM_MAGIC		0x5345434d	/* "SECM" */

#endif /* __LINUX_MAGIC_H__ */

```