package trace

import (
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Trace struct {
	NodeMetric *prometheus.GaugeVec
	EdgeMetric *prometheus.GaugeVec
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
}

func NewTrace(target *string) (Trace, error) {
	t := Trace{}
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

	// Prometheus에 메트릭 등록
	prometheus.MustRegister(t.NodeMetric)
	prometheus.MustRegister(t.EdgeMetric)
	return t, nil
}

func (t *Trace) UpdateNodeGraph() error {
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
