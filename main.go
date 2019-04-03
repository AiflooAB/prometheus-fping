package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	unreachable = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fping",
			Name:      "requests_noreply_count",
			Help:      "Number of requests missed",
		},
		[]string{"ip"},
	)
)

func init() {
	prometheus.MustRegister(responseTimes)
	prometheus.MustRegister(unreachable)
}

func main() {
	stopping := false

	network := getNetwork()

	fpingCmd := fping.NewFpingProcess(network)
	if err := fpingCmd.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Started scan of network %s\n", network)

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
		case response := <-fpingCmd.Unreachables:
			unreachable.WithLabelValues(response.IP.String()).Inc()
		}

	}
}

func getNetwork() string {
	network, exists := os.LookupEnv("NETWORK")
	if exists {
		return network
	}
	network, err := getFirstEthernetNetwork()
	if err != nil {
		log.Fatal(err)
	}

	return network
}

func getFirstEthernetNetwork() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagPointToPoint != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			ip := net.ParseIP(strings.Split(addr.String(), "/")[0])
			ipv4 := ip.To4()
			if ipv4 == nil {
				continue
			}
			return addr.String(), nil
		}
	}
	return "", nil
}
