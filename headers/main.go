package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	// "github.com/cilium/ebpf/perf"
	// "github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"golang.org/x/sys/unix"
)

func main() {

	// Subscribe to signals for terminating the program.
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	fn := "vfs_read"
	kp1, err := link.Kprobe(fn, objs.VfsReadEntry, nil)
	if err != nil {
		log.Fatalf("opening kprobe: %s", err)
	}
	defer kp1.Close()

	fn = "vfs_write"
	kp2, err := link.Kprobe(fn, objs.VfsWriteEntry, nil)
	if err != nil {
		log.Fatalf("opening kprobe: %s", err)
	}
	defer kp2.Close()

	go func() {
		<-stopper

		log.Fatalf("closing ringbuf reader: %s", err)
	}()

	// Read loop reporting the total amount of times the kernel
	// function was entered, once per second.
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	log.Println("Waiting for events..")

	for range ticker.C {
		keys := []bpfFileId{}
		vals := []bpfFileStat{}
		key := bpfFileId{}
		val := bpfFileStat{}
		iter := objs.Entries.Iterate()
		for iter.Next(&key, &val) {
			//log.Printf("%+v", key)
			//log.Printf("%+v", val)
			objs.Entries.Delete(&key)
			keys = append(keys, key)
			vals = append(vals, val)
		}

		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Pid < keys[j].Pid
		})

		sort.Slice(vals, func(i, j int) bool {
			return vals[i].ReadBytes < vals[j].ReadBytes
		})

		for k, v := range vals {
			if k >10 {
				break
			}
			comm := convToString(v.Comm[:])
			filename := convToString(v.Filename[:])
			fmt.Printf("[%d] [%d][%d] [%s] : R[%d] W[%d] R[%d] W[%d] [%d] [%s]\n ", k, v.Pid, v.Tid, comm, v.Reads, v.Writes,
				v.ReadBytes/1024, v.WriteBytes/1024, v.Type, filename)			
		}
	}

}

func convToString(param interface{}) string {
	switch v := param.(type) {
	case int:
		fmt.Println("type:", v)
	case string:
		fmt.Println("type:", v)
		return param.(string)
	case []int8:
		ival := param.([]int8)
		dbytes := make([]byte, len(ival))
		for i, v := range ival {
			dbytes[i] = byte(v)
		}
		//str:=string(dbytes) //not handle null character
		str := unix.ByteSliceToString(dbytes)
		return strings.Trim(str, " ")
	default:
		fmt.Println("type:", v)
	}
	return ""
}

const (
	CREATE uint32 = 0
	OPEN   uint32 = 1
	INODE  uint32 = 2
	UNLINK uint32 = 3
)

func strerrno(eno int32) string {
	if eno < 0 {
		eno *= -1
	}
	errnum := syscall.Errno(eno)
	fmt.Printf("%d:%s", eno, errnum.Error())
	return errnum.Error()
}

func getflags(flags uint64) string {
	slist := []string{}
	if flags == 0 {
		return "0x0"
	}

	for i := 0; i < len(flag_names); i++ {
		if ((1 << i) & flags) == 0 {
			continue
		}
		slist = append(slist, flag_names[i])
	}
	return strings.Join(slist, "|")
}

var flag_names = []string{
	"MS_RDONLY",
	"MS_NOSUID",
	"MS_NODEV",
	"MS_NOEXEC",
	"MS_SYNCHRONOUS",
	"MS_REMOUNT",
	"MS_MANDLOCK",
	"MS_DIRSYNC",
	"MS_NOSYMFOLLOW",
	"MS_NOATIME",
	"MS_NODIRATIME",
	"MS_BIND",
	"MS_MOVE",
	"MS_REC",
	"MS_VERBOSE",
	"MS_SILENT",
	"MS_POSIXACL",
	"MS_UNBINDABLE",
	"MS_PRIVATE",
	"MS_SLAVE",
	"MS_SHARED",
	"MS_RELATIME",
	"MS_KERNMOUNT",
	"MS_I_VERSION",
	"MS_STRICTATIME",
	"MS_LAZYTIME",
	"MS_SUBMOUNT",
	"MS_NOREMOTELOCK",
	"MS_NOSEC",
	"MS_BORN",
	"MS_ACTIVE",
	"MS_NOUSER",
}
