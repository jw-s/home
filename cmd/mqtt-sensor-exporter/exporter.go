package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	brokerAddress  = flag.String("broker-address", "127.0.0.1:1883", "Mqtt broker address")
	moistureSensor = prometheus.NewDesc(
		prometheus.BuildFQName("mqtt", "plant", "moisture_sensor_value"),
		"Moisture Sensor value",
		[]string{"plant"},
		nil,
	)
	topics = map[string]*int64{
		"/tomato":    new(int64),
		"/garlic":    new(int64),
		"/blueberry": new(int64),
		"/raspberry": new(int64),
	}
	tomatoMoisture, garlicMoisture, blueberryMoisture, raspberryMoisture int64
)

func parseAndSwap(p *int64, value []byte) {
	x := string(value)
	v, err := strconv.ParseInt(x, 10, 32)
	if err != nil {
		panic(err)
	}
	atomic.SwapInt64(p, v)
}

func onSensorValueReceived(topic string) func(_ mqtt.Client, message mqtt.Message) {
	sensor := topics[topic]
	return func(_ mqtt.Client, message mqtt.Message) {
		parseAndSwap(sensor, message.Payload())
	}
}

type sensorExporter struct {
	client mqtt.Client
}

func (e *sensorExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- moistureSensor
}

func (e *sensorExporter) Collect(ch chan<- prometheus.Metric) {
	for topic, topicP := range topics {
		plant := strings.ReplaceAll(topic, "/", "")
		ch <- prometheus.MustNewConstMetric(moistureSensor, prometheus.GaugeValue, float64(atomic.LoadInt64(topicP)), plant)
	}
}

func main() {
	flag.Parse()
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + *brokerAddress).SetClientID("exporter-1")

	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)

	opts.OnConnect = func(c mqtt.Client) {
		for topic, _ := range topics {
			if token := c.Subscribe(topic, 0, onSensorValueReceived(topic)); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}
		}
	}

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	exporter := &sensorExporter{
		client: c,
	}
	prometheus.MustRegister(exporter)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
