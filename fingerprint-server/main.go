package main

import (
	"github.com/oschwald/maxminddb-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"proxy/collector"
)

var requestCounter prometheus.Counter

func main() {
	viper.SetDefault("geoCityLiteDBPath", "./GeoLite2-City.mmdb")
	viper.BindEnv("geoCityLiteDBPath", "GEOCITYLITE_DB_PATH")

	viper.SetDefault("geoASNLiteDBPath", "./GeoLite2-ASN.mmdb")
	viper.BindEnv("geoASNLiteDBPath", "GEOASNLITE_DB_PATH")

	viper.SetDefault("fingerprintServerPort", ":8088")
	viper.BindEnv("fingerprintServerPort", "FINGERPRINT_SERVER_PORT")

	viper.SetDefault("enablePrometheusMetric", true)
	viper.SetDefault("metricPort", ":8091")

	viper.BindEnv("enablePrometheusMetric", "ENABLE_PROMETHEUS_METRIC")
	viper.BindEnv("metricPort", "METRIC_PORT")

	if viper.GetBool("enablePrometheusMetric") {
		requestCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "scrapoxy",
			Subsystem: "fingerprint_server",
			Name:      "requests_count",
		})
		extraCollector := NewCollector("scrapoxy", "fingerprint_server")

		c := collector.Collector{
			Namespace: "scrapoxy",
			Subsystem: "fingerprint_server",
			EnableCPU: true,
			EnableMem: true,
		}
		prometheus.MustRegister(collector.NewPrometheusMetrics(c, extraCollector))
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Starting metric server on %s\n", viper.GetString("metricPort"))
		go http.ListenAndServe(viper.GetString("metricPort"), nil)
	}

	cityDB, err := maxminddb.Open(viper.GetString("geoCityLiteDBPath"))
	if err != nil {
		log.Fatal(err)
	}
	defer cityDB.Close()

	asnDB, err := maxminddb.Open(viper.GetString("geoASNLiteDBPath"))
	if err != nil {
		log.Fatal(err)
	}
	defer asnDB.Close()

	handler := NewHandler(cityDB, asnDB)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Handle404Request)
	mux.HandleFunc("/api/json", handler.HandleJsonAPIRequest)

	// Start the server and log any errors
	log.Printf("Starting fingerprint-server server on %s\n", viper.GetString("fingerprintServerPort"))
	err = http.ListenAndServe(viper.GetString("fingerprintServerPort"), mux)
	if err != nil {
		log.Fatal("Error starting fingerprint-server server: ", err)
	}
}
