package trace

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	NodeMetric = prometheus.NewGaugeVec(
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

	EdgeMetric = prometheus.NewGaugeVec(
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

	// Prometheus에 메트릭 등록
	prometheus.MustRegister(NodeMetric)
	prometheus.MustRegister(EdgeMetric)
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
	//pid
	err = t.createNodePid()
	if err != nil {
		log.Error(err)
	}

	err = t.createNodePidFd()
	if err != nil {
		log.Error(err)
	}

	err = t.createNodeFS()
	if err != nil {
		log.Error(err)
	}

	err = t.createNodeDevice()
	if err != nil {
		log.Error(err)
	}
	return nil
}

func (t *Trace) createNodeHost() (err error) {
	nodeId := fmt.Sprintf("%s", t.Hostname)
	nodeName := fmt.Sprintf("host:%s", t.Hostname)
	nodeMainStat := "running"
	nodeSubStat := fmt.Sprintf("%d", 0)
	nodeArcFail := fmt.Sprintf("%d", 0)
	nodeArcPass := fmt.Sprintf("%d", 100)
	nodeRole := ""
	nodeColor := ""
	nodeIcon := ""
	nodeRadius := ""
	nodeHighlighted := ""
	newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
	NodeMetric.WithLabelValues(newNode...).Set(1)
	return nil
}

func (t *Trace) createNodePid() (err error) {
	//pid
	for _, pid := range t.Pid {
		nodeId := fmt.Sprintf("%s:%d", t.Hostname, pid)
		nodeName := fmt.Sprintf("pid:%d", pid)
		nodeMainStat := "running"
		nodeSubStat := fmt.Sprintf("%d", t.Io[pid].ReadIos+t.Io[pid].WriteIos)
		nodeArcFail := fmt.Sprintf("%d", 0)
		nodeArcPass := fmt.Sprintf("%d", 100)
		nodeRole := ""
		nodeColor := ""
		nodeIcon := ""
		nodeRadius := ""
		nodeHighlighted := ""
		newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
		NodeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createNodePidFd() (err error) {
	for _, pid := range t.Pid {
		for _, fd := range t.Fd[pid] {
			nodeId := fmt.Sprintf("%s:%d:%s", t.Hostname, pid, fd.Id)
			nodeName := fmt.Sprintf("fd:%s", fd.Path)
			nodeMainStat := "open"
			nodeSubStat := fmt.Sprintf("%d MiB", fd.Size/1024/12024) //size
			nodeArcFail := fmt.Sprintf("%d", 0)
			nodeArcPass := fmt.Sprintf("%d", 100)
			nodeRole := ""
			nodeColor := ""
			nodeIcon := ""
			nodeRadius := ""
			nodeHighlighted := ""
			newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
			NodeMetric.WithLabelValues(newNode...).Set(1)
		}
	}
	return nil
}

func (t *Trace) createNodeFS() (err error) {
	for _, fs := range t.FileSystem {
		nodeId := fmt.Sprintf("%s:%s", t.Hostname, fs.MountPoint)
		nodeName := fs.MountPoint
		nodeMainStat := fs.Type
		nodeSubStat := fmt.Sprintf("%d MiB", 0) //size
		nodeArcFail := fmt.Sprintf("%d", 0)
		nodeArcPass := fmt.Sprintf("%d", 100)
		nodeRole := ""
		nodeColor := ""
		nodeIcon := ""
		nodeRadius := ""
		nodeHighlighted := ""
		newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
		NodeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createNodeDevice() (err error) {
	for k, dv := range t.DeviceMap {
		ds := t.DeviceStatMap[k]
		nodeId := fmt.Sprintf("%s:%s", t.Hostname, dv)
		nodeName := dv
		nodeMainStat := "oline"
		nodeSubStat := fmt.Sprintf("%d MiB", ds.Size/1024/1024) //size
		nodeArcFail := fmt.Sprintf("%d", 0)
		nodeArcPass := fmt.Sprintf("%d", 100)
		nodeRole := ""
		nodeColor := ""
		nodeIcon := ""
		nodeRadius := ""
		nodeHighlighted := ""
		newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
		NodeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createEdgeHostPid() (err error) {
	for _, pid := range t.Pid {
		edgeId := fmt.Sprintf("%s:%s", pid)
		edgeSource := fmt.Sprintf("%s", t.Hostname)
		edgeTarget := fmt.Sprintf("%s:%d", t.Hostname, pid)
		edgeMainStat := "open"
		edgeSecondarystat := "online"
		edgeDetail__info := ""
		edgeThickness := ""
		edgeHighlighted := ""
		edgeColor := ""
		newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
		EdgeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createEdgePidFd() (err error) {
	for _, pid := range t.Pid {
		for _, fd := range t.Fd[pid] {
			edgeId := fmt.Sprintf("%s:%d:%s", t.Hostname, pid, fd.Id)
			edgeSource := fmt.Sprintf("%s:%d", t.Hostname, pid)
			edgeTarget := fmt.Sprintf("%s:%d:%s", t.Hostname, pid, fd.Id)
			edgeMainStat := "open"
			edgeSecondarystat := "online"
			edgeDetail__info := ""
			edgeThickness := ""
			edgeHighlighted := ""
			edgeColor := ""
			newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
			EdgeMetric.WithLabelValues(newNode...).Set(1)
		}
	}
	return nil
}

func (t *Trace) createEdgFdFs() (err error) {
	for _, pid := range t.Pid {
		for _, fd := range t.Fd[pid] {
			edgeId := fmt.Sprintf("%s:%d:%s", t.Hostname, pid, fd.Id)
			edgeSource := fmt.Sprintf("%s:%d", t.Hostname, pid)
			//mountPoint :=fd.MountPoint
			edgeTarget := fmt.Sprintf("%s:%s", t.Hostname, fd.MountPoint)
			edgeMainStat := "open"
			edgeSecondarystat := "online"
			edgeDetail__info := ""
			edgeThickness := ""
			edgeHighlighted := ""
			edgeColor := ""
			newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
			EdgeMetric.WithLabelValues(newNode...).Set(1)
		}
	}
	return nil
}

func (t *Trace) createEdgFsDevice() (err error) {
	for _, fs := range t.FileSystem {
		edgeId := fmt.Sprintf("%s:%s", t.Hostname, fs.MountPoint)
		edgeSource := fmt.Sprintf("%s:%s", t.Hostname, fs.MountPoint)
		deviceNumber := fs.DeviceNumber
		dv := t.DeviceMap[deviceNumber]
		edgeTarget := fmt.Sprintf("%s:%s", t.Hostname, dv)
		edgeMainStat := "open"
		edgeSecondarystat := "online"
		edgeDetail__info := ""
		edgeThickness := ""
		edgeHighlighted := ""
		edgeColor := ""
		newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
		EdgeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}
