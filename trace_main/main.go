package main

import (
	"fmt"
	"otel/model"
	"otel/trace"
	"path"
	"runtime"
	"strings"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
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
	model.SetCode()
}

// TestCalculatorTestSuite: 스위트를 실행하는 메인 테스트 함수
func main() {
	command := "fio"
	tr, err := trace.NewTrace(&command)
	if err != nil {
		log.Error(err)
		return
	}
	tr.CreateISCSIInfo()
	tr.CreateDeviceMap()
	tr.CreateFileSystemMap()
	tr.CreateNodeGraph()	
}

