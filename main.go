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

	"github.com/AiflooAB/prometheus-fping/pkg/fping"
)

var (
	responseTimes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fping",
			Name:      "response_duration_seconds",
			Help:      "A histogram of latencies for ICMP echo requests.",
			Buckets:   prometheus.ExponentialBuckets(0.00005, 2, 20),
		},
		[]string{"ip"},
	)
)

func init() {
	prometheus.MustRegister(responseTimes)
}

func main() {
	stopping := false

	fpingCmd := fping.NewFpingProcess()
	if err := fpingCmd.Start(); err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	go func() { log.Fatal(http.ListenAndServe("0.0.0.0:9299", nil)) }()

	for {
		select {
		case <-signals:
			if stopping {
				fmt.Println("Got multiple shutdown signals, forcing exit")
				os.Exit(5)
			}
			fmt.Println("Got shutdown signal, shutting down...")
			fpingCmd.Stop()
			stopping = true
			return
		case response := <-fpingCmd.Responses:
			responseTimes.WithLabelValues(response.IP.String()).Observe(response.Roundtrip.Seconds())
		}
	}
}
