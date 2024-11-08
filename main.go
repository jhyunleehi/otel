package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"otel/trace"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

var (
	targetCommand *string
)

func init() {
	log.SetLevel(log.DebugLevel)
	//	log.SetFormatter(&log.JSONFormatter{})
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

func main() {
	targetCommand = flag.String("target", "fio", "trace target command")
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":2224",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// Register handlers.
	handleFunc("/metrics", TraceHandler)
	handleFunc("/rolldice/", rolldice)

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

func TraceHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("MetricsHandler")
	t, err := trace.NewTrace(targetCommand)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		fmt.Fprintf(w, "Status: 500 Internal Server Error")
	}
	err = t.CreateNodeGraphData()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		fmt.Fprintf(w, "Status: 500 Internal Server Error")
	}
	err = t.CreatePrometheusMetric()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		fmt.Fprintf(w, "Status: 500 Internal Server Error")
	}
	promhttp.Handler().ServeHTTP(w, r)
}
