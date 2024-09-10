package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	eero "github.com/imdevinc/go-eero"
	_ "github.com/joho/godotenv/autoload"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	uploadedBytes   *prometheus.GaugeVec
	downloadedBytes *prometheus.GaugeVec
}

func registerMetrics(reg prometheus.Registerer) *metrics {
	labels := []string{"hostname", "mac", "display_name", "device_type", "device_id"}
	uploaded := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eero_uploaded_bytes",
		Help: "Number of bytes uploaded over last 30 minutes",
	}, labels)
	downloaded := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eero_downloaded_bytes",
		Help: "Number of bytes downloaded over last 30 minutes",
	}, labels)

	reg.MustRegister(uploaded)
	reg.MustRegister(downloaded)
	return &metrics{
		uploadedBytes:   uploaded,
		downloadedBytes: downloaded,
	}
}

func updateMetrics(logger *slog.Logger, metrics *metrics) {
	go func() {
		e := eero.NewEero()
		e.UserToken = os.Getenv("EERO_USERTOKEN")
		networkID := os.Getenv("EERO_NETWORK_ID")
		eeroTimezone := os.Getenv("EERO_TIMEZONE")
		for {
			data, err := e.GetDataBreakdown(networkID, time.Now().Add(-1*time.Hour), time.Now(), eeroTimezone)
			if err != nil {
				logger.Error("error getting data breakdown", "error", err)
				time.Sleep(1 * time.Minute)
				continue
			}
			for _, d := range data.Devices {
				urlParts := strings.Split(d.URL, "/")
				deviceID := urlParts[len(urlParts)-1]
				metrics.uploadedBytes.WithLabelValues(d.Hostname, d.MAC, d.DisplayName, d.Type, deviceID).Set(float64(d.Upload))
				metrics.downloadedBytes.WithLabelValues(d.Hostname, d.MAC, d.DisplayName, d.Type, deviceID).Set(float64(d.Download))
			}
			time.Sleep(1 * time.Minute)
		}
	}()
}

func startServer(logger *slog.Logger) {
	rawPort := os.Getenv("PORT")
	port, err := strconv.Atoi(rawPort)
	if err != nil {
		port = 2112
	}

	reg := prometheus.NewPedanticRegistry()
	metrics := registerMetrics(reg)
	updateMetrics(logger, metrics)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	logger.Info("server started", "port", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func login(logger *slog.Logger) {
	if len(os.Args) != 3 {
		logger.Error("invalid usage", "options", "login <email|phone>")
		os.Exit(1)
	}
	e := eero.NewEero()
	err := e.Login(os.Args[2])
	if err != nil {
		logger.Error("failed to login", "error", err)
		os.Exit(1)
	}
	logger.Info("logged in successfully, needs validation", "usertoken", e.UserToken)
}

func validate(logger *slog.Logger) {
	if len(os.Args) != 3 {
		logger.Error("invalid usage", "options", "valiate <code>")
		os.Exit(1)
	}
	code := os.Args[2]
	e := eero.NewEero()
	e.UserToken = os.Getenv("EERO_USERTOKEN")
	err := e.VerifyLogin(code)
	if err != nil {
		logger.Error("failed to validate", "error", err)
		os.Exit(1)
	}
	logger.Info("validated successfully")
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if len(os.Args) < 2 {
		logger.Error("invalid usage", "options", "serve, login, validate")
		os.Exit(1)
	}
	command := strings.ToLower(os.Args[1])
	switch command {
	case "serve":
		startServer(logger)
	case "login":
		login(logger)
	case "validate":
		validate(logger)
	default:
		logger.Error("invalid usage", "options", "serve, login, validate")
		os.Exit(1)
	}
}
