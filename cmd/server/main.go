package main

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/TheLuckymadman/metawatch/internal/models"
	"github.com/TheLuckymadman/metawatch/internal/repository"
)

func metricReceiverHandler(s repository.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Unsupported content type", http.StatusMethodNotAllowed)
			return
		}
		reqPath := regexp.MustCompile(`/update/(\w+)/(\w+)`)
		matches := reqPath.FindStringSubmatch(r.URL.Path)
		if len(matches) != 3 {
			http.Error(w, "invalid metric name. Use: /update/metrictype/metricname/values\n", http.StatusNotFound)
			return
		}
		regMetricType := matches[1]
		regMetricName := matches[2]
		reqPath = regexp.MustCompile(`/update/\w+/\w+/(\d+)`)
		matches = reqPath.FindStringSubmatch(r.URL.Path)
		if len(matches) != 2 {
			http.Error(w, "invalid metric values\n", http.StatusBadRequest)
			return
		}
		reqValue := matches[1]
		var metricType string
		agentIP := strings.Split(r.RemoteAddr, ":")[0]

		switch regMetricType {
		case models.Counter:
			metricType = models.Counter
			value, err := strconv.ParseInt(reqValue, 10, 64)
			if err != nil {
				http.Error(w, "ivalid metric values\n", http.StatusBadRequest)
			}
			s.SetMetric(agentIP, metricType, regMetricName, 0, value)
		case models.Gauge:
			metricType = models.Gauge
			value, err := strconv.ParseFloat(reqValue, 64)
			if err != nil {
				http.Error(w, "ivalid metric values\n", http.StatusBadRequest)
			}
			s.SetMetric(agentIP, metricType, regMetricName, value, 0)
		default:
			http.Error(w, "ivalid metric type\n", http.StatusBadRequest)
		}

		// memStorage, _ := s.(*repository.MemStorage)
		// for k, v := range memStorage.Metrics {
		// 	fmt.Printf("key: %v\n", k)
		// 	if v.Delta != nil {
		// 		fmt.Printf("Delta: %d\n", *v.Delta)
		// 	}
		// 	if v.Value != nil {
		// 		fmt.Printf("Delta: %f\n", *v.Value)
		// 	}
		// }
	})
}

func run() error {
	s := repository.NewStorage()
	m := http.NewServeMux()
	m.HandleFunc("/update/", metricReceiverHandler(s))
	return http.ListenAndServe("localhost:8080", m)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
