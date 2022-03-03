package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	var listenAddress string
	var metricsPath string
	var healthzPath string
	var queueitBaseURL string
	var queueitAPIKeyPath string
	var omitTestWaitingRooms bool
	var apiKey string

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	flag.StringVar(&listenAddress, "web.listen-address", ":8000", "Address on which to expose metrics and web interface.")
	flag.StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	flag.StringVar(&healthzPath, "web.healthcheck-path", "/healthz", "Path under which to run healthchecks")
	flag.StringVar(&queueitBaseURL, "config.queue-it-base-url", "", "Base URL to your Queue-it api")
	flag.StringVar(&queueitAPIKeyPath, "config.queue-it-api-key-path", "", "Absolute path to Queue-it API Key file")
	flag.BoolVar(&omitTestWaitingRooms, "config.omit-test-waiting-rooms", true, "Whether to filter out test waiting rooms metrics")
	flag.Parse()

	if queueitAPIKeyPath != "" {
		content, err := os.ReadFile(queueitAPIKeyPath)
		if err != nil {
			panic("cannot read file from config.queue-it-api-key-path: " + queueitAPIKeyPath)
		}

		apiKey = string(content)
	} else {
		apiKey = os.Getenv("QUEUE_IT_API_KEY")
	}

	// Can't do anything without an API key
	if apiKey == "" {
		panic("please provide a queue-it api key as the environment variable QUEUEIT_API_KEY or a mounted file with its path set to -config.queue-it-api-key-path")
	}

	c := newCollector(
		logger,
		newQueueitAPI(
			logger,
			queueitBaseURL,
			apiKey,
			omitTestWaitingRooms,
		),
	)

	// Register collector
	prometheus.MustRegister(c)

	// Add root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
	           <head><title>Kube Metrics Server</title></head>
	           <body>
	           <h1>Kube Metrics</h1>
			 <ul>
	           <li><a href='` + metricsPath + `'>metrics</a></li>
	           <li><a href='` + healthzPath + `'>healthz</a></li>
			 </ul>
	           </body>
	           </html>`))
	})

	// Add healthzPath
	http.HandleFunc(healthzPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	})

	// Handle metrics requests
	http.Handle(metricsPath, promhttp.Handler())

	// Listen
	logger.Info("queue-it exporter is listening", zap.String("address", listenAddress))
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
