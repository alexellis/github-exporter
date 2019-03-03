package main

import (
	"net/http"

	conf "github.com/alexellis/github-exporter/config"
	"github.com/alexellis/github-exporter/exporter"
	"github.com/fatih/structs"
	"github.com/infinityworks/go-common/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	log            *logrus.Logger
	applicationCfg conf.Config
	mets           map[string]*prometheus.Desc
)

func init() {
	applicationCfg = conf.Init()
	mets = exporter.AddMetrics()
	log = logger.Start(&applicationCfg)
}

func main() {

	log.WithFields(structs.Map(applicationCfg)).Info("Starting Exporter")

	exporter := exporter.Exporter{
		APIMetrics: mets,
		Config:     applicationCfg,
	}

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(&exporter)

	// Setup HTTP handler
	http.Handle(applicationCfg.MetricsPath(), prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		                <head><title>Github Exporter</title></head>
		                <body>
		                   <h1>GitHub Prometheus Metrics Exporter</h1>
						   <p>For more information, visit <a href=https://github.com/alexellis/github-exporter>GitHub</a></p>
		                   <p><a href='` + applicationCfg.MetricsPath() + `'>Metrics</a></p>
		                   </body>
		                </html>
		              `))
	})
	log.Fatal(http.ListenAndServe(":"+applicationCfg.ListenPort(), nil))
}
