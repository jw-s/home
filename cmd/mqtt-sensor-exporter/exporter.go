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
	brokerAddress = flag.String("broker-address", "127.0.0.1:1883", "Mqtt broker address")

	moistureSensor = prometheus.NewDesc(
		prometheus.BuildFQName("mqtt", "plant", "moisture_sensor_value"),
		"Moisture Sensor value",
		[]string{"plant"},
		nil,
	)

	topics map[string]*int64

	plants = StringSlice("plants", []string{}, "List of plants to monitor, each plant is the topic name")
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

func init() {
	flag.Parse()
	topics = make(map[string]*int64)
	for _, plant := range *plants {
		topics["/"+plant] = new(int64)
	}
}

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + *brokerAddress).SetClientID("exporter")

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

type stringSliceValue struct {
	v *[]string
}

func (s *stringSliceValue) String() string {
	if s == nil || s.v == nil {
		return ""
	}

	return strings.Join(*s.v, ",")
}
func (s *stringSliceValue) Set(value string) error {
	split := strings.Split(value, ",")
	if len(split) == 1 && split[0] == value {
		*s.v = append(*s.v, value)
		return nil
	}
	for _, v := range split {
		*s.v = append(*s.v, v)
	}
	return nil
}

func newStringSlice(p *[]string, value []string) *stringSliceValue {
	v := new(stringSliceValue)
	v.v = p
	*v.v = value
	return v
}

func StringSlice(name string, value []string, usage string) *[]string {
	var p []string
	flag.Var(newStringSlice(&p, value), name, usage)
	return &p
}
