package trace

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"otel/iscsi"
	"otel/model"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Trace struct {
	MachineId     string
	Hostname      string
	targetCommnd  string
	Pid           []int
	Fd            map[int]map[string]model.ProcessFd //pid,fd index
	Io            map[int]model.ProcessIO            //pid index
	Fs            map[string]model.FileSystem        //process file system
	Dev           map[uint64]model.BolckDeviceStat   //process device
	Nic           map[string]model.Interface
	FileSystemMap map[string]model.FileSystem      //all file system
	DevicePathMap map[uint64]string                //all device path
	DeviceStatMap map[uint64]model.BolckDeviceStat //all device info
	ISCSIInfo     iscsi.ISCSIInfo
}

func NewTrace(target *string) (t Trace, err error) {
	t = Trace{}
	t.MachineId, err = getMachinId()
	if err != nil {
		log.Error(err)
		return t, err
	}
	t.Hostname, err = getHostname()
	if err != nil {
		log.Error(err)
		return t, err
	}
	t.targetCommnd = *target
	t.Pid, err = findPidByCmd(*target)
	if err != nil {
		log.Error(err)
		return t, err
	}

	return t, nil
}

func (t *Trace) CreatePidFdMap() error {
	t.Fd = make(map[int]map[string]model.ProcessFd)
	for _, pid := range t.Pid {
		if _, exist := t.Fd[pid]; !exist {
			t.Fd[pid] = map[string]model.ProcessFd{}
		}
		fdDir := fmt.Sprintf("/proc/%d/fd", pid)
		files, err := os.ReadDir(fdDir)
		if err != nil {
			msg := fmt.Sprint("Error reading /proc/<pid>/fd directory:", err)
			log.Error(msg)
			return errors.New(msg)
		}
		for _, file := range files {
			// 파일 디스크립터 경로
			fdPath := filepath.Join(fdDir, file.Name())

			// 심볼릭 링크를 통해 실제 파일 경로 확인
			target, err := os.Readlink(fdPath)
			if err != nil {
				msg := fmt.Sprintf("Error reading symlink for fd %s: %v", file.Name(), err)
				log.Error(msg)
				continue
			}
			regularFile, err := IsRegularFile(target)
			if err != nil {
				log.Error(err)
				continue
			}
			if regularFile {
				fd := model.ProcessFd{}
				fd.Id = file.Name()
				fd.Name = file.Name()
				fd.Path = target

				finfo, err := GetFileStat(target)
				if err != nil {
					log.Error(err)
				} else {
					fd.Size = finfo.Size()
				}

				deviceNumber, err := getDeviceNumber(target)
				if err != nil {
					log.Error(err)
					deviceNumber = 0
				}
				fd.DeviceNumber = deviceNumber
				if _, exist := t.DevicePathMap[deviceNumber]; !exist {
					fd.DevicePath = t.DevicePathMap[deviceNumber]
				}
				fd.MountPoint = target

				log.Debugf("FD %s -> %s", file.Name(), target)
				fs, err := t.findFileSystem(target)
				if err != nil {
					log.Error(err)
				} else {
					fd.MountPoint = fs.MountPoint
				}
				log.Debugf("%+v", fd)
				t.Fd[pid][file.Name()] = fd
			}
		}
	}
	log.Debug(t.Fd)
	return nil
}

func (t *Trace) CreateFsMap() error {
	t.Fs = make(map[string]model.FileSystem)
	for _, pid := range t.Pid {
		for _, fd := range t.Fd[pid] {
			mountPoint := fd.MountPoint
			t.Fs[mountPoint] = t.FileSystemMap[mountPoint]
		}
	}
	log.Debug(t.Fs)
	return nil
}

func (t *Trace) CreateDevMap() error {
	t.Dev = make(map[uint64]model.BolckDeviceStat)
	for _, pid := range t.Pid {
		for _, fd := range t.Fd[pid] {
			deviceNumber := fd.DeviceNumber
			if devicePath, exists := t.DevicePathMap[deviceNumber]; exists {
				if len(devicePath) == 0 {
					continue
				}
				t.Dev[deviceNumber] = t.DeviceStatMap[deviceNumber]
			}
		}
	}
	log.Debug(t.Dev)
	return nil
}

