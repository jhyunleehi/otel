package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"ebpf-go/utils/command"
	"ebpf-go/utils/resty"

	"github.com/cilium/ebpf/link"

	// "github.com/cilium/ebpf/perf"
	// "github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type blkdevice struct {
	//lsblk --help
	path      string
	major     int
	minor     int
	majmin    string
	wwn       string
	mountpint string
	fstype    string
	fssize    int64
	fsused    int64
	vendor    string
}

var diskmap map[int]map[int]string
var devicemap map[int]map[int]blkdevice

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
		defer file.Close()
	}
	log.SetOutput(os.Stdout)
}

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

	// fn1 := "blk_account_io_done"
	// kp1, err := link.Kprobe(fn1, objs.BlkAccountIoDone, nil)
	// if err != nil {
	// 	log.Fatalf("opening kprobe: %s", err)
	// }
	// defer kp1.Close()

	// fn2 := "blk_account_io_start"
	// kp2, err := link.Kprobe(fn2, objs.BlkAccountIoStart, nil)
	// if err != nil {
	// 	log.Fatalf("opening kprobe: %s", err)
	// }
	// defer kp2.Close()

	fn3 := "blk_mq_start_request"
	kp3, err := link.Kprobe(fn3, objs.BlkMqStartRequest, nil)
	if err != nil {
		log.Fatalf("opening kprobe: %s", err)
	}
	defer kp3.Close()

	link1, err := link.AttachTracing(link.TracingOptions{
		Program: objs.bpfPrograms.BlockIoDone,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer link1.Close()

	link2, err := link.AttachTracing(link.TracingOptions{
		Program: objs.bpfPrograms.BlockIoStart,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer link2.Close()

	err = parseDiskStat()
	if err != nil {
		log.Fatal(err)
	}

	err = parseLsblk()
	if err != nil {
		log.Fatal(err)
	}

	l1, l5, l15, err := parseLoadAvg()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("[%f][%f][%f]", l1, l5, l15)

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
		keys := []bpfInfoT{}
		vals := []bpfValT{}
		key := bpfInfoT{}
		val := bpfValT{}
		iter := objs.Counts.Iterate()
		for iter.Next(&key, &val) {
			//log.Debugf("%+v", key)
			//log.Debugf("%+v", val)
			objs.Counts.Delete(&key)
			keys = append(keys, key)
			vals = append(vals, val)
		}

		for k, info := range keys {
			v := vals[k]

			var rw, diskname string
			var avgms float32
			major := info.Major
			minor := info.Minor
			pid := info.Pid
			name := convToString(info.Name[:])
			if info.Rwflag == 0 {
				rw = "W"
			} else {
				rw = "R"
			}
			diskname, exist := diskmap[int(major)][int(minor)]
			if !exist {
				diskname = "no disk name"
			}
			io := val.Io
			bytes := val.Bytes / 1024
			if val.Io != 0 {
				avgms = float32(v.Us) / float32(1000) / float32(v.Io)
			}
			fmt.Printf("%-7d %-16s %1s %-3d %-3d %-8s %5d %7d %6.2f\n",
				pid, name, rw, major, minor, diskname, io, bytes, avgms)
		}
	}

}

func parseLsblk() error {
	devicemap = make(map[int]map[int]blkdevice)
	output, err := command.ExecCommandTimeout("lsblk", 5, "-o", "path,maj:min,wwn")
	if err != nil {
		log.Error(err)
		return err
	}
	reader := bytes.NewReader([]byte(output))
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// Parse major and minor numbers
		var path, majmin, wwn string

		fmt.Sscanf(fields[0], "%d", &path)
		fmt.Sscanf(fields[1], "%d", &majmin)
		fmt.Sscanf(fields[2], "%s", &wwn)
		s := strings.Split(majmin, ":")
		major, _ := strconv.Atoi(s[0])
		minor, _ := strconv.Atoi(s[1])

		if _, exists := diskmap[major]; !exists {
			devicemap[major] = make(map[int]blkdevice)
		}
		blk := blkdevice{
			path:   path,
			major:  major,
			minor:  minor,
			majmin: majmin,
			wwn:    wwn,
		}
		devicemap[major][minor] = blk
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error reading file:", err)
		return err
	}
	return nil
}

func parseDiskStat() error {
	diskmap = make(map[int]map[int]string)
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		log.Error("Error opening file:", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Parse major and minor numbers
		var major, minor int
		var name string
		fmt.Sscanf(fields[0], "%d", &major)
		fmt.Sscanf(fields[1], "%d", &minor)
		fmt.Sscanf(fields[2], "%s", &name)

		if _, exists := diskmap[major]; !exists {
			diskmap[major] = make(map[int]string)
		}
		diskmap[major][minor] = name
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error reading file:", err)
		return err
	}
	return nil
}

func parseLoadAvg() (l1, l5, l15 float32, err error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		log.Error("Error opening file:", err)
		return 0.0, 0.0, 0.0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		fmt.Sscanf(fields[0], "%f", &l1)
		fmt.Sscanf(fields[1], "%f", &l5)
		fmt.Sscanf(fields[2], "%f", &l15)
	}
	if err := scanner.Err(); err != nil {
		log.Error("Error reading file:", err)
		return 0.0, 0.0, 0.0, err
	}
	return l1, l5, l15, nil
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
