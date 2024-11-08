package trace

import (
	"errors"
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
	NodeMetric *prometheus.GaugeVec
	EdgeMetric *prometheus.GaugeVec
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
	// Prometheus에 메트릭 등록
	prometheus.MustRegister(NodeMetric)
	prometheus.MustRegister(EdgeMetric)

}

func (t *Trace) CreateNodeGraphData() error {

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

	err = t.CreateNetworkMap()
	if err != nil {
		log.Error(err)
		return err
	}

	err = t.UpdatePidIo()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (t *Trace) CreatePrometheusMetric() (err error) {

	NodeMetric.Reset()
	EdgeMetric.Reset()

	t.createNodeHost()
	t.createNodePid()
	t.createNodePidFd()
	t.createNodeFS()
	t.createNodeDevice()
	t.createNodeDeviceMapper()
	t.createNodeIscsi()
	t.createNodeNic()

	t.createEdgeHostPid()
	t.createEdgePidFd()
	t.createEdgFdFs()
	t.createEdgFsDeviceNic()
	t.createEdgDeviceMapper()
	t.createEdgDeviceIscsi()
	t.createEdgIscsiNic()

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
	nodeIcon := "apps"
	nodeRadius := ""
	nodeHighlighted := "true"
	newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
	NodeMetric.WithLabelValues(newNode...).Set(1)
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
		NodeMetric.WithLabelValues(newNode...).Set(1)
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
			nodeIcon := "file-alt" //"file-alt","file-edit-alt"
			nodeRadius := ""
			nodeHighlighted := ""
			newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
			NodeMetric.WithLabelValues(newNode...).Set(1)
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
		nodeHighlighted := ""
		newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
		NodeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createNodeDevice() (err error) {
	for k, ds := range t.Dev {
		devicePath := t.DevicePathMap[k]
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
		NodeMetric.WithLabelValues(newNode...).Set(1)
	}
	return nil
}

func (t *Trace) createNodeDeviceMapper() (err error) {
	for _, ds := range t.Dev {
		for _, v := range ds.SlavesDev {
			ds := t.DeviceStatMap[v]
			devicePath := ds.DevicePath
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
			NodeMetric.WithLabelValues(newNode...).Set(1)
		}
	}
	return nil
}

func (t *Trace) createNodeIscsi() (err error) {
	initiator := t.ISCSIInfo.Interface.Initiator
	ipaddr := t.ISCSIInfo.Interface.IPAddress
	nodeId := fmt.Sprintf("iscsi:%s:%s", t.Hostname, initiator)
	nodeName := fmt.Sprintf("%s:%s", ipaddr, initiator)
	nodeMainStat := "online"
	nodeSubStat := fmt.Sprintf("%d MiB", 0) //size
	nodeArcFail := fmt.Sprintf("%f", 0.0)
	nodeArcPass := fmt.Sprintf("%f", 1.0)
	nodeRole := ""
	nodeColor := ""
	nodeIcon := ""
	nodeRadius := ""
	nodeHighlighted := ""
	newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
	NodeMetric.WithLabelValues(newNode...).Set(1)
	return nil
}

func (t *Trace) createNodeNic() (err error) {
	ipaddr := t.ISCSIInfo.Interface.IPAddress
	nicName, err := t.findNicByAddr(ipaddr)
	if err != nil {
		log.Error(err)
		return err
	}
	nodeId := fmt.Sprintf("nic:%s:%s", t.Hostname, nicName)
	nodeName := fmt.Sprintf("%s:%s", nicName, ipaddr)
	nodeMainStat := "online"
	nodeSubStat := fmt.Sprintf("%d MiB", 0) //size
	nodeArcFail := fmt.Sprintf("%f", 0.0)
	nodeArcPass := fmt.Sprintf("%f", 1.0)
	nodeRole := ""
	nodeColor := ""
	nodeIcon := ""
	nodeRadius := ""
	nodeHighlighted := ""
	newNode := []string{nodeId, nodeName, nodeMainStat, nodeSubStat, nodeArcFail, nodeArcPass, nodeRole, nodeColor, nodeIcon, nodeRadius, nodeHighlighted}
	NodeMetric.WithLabelValues(newNode...).Set(1)
	return nil
}

func (t *Trace) createEdgeHostPid() (err error) {
	for _, pid := range t.Pid {
		edgeId := fmt.Sprintf("%s:%d", t.Hostname, pid)
		edgeSource := fmt.Sprintf("host:%s", t.Hostname)
		edgeTarget := fmt.Sprintf("pid:%s:%d", t.Hostname, pid)
		edgeMainStat := "hostE1"
		edgeSecondarystat := "hostE2"
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
			edgeSource := fmt.Sprintf("pid:%s:%d", t.Hostname, pid)
			edgeTarget := fmt.Sprintf("fd:%s:%d:%s", t.Hostname, pid, fd.Id)
			edgeMainStat := "pidE1"
			edgeSecondarystat := "pidE2"
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
			edgeSource := fmt.Sprintf("fd:%s:%d:%s", t.Hostname, pid, fd.Id)
			edgeTarget := fmt.Sprintf("fs:%s:%s", t.Hostname, fd.MountPoint)
			edgeMainStat := "fdE1"
			edgeSecondarystat := "fdE2"
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

func (t *Trace) createEdgFsDeviceNic() (err error) {
	for _, fs := range t.Fs {
		switch fs.Type {
		case "nfs4", "nfs3":
			log.Debug("NFS Connection")
			log.Debugf("Mount device[%s] point[%s] type[%s]", fs.MountDevice, fs.MountPoint, fs.Type)
			nicName := ""
			nfsSvrAddrs, err := getIpAddrFromMounts(fs.MountDevice)
			if err != nil {
				log.Error(err)
				return err
			}
			for _, nfsAddr := range nfsSvrAddrs {
				localaddr, err := findInterfaceForAddress(nfsAddr+":22")
				if err != nil {
					log.Error(err)
					return err
				}				
				nicName, err = t.findNicByAddr(localaddr)
				if err != nil {
					log.Debug(err)
				}
			}
			if nicName == "" {
				msg := fmt.Sprintf("no nic [%v]", fs.MountDevice)
				log.Error(msg)
				return errors.New(msg)
			}
			edgeId := fmt.Sprintf("%s:%s:%s", t.Hostname, fs.MountPoint, nicName)
			edgeSource := fmt.Sprintf("fs:%s:%s", t.Hostname, fs.MountPoint)
			edgeTarget := fmt.Sprintf("nic:%s:%s", t.Hostname, nicName)
			edgeMainStat := "fsE1"
			edgeSecondarystat := "fsE2"
			edgeDetail__info := ""
			edgeThickness := ""
			edgeHighlighted := ""
			edgeColor := ""
			newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
			EdgeMetric.WithLabelValues(newNode...).Set(1)

		default:
			deviceNumber := fs.DeviceNumber
			devicePath := t.DevicePathMap[deviceNumber]
			edgeId := fmt.Sprintf("%s:%s:%s", t.Hostname, fs.MountPoint, devicePath)
			edgeSource := fmt.Sprintf("fs:%s:%s", t.Hostname, fs.MountPoint)
			edgeTarget := fmt.Sprintf("dev:%s:%s", t.Hostname, devicePath)
			edgeMainStat := "fsE1"
			edgeSecondarystat := "fsE2"
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

func (t *Trace) createEdgDeviceMapper() (err error) {
	for key, dv := range t.Dev {
		for _, sv := range dv.SlavesDev {
			sdevicePath := t.DevicePathMap[key]
			tdevicePath := t.DevicePathMap[sv]
			edgeId := fmt.Sprintf("%s:%s:%s", t.Hostname, sdevicePath, tdevicePath)
			edgeSource := fmt.Sprintf("dev:%s:%s", t.Hostname, sdevicePath)
			edgeTarget := fmt.Sprintf("dev:%s:%s", t.Hostname, tdevicePath)
			edgeMainStat := "devE1"
			edgeSecondarystat := "devE2"
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

func (t *Trace) createEdgDeviceIscsi() (err error) {
	for _, dv := range t.Dev {
		for _, slave := range dv.SlavesDev {
			sdevicePath := t.DevicePathMap[slave]
			initiator, err := t.findInitiatorByDevice(sdevicePath)
			if err != nil {
				log.Error(err)
				continue
			}
			if len(initiator) == 0 {
				continue
			}
			edgeId := fmt.Sprintf("%s:%s:%s", t.Hostname, sdevicePath, initiator)
			edgeSource := fmt.Sprintf("dev:%s:%s", t.Hostname, sdevicePath)
			edgeTarget := fmt.Sprintf("iscsi:%s:%s", t.Hostname, initiator)
			edgeMainStat := "iscsiE1"
			edgeSecondarystat := "iscsiE2"
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

func (t *Trace) createEdgIscsiNic() (err error) {
	initiator := t.ISCSIInfo.Interface.Initiator
	ipaddr := t.ISCSIInfo.Interface.IPAddress
	nicName, err := t.findNicByAddr(ipaddr)
	if err != nil {
		log.Error(err)
		return err
	}
	edgeId := fmt.Sprintf("%s:%s:%s", t.Hostname, initiator, ipaddr)
	edgeSource := fmt.Sprintf("iscsi:%s:%s", t.Hostname, initiator)
	edgeTarget := fmt.Sprintf("nic:%s:%s", t.Hostname, nicName)

	edgeMainStat := "iscsiE1"
	edgeSecondarystat := "iscsiE2"
	edgeDetail__info := ""
	edgeThickness := ""
	edgeHighlighted := ""
	edgeColor := ""
	newNode := []string{edgeId, edgeSource, edgeTarget, edgeMainStat, edgeSecondarystat, edgeDetail__info, edgeThickness, edgeHighlighted, edgeColor}
	EdgeMetric.WithLabelValues(newNode...).Set(1)
	return nil
}
