package exporter

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/common/version"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	namespace = "ttn"
)

type ttnCollector struct {
	up            *prometheus.Desc
	uplinkCount   *prometheus.Desc
	downlinkCount *prometheus.Desc
	metricsAckr   *prometheus.Desc
	metricsLpps   *prometheus.Desc
	metricsRxfw   *prometheus.Desc
	metricsRxin   *prometheus.Desc
	metricsRxok   *prometheus.Desc
	metricsTxin   *prometheus.Desc
	metricsTxok   *prometheus.Desc

	apiToken string
	gateway  string
}

func newTtnCollector(apiToken string, gateway string) *ttnCollector {
	return &ttnCollector{
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Was the last API query successful.",
			nil, nil,
		),
		uplinkCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uplink_count"),
			"How many uplinks have been done.",
			nil, nil,
		),
		downlinkCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "downlink_count"),
			"How many downlinks have been done.",
			nil, nil,
		),
		metricsAckr: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "ackr"),
			"How many ackr have been done.",
			nil, nil,
		),
		metricsLpps: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "llps"),
			"How many llps have been done.",
			nil, nil,
		),
		metricsRxfw: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "rxfw"),
			"How many rxfw have been done.",
			nil, nil,
		),
		metricsRxin: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "rxin"),
			"How many rxin have been done.",
			nil, nil,
		),
		metricsRxok: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "rxok"),
			"How many rxok have been done.",
			nil, nil,
		),
		metricsTxin: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "txin"),
			"How many txin have been done.",
			nil, nil,
		),
		metricsTxok: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "txok"),
			"How many txok have been done.",
			nil, nil,
		),
		apiToken: apiToken,
		gateway:  gateway,
	}
}

func (collector *ttnCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.up
	ch <- collector.uplinkCount
	ch <- collector.downlinkCount
}

func (collector *ttnCollector) Collect(ch chan<- prometheus.Metric) {
	url := fmt.Sprintf("https://eu1.cloud.thethings.network/api/v3/gs/gateways/%s/connection/stats",
		collector.gateway)
	var bearer = "Bearer " + collector.apiToken

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}
	var gatewayStats GatewayConnectionStats
	json.Unmarshal(body, &gatewayStats)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			collector.up, prometheus.GaugeValue, 0,
		)
		log.Println(err)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		collector.up, prometheus.GaugeValue, 1,
	)
	uplinkCount, _ := strconv.ParseFloat(gatewayStats.UplinkCount, 8)
	ch <- prometheus.MustNewConstMetric(
		collector.uplinkCount,
		prometheus.GaugeValue,
		uplinkCount,
	)
	downlinkCount, _ := strconv.ParseFloat(gatewayStats.DownlinkCount, 8)
	ch <- prometheus.MustNewConstMetric(
		collector.downlinkCount,
		prometheus.GaugeValue,
		downlinkCount,
	)
	ch <- prometheus.MustNewConstMetric(
		collector.metricsAckr,
		prometheus.GaugeValue,
		float64(gatewayStats.LastStatus.Metrics.Ackr),
	)
	ch <- prometheus.MustNewConstMetric(
		collector.metricsLpps,
		prometheus.GaugeValue,
		float64(gatewayStats.LastStatus.Metrics.Lpps),
	)
	ch <- prometheus.MustNewConstMetric(
		collector.metricsRxfw,
		prometheus.GaugeValue,
		float64(gatewayStats.LastStatus.Metrics.Rxfw),
	)
	ch <- prometheus.MustNewConstMetric(
		collector.metricsRxin,
		prometheus.GaugeValue,
		float64(gatewayStats.LastStatus.Metrics.Rxin),
	)
	ch <- prometheus.MustNewConstMetric(
		collector.metricsRxok,
		prometheus.GaugeValue,
		float64(gatewayStats.LastStatus.Metrics.Rxok),
	)
	ch <- prometheus.MustNewConstMetric(
		collector.metricsTxin,
		prometheus.GaugeValue,
		float64(gatewayStats.LastStatus.Metrics.Txin),
	)
	ch <- prometheus.MustNewConstMetric(
		collector.metricsTxok,
		prometheus.GaugeValue,
		float64(gatewayStats.LastStatus.Metrics.Txok),
	)
}

// Register registers the metrics
func Register(apiToken string, gateway string) {
	collector := newTtnCollector(apiToken, gateway)
	prometheus.MustRegister(version.NewCollector("ttn_exporter"))
	prometheus.MustRegister(collector)
	prometheus.Unregister(collectors.NewGoCollector())
}

type GatewayConnectionStats struct {
	ConnectedAt          time.Time `json:"connected_at"`
	Protocol             string    `json:"protocol"`
	LastStatusReceivedAt time.Time `json:"last_status_received_at"`
	LastStatus           struct {
		Time     time.Time `json:"time"`
		BootTime time.Time `json:"boot_time"`
		Versions struct {
			Fpga               string `json:"fpga"`
			Hal                string `json:"hal"`
			TtnLwGatewayServer string `json:"ttn-lw-gateway-server"`
		} `json:"versions"`
		AntennaLocations []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Altitude  int     `json:"altitude"`
		} `json:"antenna_locations"`
		Ip      []string `json:"ip"`
		Metrics struct {
			Ackr int `json:"ackr"`
			Lpps int `json:"lpps"`
			Rxfw int `json:"rxfw"`
			Rxin int `json:"rxin"`
			Rxok int `json:"rxok"`
			Txin int `json:"txin"`
			Txok int `json:"txok"`
		} `json:"metrics"`
	} `json:"last_status"`
	LastUplinkReceivedAt   time.Time `json:"last_uplink_received_at"`
	UplinkCount            string    `json:"uplink_count"`
	LastDownlinkReceivedAt time.Time `json:"last_downlink_received_at"`
	DownlinkCount          string    `json:"downlink_count"`
	RoundTripTimes         struct {
		Min    string `json:"min"`
		Max    string `json:"max"`
		Median string `json:"median"`
		Count  int    `json:"count"`
	} `json:"round_trip_times"`
	SubBands []struct {
		MinFrequency             string  `json:"min_frequency"`
		MaxFrequency             string  `json:"max_frequency"`
		DownlinkUtilizationLimit float64 `json:"downlink_utilization_limit"`
	} `json:"sub_bands"`
}
