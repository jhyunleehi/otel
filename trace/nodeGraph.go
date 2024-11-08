package trace

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"	
	log "github.com/sirupsen/logrus"
)

var (
)

func init() {
	log.SetLevel(log.DebugLevel)
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: time.RFC3339,
		NoColors:        true,
		CustomCallerFormatter: func(f *runtime.Frame) string {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return fmt.Sprintf("[%s:%d %s()] ", path.Base(f.File), f.Line, funcName)
		},
	})


}

func (t *Trace) CreateNodeGraph() error {

	err := t.CreateISCSIInfo()
	if err != nil {
		log.Error(err)
		return err
	}

	err = t.CreateDeviceMap()
	if err != nil {
		log.Error(err)
		return err
	}
	err = t.CreateFileSystemMap()
	if err != nil {
		log.Error(err)
		return err
	}
	err = t.CreatePidFdMap()
	if err != nil {
		log.Error(err)
		return err
	}

	err = t.CreateFsMap()
	if err != nil {
		log.Error(err)
		return err
	}

	err = t.CreateDevMap()
	if err != nil {
		log.Error(err)
		return err
	}

	err = t.CreatePidIo()
	if err != nil {
		log.Error(err)
		return err
	}

	err = t.CreatePrometheusMetric()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (t *Trace) CreatePrometheusMetric() (err error) {
	
	t.NodeMetric.Reset()
	t.EdgeMetric.Reset()

	t.createNodeHost()
	t.createNodePid()
	t.createNodePidFd()
	t.createNodeFS()
	t.createNodeDevice()

	t.createEdgeHostPid()
	t.createEdgePidFd()
	t.createEdgFdFs()
	//t.createEdgFsDevice()

	return nil
}

func (t *Trace) createNodeHost() (err error) {
	nodeId := fmt.Sprintf("host:%s", t.Hostname)
	nodeName := fmt.Sprintf("host:%s", t.Hostname)
	nodeMainStat := t.targetCommnd
	nodeSubStat := fmt.Sprintf("%d", 0)
	nodeArcFail := fmt.Sprintf("%f", 0.0)
	nodeArcPass := fmt.Sprintf("%f", 1.0)
	nodeRole := ""
	nodeColor := "green"
	nodeIcon := ""
	nodeRadius := ""
	nodeHighlighted := "true"
	newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
	t.NodeMetric.WithLabelValues(newNode...).Set(1)
	return nil
}

func (t *Trace) createNodePid() (err error) {
	//pid
	for _, pid := range t.Pid {
		nodeId := fmt.Sprintf("pid:%s:%d", t.Hostname, pid)
		nodeName := fmt.Sprintf("pid:%d", pid)
		nodeMainStat := "running"
		nodeSubStat := fmt.Sprintf("%d", t.Io[pid].ReadIos+t.Io[pid].WriteIos)
		nodeArcFail := fmt.Sprintf("%f", 0.0)
		nodeArcPass := fmt.Sprintf("%f", 1.0)
		nodeRole := ""
		nodeColor := "blue"
		nodeIcon := ""
		nodeRadius := ""
		nodeHighlighted := ""
		newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
		t.NodeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createNodePidFd() (err error) {
	for _, pid := range t.Pid {
		for _, fd := range t.Fd[pid] {
			nodeId := fmt.Sprintf("fd:%s:%d:%s", t.Hostname, pid, fd.Id)
			nodeName := fmt.Sprintf("fd:%s", fd.Path)
			nodeMainStat := "open"
			nodeSubStat := fmt.Sprintf("%d MiB", fd.Size/1024/12024) //size
			nodeArcFail := fmt.Sprintf("%f", 0.0)
			nodeArcPass := fmt.Sprintf("%f", 0.0)
			nodeRole := ""
			nodeColor := ""
			nodeIcon := "file-alt"
			nodeRadius := ""
			nodeHighlighted := ""
			newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
			t.NodeMetric.WithLabelValues(newNode...).Set(1)
		}
	}
	return nil
}

func (t *Trace) createNodeFS() (err error) {

	for mountPoint := range t.Fs {
		fs := t.FileSystemMap[mountPoint]
		nodeId := fmt.Sprintf("fs:%s:%s", t.Hostname, fs.MountPoint)
		nodeName := fs.MountPoint
		nodeMainStat := fs.Type
		nodeSubStat := fmt.Sprintf("%d MiB", 0) //size
		nodeArcFail := fmt.Sprintf("%f", 0.0)
		nodeArcPass := fmt.Sprintf("%f", 1.0)
		nodeRole := ""
		nodeColor := ""
		nodeIcon := ""
		nodeRadius := ""
		nodeHighlighted := "true"
		newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
		t.NodeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createNodeDevice() (err error) {
	for k, ds := range t.Dev {
		devicePath := t.DeviceMap[k]
		nodeId := fmt.Sprintf("dev:%s:%s", t.Hostname, devicePath)
		nodeName := devicePath
		nodeMainStat := "online"
		nodeSubStat := fmt.Sprintf("%d MiB", ds.Size/1024/1024) //size
		nodeArcFail := fmt.Sprintf("%f", 0.0)
		nodeArcPass := fmt.Sprintf("%f", 1.0)
		nodeRole := ""
		nodeColor := ""
		nodeIcon := ""
		nodeRadius := ""
		nodeHighlighted := ""
		newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
		t.NodeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createEdgeHostPid() (err error) {
	for _, pid := range t.Pid {
		edgeId := fmt.Sprintf("%s:%d", t.Hostname, pid)
		edgeSource := fmt.Sprintf("host:%s", t.Hostname)
		edgeTarget := fmt.Sprintf("pid:%s:%d", t.Hostname, pid)
		edgeMainStat := "open"
		edgeSecondarystat := "online"
		edgeDetail__info := ""
		edgeThickness := ""
		edgeHighlighted := ""
		edgeColor := ""
		newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
		t.EdgeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createEdgePidFd() (err error) {
	for _, pid := range t.Pid {
		for _, fd := range t.Fd[pid] {
			edgeId := fmt.Sprintf("%s:%d:%s", t.Hostname, pid, fd.Id)
			edgeSource := fmt.Sprintf("pid:%s:%d", t.Hostname, pid)
			edgeTarget := fmt.Sprintf("fd:%s:%d:%s", t.Hostname, pid, fd.Id)
			edgeMainStat := "open"
			edgeSecondarystat := "online"
			edgeDetail__info := ""
			edgeThickness := ""
			edgeHighlighted := ""
			edgeColor := ""
			newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
			t.EdgeMetric.WithLabelValues(newNode...).Set(1)
		}
	}
	return nil
}

func (t *Trace) createEdgFdFs() (err error) {
	for _, pid := range t.Pid {
		for _, fd := range t.Fd[pid] {
			edgeId := fmt.Sprintf("%s:%d:%s", t.Hostname, pid, fd.Id)
			edgeSource := fmt.Sprintf("fd:%s:%d:%s", t.Hostname, pid, fd.Id)			
			edgeTarget := fmt.Sprintf("fs:%s:%s", t.Hostname, fd.MountPoint)
			edgeMainStat := "open"
			edgeSecondarystat := "online"
			edgeDetail__info := ""
			edgeThickness := ""
			edgeHighlighted := ""
			edgeColor := ""
			newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
			t.EdgeMetric.WithLabelValues(newNode...).Set(1)
		}
	}
	return nil
}

func (t *Trace) createEdgFsDevice() (err error) {	
	for _, fs := range t.Fs {
		edgeId := fmt.Sprintf("%s:%s", t.Hostname, fs.MountPoint)
		edgeSource := fmt.Sprintf("fs:%s:%s", t.Hostname, fs.MountPoint)
		deviceNumber := fs.DeviceNumber
		dv := t.DeviceMap[deviceNumber]
		edgeTarget := fmt.Sprintf("dev:%s:%s", t.Hostname, dv)
		edgeMainStat := "open"
		edgeSecondarystat := "online"
		edgeDetail__info := ""
		edgeThickness := ""
		edgeHighlighted := ""
		edgeColor := ""
		newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
		t.EdgeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}