func (t *Trace) UpdateDevSlave() error {
	// dm-0 device 처리
	// /sys/class/block/dm-0/slaves device 추출
	// slaves 항목에 device 저장
	//1. t.Dev에서 used device 추출
	//2. t.Dev의 deviceNumber를 t.DeviceMap[n] 넣어서 device path 식별
	//3. /dev/dm-0 경로 정상 확인
	//4. /sys/calss/block/dm-0/slaves device 추출
	//5. slave device number를 저장
	//6. 나중에 edge 그릴때 slave 있는 경우는 추적해서 추적하여 그리도록 DFS 방식으로 구현
	for deviceNumber := range t.Dev {
		devicePath := t.DevicePathMap[deviceNumber]
		log.Debugf("[%d][%s]", deviceNumber, devicePath)
		if strings.Contains(devicePath, "dm") {
			fields := strings.Split(devicePath, "/")
			deviceMapper := fields[len(fields)-1]

			// /sys/class/block/dm-0/slaves/ 디렉터리에서 기본 디바이스 찾기
			slaveDir := filepath.Join("/sys/class/block", deviceMapper, "slaves")
			slaveDevices, err := os.ReadDir(slaveDir)
			if err != nil {
				log.Errorf("Error reading %s: %v", slaveDir, err)
				return err
			}

			if len(slaveDevices) == 0 {
				log.Debug("No underlying devices found for", deviceMapper)
				return nil
			}

			// 기본 디바이스에 대한 정보를 출력
			log.Printf("Underlying devices for %s:\n", deviceMapper)
			slaveDevName := []string{}
			slaveDev := []uint64{}
			for _, slave := range slaveDevices {
				// 기본 디바이스 경로 (예: "/dev/sda1")
				slavePath := filepath.Join("/dev", slave.Name())
				slaveDevName = append(slaveDevName, slavePath)
				dn, err := getDeviceRdev(slavePath)
				if err != nil {
					log.Error(err)
				} else {
					slaveDev = append(slaveDev, dn)
				}
			}
			obj := t.Dev[deviceNumber]
			obj.SlavesDevName = slaveDevName
			obj.SlavesDev = slaveDev
			t.Dev[deviceNumber] = obj
		}

	}
	log.Debug(t.Dev)
	return nil
}

func (t *Trace) UpdatePidIo() error {
	t.Io = make(map[int]model.ProcessIO)
	for _, pid := range t.Pid {
		ioPath := fmt.Sprintf("/proc/%d/io", pid)
		data, err := os.ReadFile(ioPath)
		if err != nil {
			log.Error(err)
			return err
		}
		pio := model.ProcessIO{}
		// 파일 내용에서 필요한 값 추출 (read_bytes, write_bytes, read_ios, write_ios)
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				switch parts[0] {
				case "read_bytes:":
					pio.ReadBytes, _ = strconv.ParseInt(parts[1], 10, 64)
				case "write_bytes:":
					pio.WriteBytes, _ = strconv.ParseInt(parts[1], 10, 64)
				case "syscr:":
					pio.ReadIos, _ = strconv.ParseInt(parts[1], 10, 64)
				case "syscw:":
					pio.WriteIos, _ = strconv.ParseInt(parts[1], 10, 64)
				}
			}
		}
		t.Io[pid] = pio
	}
	log.Debug(t.Io)
	return nil
}

