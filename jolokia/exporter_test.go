package jolokia

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"path"

	"bytes"

	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

func authTestHandler(w http.ResponseWriter, r *http.Request) {
	if u, p, ok := r.BasicAuth(); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized")
		return
	} else if u != "admin" || p != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Wrong credentials")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	http.ServeFile(w, r, path.Join("fixtures", "response.json"))
}

func checkRequestBody(t *testing.T, handlerFunc http.HandlerFunc) http.HandlerFunc {
	expectedBody, err := ioutil.ReadFile(path.Join("fixtures", "request.json"))
	if err != nil {
		t.Fatalf("error reading request.json: %v", err)
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("error reading request body: %s", err)
		}

		if bytes.Compare(expectedBody, b) != 0 {
			t.Fatalf("Requested body does not match. Expected to get %s, but got %s", expectedBody, b)
		}

		handlerFunc(rw, r)
	}
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	http.ServeFile(w, r, path.Join("fixtures", "response.json"))
}

func getPromResponse(t *testing.T) string {
	file, err := ioutil.ReadFile(filepath.Join("fixtures", "metrics.txt"))

	if err != nil {
		t.Fatalf("Unexpected exception reading file: %v", err)
	}

	return strings.TrimSpace(string(file))
}

func TestExporter_Describe(t *testing.T) {
	exp, err := NewExporter(log.Base(), expectedConfig, Namespace, false, "http://test/test", "", "")
	if err != nil {
		t.Fatal(err)
	}

	c := make(chan *prometheus.Desc, 1024)
	exp.Describe(c)

	if len(c) != 2 {
		t.Fatalf("Expected channel to have 2 objects, got %d", len(c))
	}

	up := <-c
	if up.String() != "Desc{fqName: \"jolokia_up\", help: \"Could jolokia endpoint be reached\", constLabels: {}, variableLabels: []}" {
		t.Errorf("Unexpected up metric description: %s", up.String())
	}

	duration := <-c
	if duration.String() != "Desc{fqName: \"jolokia_response_duration\", help: \"How long the jolokia endpoint took to deliver the metrics\", constLabels: {}, variableLabels: []}" {
		t.Errorf("Unexpected duration metric description: %s", duration.String())
	}
}

func TestExporter_Collect_NoAuth(t *testing.T) {
	srv := httptest.NewServer(checkRequestBody(t, http.HandlerFunc(testHandler)))

	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")
	exp, err := NewExporter(logger, expectedConfig, Namespace, false, srv.URL, "", "")
	if err != nil {
		t.Fatal(err)
	}

	c := make(chan prometheus.Metric, 1024)
	exp.Collect(c)

	bufStr := buf.String()
	if len(bufStr) != 0 {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}

	if len(c) != 17 {
		t.Fatalf("Expected channel to have 17 objects, got %d", len(c))
	}
}

func TestExporter_Collect_WithAuth(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")

	srv := httptest.NewServer(checkRequestBody(t, http.HandlerFunc(authTestHandler)))
	exp, err := NewExporter(logger, expectedConfig, Namespace, false, srv.URL, "admin", "secret")
	if err != nil {
		t.Fatal(err)
	}

	c := make(chan prometheus.Metric, 1024)
	exp.Collect(c)

	bufStr := buf.String()
	if len(bufStr) != 0 {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}

	if len(c) != 17 {
		t.Fatalf("Expected channel to have 17 objects, got %d", len(c))
	}
}

func TestExporter_Collect_WithAuthButNoneGiven(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")

	srv := httptest.NewServer(checkRequestBody(t, http.HandlerFunc(authTestHandler)))
	exp, err := NewExporter(logger, expectedConfig, Namespace, false, srv.URL, "", "")
	if err != nil {
		t.Fatal(err)
	}

	c := make(chan prometheus.Metric, 1024)
	exp.Collect(c)

	bufStr := buf.String()
	if ! strings.Contains(bufStr, "Error scraping jolokia endpoint: there was an error, response code is 401, expected 200") {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}
}

func TestExporter_Collect_WithAuthButWrongGiven(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")

	srv := httptest.NewServer(checkRequestBody(t, http.HandlerFunc(authTestHandler)))
	exp, err := NewExporter(logger, expectedConfig, Namespace, false, srv.URL, "", "")
	if err != nil {
		t.Fatal(err)
	}

	c := make(chan prometheus.Metric, 1024)
	exp.Collect(c)

	bufStr := buf.String()
	if ! strings.Contains(bufStr, "Error scraping jolokia endpoint: there was an error, response code is 401, expected 200") {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}
}

func TestExporter_Collect_WithPrometheus(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := log.NewLogger(buf)
	logger.SetLevel("warn")

	fixtureSrv := httptest.NewServer(http.HandlerFunc(authTestHandler))
	exp, err := NewExporter(logger, expectedConfig, Namespace, false, fixtureSrv.URL, "admin", "secret")
	if err != nil {
		t.Fatal(err)
	}

	prometheus.MustRegister(exp)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rw := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rw, req)

	if rw.Code != 200 {
		t.Errorf("expected status code to be %d, got %d", 200, rw.Code)
	}

	if rw.Body == nil {
		t.Fatal("Response does not have a body")
	}

	bufStr := buf.String()
	if len(bufStr) != 0 {
		t.Fatalf("unexpect collect output: %v", bufStr)
	}

	resBody := rw.Body.String()
	expectedBody := getPromResponse(t)

	if !strings.Contains(resBody, expectedBody) {
		t.Errorf("expected body to contain metrics %s, but doesn't: %s", expectedBody, resBody)
	}
}
