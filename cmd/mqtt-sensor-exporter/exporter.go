package main

import (
	"crypto/tls"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
	"sync/atomic"
	"net/http"
	"flag"
)

var (
	brokerAddress = flag.String("broker-address", "127.0.0.1:1883", "Mqtt broker address")
	moistureSensor = prometheus.NewDesc(
		prometheus.BuildFQName("mqtt", "plant", "moisture_sensor_value"),
		"Moisture Sensor value",
		[]string{"plant"},
		nil,
	)
	tomatoMoisture int64
)

func onSensorValueReceived(_ mqtt.Client, message mqtt.Message) {
	x := string(message.Payload())
	v, err := strconv.ParseInt(x, 10, 32)
	if err != nil {
		panic(err)
	}
	atomic.SwapInt64(&tomatoMoisture, v)
}

type sensorExporter struct {
	client mqtt.Client
}

func (e *sensorExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- moistureSensor
}

func (e *sensorExporter) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(moistureSensor, prometheus.GaugeValue, float64(atomic.LoadInt64(&tomatoMoisture)), "tomato")
}

func main() {
	flag.Parse()
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + *brokerAddress).SetClientID("exporter")

	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)

	opts.OnConnect = func(c mqtt.Client) {
		if token := c.Subscribe("/tomato", 0, onSensorValueReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
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
