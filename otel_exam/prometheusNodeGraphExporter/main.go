package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

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

var (
	// 노드 ID를 위한 Prometheus Gauge
	nodeMetric = prometheus.NewGaugeVec(
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

	// 엣지 관계를 위한 Prometheus Gauge
	edgeMetric = prometheus.NewGaugeVec(
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
	log.SetFormatter(&log.JSONFormatter{})
	// Prometheus에 메트릭 등록
	prometheus.MustRegister(nodeMetric)
	prometheus.MustRegister(edgeMetric)
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("MetricsHandler")
	// 예시 노드들 추가 (node_id와 라벨)
	n1 := []string{"node1", "Service A", "ok", "100", "0.5", "0.0", "role1", "", "", "", "true"}
	n2 := []string{"node2", "Service B", "ko", "200", "0.7", "0.0", "role2", "", "", "", ""}
	n3 := []string{"node3", "Service C", "ok", "300", "0.8", "0.0", "role3", "", "", "", ""}
	n4 := []string{"node4", "Service D", "ko", "400", "1.0", "0.0", "role4", "", "", "", ""}
	n5 := []string{"node5", "Service E", "ko", "500", "1.0", "0.0", "role5", "", "", "", ""}
	n6 := []string{"node6", "Service F", "ko", "500", "1.0", "0.0", "role5", "", "", "", ""}
	n7 := []string{"node7", "Service G", "ko", "500", "1.0", "0.0", "role5", "", "", "", ""}
	nodeMetric.WithLabelValues(n1...).Set(1)
	nodeMetric.WithLabelValues(n2...).Set(1)
	nodeMetric.WithLabelValues(n3...).Set(1)
	nodeMetric.WithLabelValues(n4...).Set(1)
	nodeMetric.WithLabelValues(n5...).Set(1)
	nodeMetric.WithLabelValues(n6...).Set(1)
	nodeMetric.WithLabelValues(n7...).Set(1)

	e1 := []string{"A-B", "node1", "node2", "E1", "100", "info1", "", "true", ""}
	e2 := []string{"A-C", "node1", "node3", "E2", "200", "info2", "0.5", "", ""}
	e3 := []string{"A-D", "node1", "node4", "E3", "300", "info3", "1", "", ""}
	e4 := []string{"2-3", "node2", "node3", "E4", "400", "info3", "2", "", ""}
	e5 := []string{"2-5", "node2", "node5", "E5", "500", "info4", "5", "", "blue"}
	e6 := []string{"5-6", "node5", "node6", "E5", "500", "info4", "10", "", "blue"}
	e7 := []string{"5-7", "node6", "node7", "E1", "100", "info1", "", "true", "green"}
	// 예시 엣지들 추가 (source_node와 target_node)
	edgeMetric.WithLabelValues(e1...).Set(1)
	edgeMetric.WithLabelValues(e2...).Set(1)
	edgeMetric.WithLabelValues(e3...).Set(1)
	edgeMetric.WithLabelValues(e4...).Set(1)
	edgeMetric.WithLabelValues(e5...).Set(1)
	edgeMetric.WithLabelValues(e6...).Set(1)
	edgeMetric.WithLabelValues(e7...).Set(1)

	promhttp.Handler().ServeHTTP(w, r)
}

func main() {

	// Prometheus 핸들러 설정
	http.HandleFunc("/metrics", MetricsHandler)
	//http.Handle("/metrics", promhttp.Handler())

	// HTTP 서버 실행
	log.Fatal(http.ListenAndServe(":2224", nil))
}