func (t *Trace) CreateNetworkMap() (err error) {
	t.Nic, err = getNetworkInterfaces()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (t *Trace) findFileSystem(filePath string) (fs model.FileSystem, err error) {
	longestMatch := ""
	for _, v := range t.FileSystemMap {
		// 지정된 경로가 현재 마운트 지점 하위에 있는지 확인
		if strings.HasPrefix(filePath, v.MountPoint) && len(v.MountPoint) > len(longestMatch) {
			longestMatch = v.MountPoint
			fs = v
		}
	}
	if longestMatch == "" {
		msg := fmt.Sprintf("no matching mount point found for path: %s", filePath)
		log.Error(msg)
		return fs, errors.New(msg)
	}
	return fs, nil
}

func (t *Trace) CreateFileSystemMap() error {
	t.FileSystemMap = make(map[string]model.FileSystem)

	file, err := os.Open("/proc/mounts")
	if err != nil {
		log.Error(err)
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
		mntDevice := fields[0] // 장치 이름 (예: /dev/sda1)
		fsType := fields[2]    // 파일 시스템 타입
		// 장치가 /dev로 시작하지 않는 경우, 스킵
		if strings.HasPrefix(mntDevice, "/dev") || strings.HasPrefix(fsType, "nfs") {
			fs := model.FileSystem{}
			fs.MountDevice = mntDevice // 장치 이름 (예: /dev/sda1)
			fs.MountPoint = fields[1]  // 마운트 지점
			fs.Type = fields[2]        // 파일 시스템 타입
			fs.Option = fields[3]      // 파일 시스템 타입

			var stat syscall.Stat_t
			err := syscall.Stat(mntDevice, &stat)
			if err != nil {
				log.Errorf("Could not stat device %s: %v", mntDevice, err)
			} else {
				fs.DeviceNumber = stat.Rdev
				fs.Major = (stat.Rdev >> 8) & 0xff
				fs.Minor = stat.Rdev & 0xff
				fs.DevicePath = t.DevicePathMap[fs.DeviceNumber]
			}
			t.FileSystemMap[fs.MountPoint] = fs
		}
	}
	log.Debug(t.FileSystemMap)
	return nil
}

func (t *Trace) CreateDeviceMap() error {
	t.DevicePathMap = make(map[uint64]string)
	t.DeviceStatMap = make(map[uint64]model.BolckDeviceStat)

	// /dev 디렉토리 읽기
	devices, err := os.ReadDir("/dev")
	if err != nil {
		log.Errorf("Failed to read /dev directory: %v", err)
		return err
	}

	// /dev 내의 각 장치 파일에 대해 device number를 가져와서 map에 저장
	for _, device := range devices {
		devicePath := filepath.Join("/dev", device.Name())

		// 장치 번호 가져오기
		stat, err := getDeviceStat_t(devicePath)
		if err != nil {
			log.Debugf("Skipping %s: %v\n", devicePath, err)
			continue
		}
		// block device 만 처리
		if (stat.Mode & syscall.S_IFMT) != syscall.S_IFBLK {
			continue
		}

		deviceNumber := uint64(stat.Rdev)
		t.DevicePathMap[deviceNumber] = devicePath

		//device mapper, slaves 처리
		slaveDevName := []string{}
		slaveDev := []uint64{}
		if strings.Contains(device.Name(), "dm") {
			deviceMapper := device.Name()
			// /sys/class/block/dm-0/slaves/ 디렉터리에서 기본 디바이스 찾기
			slaveDir := filepath.Join("/sys/class/block", deviceMapper, "slaves")
			slaveDevices, err := os.ReadDir(slaveDir)
			if err != nil {
				log.Errorf("Error reading %s: %v", slaveDir, err)
				return err
			}
			if len(slaveDevices) != 0 {
				// 기본 디바이스에 대한 정보를 출력
				log.Printf("Underlying devices for %s:\n", deviceMapper)
				for _, slave := range slaveDevices {
					// 기본 디바이스 경로 (예: "/dev/sda1")
					slavePath := filepath.Join("/dev", slave.Name())
					slaveDevName = append(slaveDevName, slavePath)
					dn, err := getDeviceRdev(slavePath)
					if err != nil {
						log.Error(err)
					} else {
						slaveDev = append(slaveDev, dn)
					}
				}
			}
		}

		ds := model.BolckDeviceStat{
			Dev:           stat.Dev,
			Ino:           stat.Ino,
			Nlink:         stat.Nlink,
			Mode:          stat.Mode,
			Uid:           stat.Uid,
			Gid:           stat.Gid,
			Rdev:          stat.Rdev,
			Size:          stat.Size,
			Blksize:       stat.Blksize,
			Blocks:        stat.Blocks,
			DevicePath:    devicePath,
			SlavesDev:     slaveDev,
			SlavesDevName: slaveDevName,
		}
		// map에 추가
		t.DeviceStatMap[deviceNumber] = ds
	}

	// 결과 출력
	// log.Debug("Device Map (device number -> device path):")
	// for devNum, path := range t.DeviceMap {
	// 	log.Debugf("Device Number: %d, Path: %s", devNum, path)
	// }

	return nil
}

func (t *Trace) CreateISCSIInfo() (err error) {
	t.ISCSIInfo, err = iscsi.GetISCSISession()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug(t.ISCSIInfo)
	return nil
}

func (t *Trace) findInitiatorByDevice(devicePath string) (initiator string, err error) {
	fields := strings.Split(devicePath, "/")
	if len(fields) < 2 {
		msg := fmt.Sprintf("no device name [%s]", devicePath)
		log.Error(msg)
		return initiator, errors.New(msg)
	}
	dname := fields[len(fields)-1]
	for _, device := range t.ISCSIInfo.AttachedSCSIDevices {
		if device.Device == dname {
			return t.ISCSIInfo.Interface.Initiator, nil
		}
	}
	return "", nil
}

func (t *Trace) findNicByAddr(addr string) (nicName string, err error) {
	nicName = ""
	for nicName, nic := range t.Nic {
		for _, v := range nic.Addrs {
			if strings.HasPrefix(v, addr) {
				return nicName, nil
			}
		}
	}
	msg := fmt.Sprintf("not found nic name [%s]", addr)
	log.Error(msg)
	return "", errors.New(msg)
}

func getMachinId() (string, error) {
	// /etc/machine-id 파일 경로
	const machineIDPath = "/etc/machine-id"

	// 파일 읽기
	data, err := os.ReadFile(machineIDPath)
	if err != nil {
		log.Errorf("Failed to read machine-id: %v", err)
		return "", err

	}

	// 파일 내용 출력 (Trailing newline 제거)
	machineID := strings.TrimSpace(string(data))
	log.Debugf("Machine ID: %s", machineID)
	return machineID, nil
}

func GetFileStat(filePath string) (os.FileInfo, error) {
	// 파일 정보 가져오기
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("File does not exist: %v", err)
		} else {
			log.Errorf("Error checking file: %v", err)
		}
		return nil, err
	}
	return fileInfo, nil

}

