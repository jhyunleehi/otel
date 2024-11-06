package trace

import (
	"errors"
	"fmt"
	"os"
	"otel/model"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Trace struct {
	NodeMetric *prometheus.GaugeVec
	EdgeMetric *prometheus.GaugeVec
	MachineId  string
	Pids       []int
	Fds        map[int]map[string]model.ProcessFd
	Io         map[int]model.ProcessIO
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
}

func NewTrace(target *string) (t Trace, err error) {
	t = Trace{}
	t.NodeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "graph_node",
			Help: "Node data for Grafana's Node Graph",
		},
		[]string{
			"id",            //Unique identifier of the node. This ID is referenced by edge in its source and target field.
			"title",         //Name of the node visible in just under the node.
			"mainStat",      //First stat shown inside the node itself
			"secondaryStat", //Same as mainStat, but shown under it inside the node
			"arc__failed",   //to create the color circle around the node. All values in these fields should add up to 1.
			"arc__passed",
			"detail__role", //shown in the header of context menu when clicked on the node
			"color",        //Can be used to specify a single color instead of using the arc__ fields to specify color sections
			"icon",
			"nodeRadius",  //Radius value in pixels. Used to manage node size.
			"highlighted", //Sets whether the node should be highlighted.
		},
	)
	t.EdgeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "graph_edge",
			Help: "Edge relationship data for Grafana's Node Graph",
		},
		[]string{
			"id",            //Unique identifier of the edge.
			"source",        //Id of the source node.
			"target",        //Id of the target.
			"mainStat",      //First stat shown in the overlay when hovering over the edge.
			"secondarystat", //Same as mainStat, but shown right under it.
			"detail__info",  //will be shown in the header of context menu when clicked on the edge
			"thickness",     //The thickness of the edge. Default: 1
			"highlighted",   //boolean	Sets whether the edge should be highlighted.
			"color",         //string	Sets the default color of the edge. It can be an acceptable HTML color string. Default: #999
		},
	)
	t.MachineId, err = getMachinId()
	if err != nil {
		log.Error(err)
		return t, err
	}
	t.Pids, err = findPidByCmd(*target)
	if err != nil {
		log.Error(err)
		return t, err
	}
	// Prometheus에 메트릭 등록
	prometheus.MustRegister(t.NodeMetric)
	prometheus.MustRegister(t.EdgeMetric)
	return t, nil
}

func (t *Trace) UpdateNodeGraph() error {
	err := t.UpdateFd()
	if err != nil {
		log.Error(err)
		return err
	}
	err = t.UpdateIo()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (t *Trace) UpdateFd() error {
	t.Fds = make(map[int]map[string]model.ProcessFd)
	for _, pid := range t.Pids {
		if _, exist := t.Fds[pid]; !exist {
			t.Fds[pid] = map[string]model.ProcessFd{}
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
			fd := model.ProcessFd{}
			regularFile, err := IsRegularFile(target)
			if err != nil {
				log.Error(err)
				continue
			}
			if regularFile {
				fd.Id = file.Name()
				fd.Name = file.Name()
				fd.Path = target
				log.Debugf("FD %s -> %s\n", file.Name(), target)
				t.Fds[pid][file.Name()] = fd
			}
		}
	}
	log.Debug(t.Fds)
	return nil
}

func (t *Trace) UpdateIo() error {
	t.Io = make(map[int]model.ProcessIO)
	for _, pid := range t.Pids {
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
				case "read_bytes":
					pio.ReadBytes, _ = strconv.ParseInt(parts[1], 10, 64)
				case "write_bytes":
					pio.WriteBytes, _ = strconv.ParseInt(parts[1], 10, 64)
				case "read_ios":
					pio.ReadIos, _ = strconv.ParseInt(parts[1], 10, 64)
				case "write_ios":
					pio.WriteIos, _ = strconv.ParseInt(parts[1], 10, 64)
				}
			}
		}
		t.Io[pid] = pio
	}
	log.Debug(t.Io)
	return nil
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

func IsRegularFile(filePath string) (bool, error) {
	// 파일 정보 가져오기
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("File does not exist: %v", err)
		} else {
			log.Error("Error checking file: %v", err)
		}
		return false, err
	}

	// 일반 파일인지 확인
	if fileInfo.Mode().IsRegular() {
		log.Debugf("%s is a regular file.", filePath)
		return true, nil
	} else {
		log.Debugf("%s is not a regular file.\n", filePath)
		return false, nil
	}

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
