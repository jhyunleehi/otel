package main

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
}

// Field 구조체 정의
type Field struct {
	FieldName string `json:"field_name"`
	Type      string `json:"type"`
}

// Response 구조체 정의
type Response struct {
	Edges []Field `json:"edges_fields"`
	Nodes []Field `json:"nodes_fields"`
}

// nodeGraphHandler 함수
func fieldsHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("fieldsHandler")
	// Edges 필드 정의
	edgesFields := []Field{
		{FieldName: "id", Type: "string"},
		{FieldName: "source", Type: "string"},
		{FieldName: "target", Type: "string"},
		{FieldName: "mainStat", Type: "number"},
	}

	// Nodes 필드 정의
	nodesFields := []Field{
		{FieldName: "id", Type: "string"},
		{FieldName: "title", Type: "string"},
		{FieldName: "mainStat", Type: "string"},
		{FieldName: "secondaryStat", Type: "number"},
		{FieldName: "arc__failed", Type: "number"},
		{FieldName: "arc__passed", Type: "number"},
		{FieldName: "detail__role", Type: "string"},
		{FieldName: "color", Type: "string"},
		{FieldName: "display_name", Type: "string"},
		{FieldName: "icon", Type: "string"},
	}

	// 응답 데이터 생성
	response := Response{
		Edges: edgesFields,
		Nodes: nodesFields,
	}

	// JSON 형식으로 응답 작성
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Node 구조체
type Node struct {
	ID            string  `json:"id"`
	Title         string  `json:"title,omitempty"`
	MainStat      string  `json:"mainStat,omitempty"`
	SecondaryStat float32 `json:"secondaryStat,omitempty"`
	ArcFailed     float32 `json:"arc__failed,omitempty"`
	ArcPassed     float32 `json:"arc__passed,omitempty"`
	DetailRole    float32 `json:"detail__role,omitempty"`
	Color         string  `json:"color,omitempty"`
	DisplayName   string  `json:"display_name,omitempty"`
	Icon          string  `json:"icon,omitempty"`
}

// Edge 구조체
type Edge struct {
	ID       string  `json:"id"`
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	MainStat float32 `json:"mainStat,omitempty"`
}

// Response 구조체 (Node와 Edge 데이터 포함)
type ResponseData struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func nodeGraphHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("nodeGraphHandler")
	// Node 정보 정의
	nodes := []Node{
		{ID: "node1", Title: "Database", MainStat: "ok", SecondaryStat: 10, ArcFailed: 0.5, ArcPassed: 0.5, Icon: "database"},
		{ID: "node2", Title: "API Server", MainStat: "bad", SecondaryStat: 20, ArcFailed: 0.9, ArcPassed: 0.1, Icon: "server"},
		{ID: "node3", Title: "Frontend", MainStat: "good", SecondaryStat: 30, ArcFailed: 0.1, ArcPassed: 0.1},
	}

	// Edge 정보 정의
	edges := []Edge{
		{ID: "1", Source: "node1", Target: "node2", MainStat: 1},
		{ID: "2", Source: "node2", Target: "node3", MainStat: 2},
		{ID: "3", Source: "node1", Target: "node3", MainStat: 5},
	}

	// Response 데이터 생성
	response := ResponseData{
		Nodes: nodes,
		Edges: edges,
	}

	// JSON 형식으로 응답 작성
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("healthHandler")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func main() {
	// /nodegraph 엔드포인트에 핸들러 연결
	http.HandleFunc("/api/graph/fields", fieldsHandler)
	http.HandleFunc("/api/graph/data", nodeGraphHandler)
	http.HandleFunc("/api/health", healthHandler)

	// 서버 시작
	log.Println("Starting server on :5000...")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}
