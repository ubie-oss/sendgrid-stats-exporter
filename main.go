package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/ubie-oss/sendgrid-stats-exporter/sendgrid"
)

const (
	namespace    = "sendgrid"
	exporterName = "sendgrid-stats-exporter"
)

const (
	stopTimeoutSecond = 10
)

var (
	gitCommit     string
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address to listen on for web interface and telemetry.",
	).Default(":9154").Envar("LISTEN_ADDRESS").String()
	disableExporterMetrics = kingpin.Flag(
		"web.disable-exporter-metrics",
		"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
	).Envar("DISABLE_EXPORTER_METRICS").Bool()
	sendGridAPIKey = kingpin.Flag(
		"sendgrid.api-key",
		"[Required] Set SendGrid API key",
	).Required().Envar("SENDGRID_API_KEY").String()
	sendGridUserName = kingpin.Flag(
		"sendgrid.username",
		"[Optional] Set SendGrid username as a label for each metrics. This is for identifying multiple SendGrid users metrics.",
	).Default("").Envar("SENDGRID_USER_NAME").String()
	location = kingpin.Flag(
		"sendgrid.location",
		"[Optional] Set a zone name.(e.g. 'Asia/Tokyo') The default is UTC.",
	).Default("").Envar("SENDGRID_LOCATION").String()
	timeOffset = kingpin.Flag(
		"sendgrid.time-offset",
		"[Optional] Specify the offset in second from UTC as an integer.(e.g. '32400') This needs to be set along with location.",
	).Default("0").Envar("SENDGRID_TIME_OFFSET").Int()
	sendGridCategoriesStr = kingpin.Flag(
		"sendgrid.categories",
		"[Optional] Comma-separeted SendGrid categories. If specified, corresponding category stats will be collected.",
	).Envar("SENDGRID_CATEGORIES").String()
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Info())
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting", exporterName, "version", version.Info(), gitCommit)
	level.Info(logger).Log("Build context", version.BuildContext())

	categories := sendGridCategories()
	level.Info(logger).Log("Target SendGrid categories:", strings.Join(categories, ", "))

	level.Info(logger).Log("msg", "Listening on", *listenAddress)

	client := sendgrid.NewDefaultClient(*sendGridAPIKey, logger)
	collector := NewCollector(logger, client, categories)
	prometheus.MustRegister(collector)
	prometheus.Unregister(collectors.NewGoCollector())
	registry := prometheus.NewRegistry()

	if !*disableExporterMetrics {
		registry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	}

	registry.MustRegister(collector)

	sig := make(chan os.Signal, 1)
	signal.Notify(
		sig,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	defer signal.Stop(sig)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	})

	srv := &http.Server{
		Addr:    *listenAddress,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			level.Error(logger).Log("err", err)
		}
	}()

	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), stopTimeoutSecond*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		level.Error(logger).Log("err", err)
	}
}

func sendGridCategories() []string {
	sendGridCategories := make([]string, 0)

	for _, v := range strings.Split(*sendGridCategoriesStr, ",") {
		if len(v) > 0 {
			sendGridCategories = append(sendGridCategories, v)
		}
	}

	return sendGridCategories
}
