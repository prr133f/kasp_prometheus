package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prr133f/kasp_prometheus/internal/detect"
)

const (
	defaultPort = 8080
)

var (
	// Метрика: тип хоста
	hostEnvironmentType = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "host_environment_type",
			Help: "Type of host where the service is running (vm, container, or physical)",
		},
		[]string{"type"},
	)

	// Метрика: информация о хосте
	hostInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "host_info",
			Help: "Metadata about the host where service runs",
		},
		[]string{"hostname", "os", "arch"},
	)
)

func init() {
	prometheus.MustRegister(hostEnvironmentType, hostInfo)
}

func main() {
	hostType := detect.DetectVirtualization()
	log.Printf("Detected host type: %s", hostType)

	hostEnvironmentType.WithLabelValues(string(hostType)).Set(1)

	metadata := detect.GetHostMetadata()
	hostInfo.WithLabelValues(
		metadata["hostname"],
		metadata["os"],
		metadata["arch"],
	).Set(1)

	port := os.Getenv("METRICS_PORT")
	if port == "" {
		port = fmt.Sprintf("%d", defaultPort)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	http.Handle("/", promhttp.Handler())

	// Запуск сервера
	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{Addr: addr}

	log.Printf("Starting Prometheus exporter on port %s", port)
	log.Printf("Metrics available at http://localhost:%s/", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}
