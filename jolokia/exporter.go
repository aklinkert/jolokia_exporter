package jolokia

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"regexp"

	"crypto/tls"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"encoding/json"
	"bytes"
	"io/ioutil"
)

var (
	keyRegExp = regexp.MustCompile("[^a-zA-Z0-9:_]")
)

// Exporter exports jolokia metrics for prometheus.
type Exporter struct {
	logger            log.Logger
	namespace         string
	URI               string
	mutex             sync.Mutex
	basicAuthUser     string
	basicAuthPassword string

	client   *http.Client
	up       *prometheus.Desc
	duration *prometheus.Desc

	config        *Config
	requestBody   []byte
	metricMapping map[string]string
}

// NewExporter returns an initialized Exporter.
func NewExporter(logger log.Logger, config *Config, namespace string, insecure bool, uri, basicAuthUser, basicAuthPassword string) (*Exporter, error) {
	exporter := &Exporter{
		config:            config,
		logger:            logger,
		URI:               uri,
		namespace:         namespace,
		basicAuthUser:     basicAuthUser,
		basicAuthPassword: basicAuthPassword,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could jolokia endpoint be reached",
			nil,
			nil),
		duration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "response_duration"),
			"How long the jolokia endpoint took to deliver the metrics",
			nil,
			nil),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
			},
		},
		metricMapping: make(map[string]string, 0),
	}

	if err := exporter.prepare(config); err != nil {
		return nil, err
	}

	return exporter, nil
}

// Describe describes all the metrics ever exported by the jolokia endpoint exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.duration
}

// Collect fetches the stats from configured location and delivers them
// as Prometheus metrics.
// It implements prometheus.Collector.
func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	req, err := http.NewRequest(http.MethodPost, e.URI, bytes.NewReader(e.requestBody))
	if err != nil {
		return err
	}

	req.SetBasicAuth(e.basicAuthUser, e.basicAuthPassword)
	startTime := time.Now()

	resp, err := e.client.Do(req)
	ch <- prometheus.MustNewConstMetric(e.duration, prometheus.GaugeValue, time.Since(startTime).Seconds())

	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return fmt.Errorf("error scraping jolokia endpoint: %v", err)
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	defer resp.Body.Close()
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("there was an error, response code is %d, expected 200", resp.StatusCode)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error unmarshalling json data: %v", err)
	}

	e.logger.Debugf("Result has %d rows", len(response))

	for _, metric := range response {
		target, ok := e.metricMapping[metric.Request.String()]
		if !ok {
			log.Errorf("Unable to find mapping for key %s", metric.Request.String())
		}

		values, err := getValues(target, metric.Value)
		if err != nil {
			log.Warnf("Failed to handle value %s for metric %s as understandable value: %v", metric.Value, metric.Request.String(), err)
		}

		for key, value := range values {
			e.logger.Debugf("Adding key %s with value %v", key, value)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(e.namespace, "", key),
					key,
					nil,
					nil),
				prometheus.UntypedValue,
				value)
		}

	}

	return nil
}

// converts any given key string to a prometheus acceptable key string
func keyToSnake(key string) string {
	return keyRegExp.ReplaceAllString(key, "_")
}

// Collects metrics, implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		e.logger.Errorf("Error scraping jolokia endpoint: %s", err)
	}
	return
}

func (e *Exporter) prepare(config *Config) (error) {
	req := Request{}

	for _, m := range config.Metrics {
		reqMetric := RequestMetric{
			Type:      requestTypeRead,
			Mbean:     m.Source.Mbean,
			Attribute: m.Source.Attribute,
			Path:      m.Source.Path,
		}

		e.metricMapping[reqMetric.String()] = m.Target
		req = append(req, reqMetric)
	}

	var err error
	e.requestBody, err = json.Marshal(req)
	return err
}
