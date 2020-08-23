package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gokitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	gokitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/jnikolaeva/eshop-common/httpkit"
	postgresadapter "github.com/jnikolaeva/eshop-common/postgres"

	"github.com/jnikolaeva/catalogservice/internal/catalog/application"
	httptransport "github.com/jnikolaeva/catalogservice/internal/catalog/infrastructure/http"

	"github.com/jnikolaeva/catalogservice/internal/probes"

	"github.com/jnikolaeva/catalogservice/internal/catalog/infrastructure/postgres"
)

const (
	appName     = "catalogservice"
	defaultPort = "8080"
)

func main() {
	serverAddr := ":" + envString("APP_PORT", defaultPort)

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "@timestamp",
			logrus.FieldKeyMsg:  "message",
		},
	})

	errorLogger := gokitlog.NewJSONLogger(gokitlog.NewSyncWriter(os.Stderr))
	errorLogger = level.NewFilter(errorLogger, level.AllowDebug())
	errorLogger = gokitlog.With(errorLogger,
		"appName", appName,
		"@timestamp", gokitlog.DefaultTimestampUTC,
	)

	connConfig, err := postgresadapter.ParseEnvConfig(appName)
	if err != nil {
		logger.Fatal(err.Error())
	}
	connectionPool, err := postgresadapter.NewConnectionPool(connConfig)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer connectionPool.Close()

	repository := postgres.New(connectionPool)
	service := application.NewService(repository)
	endpoints := httptransport.MakeEndpoints(service)

	metrics := httpkit.NewMetricsHolder(gokitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Namespace: "catalog",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, []string{"method", "endpoint", "status_code"}),
		gokitprometheus.NewHistogramFrom(prometheus.HistogramOpts{
			Namespace: "catalog",
			Name:      "request_latency_seconds",
			Help:      "Total duration of request in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "endpoint"}))

	mux := http.NewServeMux()

	mux.Handle("/api/v1/", httptransport.MakeHandler("/api/v1/catalog", endpoints, errorLogger, metrics))
	mux.Handle("/ready", probes.MakeReadyHandler())
	mux.Handle("/live", probes.MakeLiveHandler())
	mux.Handle("/metrics", promhttp.Handler())

	srv := startServer(serverAddr, mux, logger)

	waitForShutdown(srv)
	logger.Info("shutting down")
}

func startServer(serverAddr string, handler http.Handler, logger *logrus.Logger) *http.Server {
	srv := &http.Server{Addr: serverAddr, Handler: handler}

	go func() {
		logger.WithFields(logrus.Fields{"url": serverAddr}).Info("starting the server")
		logger.Fatal(srv.ListenAndServe())
	}()

	return srv
}

func waitForShutdown(srv *http.Server) {
	killSignalChan := make(chan os.Signal, 1)
	signal.Notify(killSignalChan, os.Kill, os.Interrupt, syscall.SIGTERM)

	<-killSignalChan
	_ = srv.Shutdown(context.Background())
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