func IsRegularFile(filePath string) (bool, error) {
	// 파일 정보 가져오기
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("File does not exist: %v", err)
		} else {
			log.Errorf("Error checking file: %v", err)
		}
		return false, err
	}

	// 일반 파일인지 확인
	if fileInfo.Mode().IsRegular() {
		log.Debugf("%s is a regular file.", filePath)
		return true, nil
	} else {
		log.Debugf("%s is not a regular file.", filePath)
		return false, nil
	}
}

func getDeviceStat_t(path string) (stat *syscall.Stat_t, err error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Errorf("Error getting file info for %s: %v", path, err)
		return stat, err
	}

	// 파일의 syscall.Stat_t로부터 Device ID를 얻음
	s, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return stat, fmt.Errorf("failed to get stat_t for %s", path)
	}
	//Rdev Rdev 필드는 특수 파일(특히 블록 장치 또는 문자 장치 파일)의 실제 장치를 나타냅니다.
	//일반 파일이나 디렉토리에서는 Rdev 필드는 의미가 없습니다.
	return s, nil
}

func getDeviceRdev(path string) (uint64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Errorf("Error getting file info for %s: %v", path, err)
		return 0, err
	}

	// 파일의 syscall.Stat_t로부터 Device ID를 얻음
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get stat_t for %s", path)
	}
	//Rdev Rdev 필드는 특수 파일(특히 블록 장치 또는 문자 장치 파일)의 실제 장치를 나타냅니다.
	//일반 파일이나 디렉토리에서는 Rdev 필드는 의미가 없습니다.
	return uint64(stat.Rdev), nil
}

func getDeviceNumber(filePath string) (deviceNumber uint64, err error) {
	// 파일의 os.FileInfo 정보 가져오기
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Errorf("Error getting file info for %s: %v", filePath, err)
		return 0, err
	}

	// 파일의 디바이스 정보 추출
	stat := fileInfo.Sys().(*syscall.Stat_t)
	deviceNumber = stat.Dev
	//Dev 필드는 파일 시스템이 위치한 장치를 나타냅니다.
	//주로 파일 시스템의 루트 디렉토리나 일반 파일, 디렉토리 등에서 사용됩니다.
	return deviceNumber, nil
}

func getDevicePath(filePath string) (devicePath string, err error) {
	// 파일의 os.FileInfo 정보 가져오기
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Errorf("Error getting file info for %s: %v", filePath, err)
		return "", err
	}

	// 파일의 디바이스 정보 추출
	stat := fileInfo.Sys().(*syscall.Stat_t)
	deviceNumber := stat.Dev

	devDir := "/dev"
	files, err := os.ReadDir(devDir)
	if err != nil {
		log.Error("Error reading /dev directory:", err)
		return "", err
	}

	// 각 파일을 확인하여 디바이스 번호가 일치하는지 검사
	for _, file := range files {
		devPath := filepath.Join(devDir, file.Name())
		info, err := os.Stat(devPath)
		if err != nil {
			continue
		}

		// 디바이스 파일의 장치 번호를 확인하여 비교
		if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Rdev == deviceNumber {
			return devPath, nil
		}
	}
	msg := fmt.Sprintf("not found file[%s]", filePath)
	log.Error(msg)
	return "", errors.New(msg)
}

func getFilesystemType(path string) (string, error) {
	var statfs syscall.Statfs_t

	// Statfs 호출로 파일 시스템 정보 가져오기
	err := syscall.Statfs(path, &statfs)
	if err != nil {
		msg := fmt.Sprintf("failed to get filesystem type for %s: %v", path, err)
		log.Error(msg)
		return "", errors.New(msg)
	}

	// fsTypeMap을 통해 파일 시스템 타입 문자열 반환
	fsName, found := model.FsTypeMap[int64(statfs.Type)]
	if !found {
		fsName = "unknown"
	}

	return fsName, nil
}

func findPidByCmd(command string) (pids []int, err error) {
	// /proc 디렉터리에서 실행 중인 "fio" 프로세스 찾기
	files, err := os.ReadDir("/proc")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// "fio" 프로세스 PID 찾기
	for _, file := range files {
		if pid, err := strconv.Atoi(file.Name()); err == nil {
			cmdLine, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
			if err != nil {
				log.Error(err)
				continue
			}
			if strings.TrimSpace(string(cmdLine)) == command {
				pids = append(pids, pid)
			}
		}
	}
	log.Debug(pids)
	return pids, nil
}

func parseInt(s string) int {
	var res int
	fmt.Sscanf(s, "%d", &res)
	return res
}

func getHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("Error getting hostname: %v", err)
		return "", err
	}

	log.Debugf("Hostname: [%s]", hostname)
	return hostname, nil
}

func getNetworkInterfaces() (nic map[string]model.Interface, err error) {
	nic = map[string]model.Interface{}
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Errorf("Failed to get network interfaces: %v\n", err)
		return nic, err
	}
	for _, iface := range interfaces {
		if len(iface.Name) == 0 {
			continue
		}
		log.Debugf("Interface: %s", iface.Name)
		addrs, err := iface.Addrs()
		if err != nil {
			log.Errorf("Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}
		address := []string{}
		for _, addr := range addrs {
			log.Debugf("%s", addr.String())
			address = append(address, addr.String())
		}
		ifs := model.Interface{
			Name:  iface.Name,
			Addrs: address,
		}
		nic[iface.Name] = ifs
	}
	return nic, nil
}

// resolveHostnameToIP는 hostname을 IP 주소로 변환합니다.
func resolveHostnameToIP(hostname string) ([]net.IP, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		log.Errorf("failed to lookup IP for hostname %s: %w", hostname, err)
		return ips, err
	}
	return ips, nil
}

// getHostnameFromMounts는 /proc/mounts 파일에서 hostname을 추출합니다.
func getIpAddrFromMounts(mountdev string) (addr []string, err error) {
	hostname := ""
	if strings.Contains(mountdev, ":") {
		parts := strings.Split(mountdev, ":")
		if len(parts) > 0 {
			hostname = parts[0]
		}
	}

	if hostname == "" {
		msg := fmt.Sprintf("not fount hostname [%s]", mountdev)
		log.Error(msg)
		return addr, errors.New(msg)
	}

	ips, err := resolveHostnameToIP(hostname)
	if err != nil {
		log.Error(err)
		return addr, err
	}

	for _, v := range ips {
		ipaddr := v.String()
		addr = append(addr, ipaddr)
	}
	if len(addr) == 0 {
		msg := fmt.Sprintf("not fount hostname [%s]", mountdev)
		log.Error(msg)
		return addr, errors.New(msg)
	}
	return addr, nil
}

// findInterfaceForAddress는 특정 IP 주소와의 연결에서 사용되는 네트워크 인터페이스를 확인합니다.
func findInterfaceForAddress(target string) (string, error) {
	conn, err := net.Dial("udp", target)
	if err != nil {
		log.Errorf("Failed to dial %s: %v", target, err)
		return "", err
	}
	defer conn.Close()

	// 로컬 주소에서 사용된 IP 및 포트를 가져옴
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	addr := localAddr.IP.String()
	return addr, nil
}
