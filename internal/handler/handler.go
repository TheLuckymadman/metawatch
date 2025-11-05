package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/repository"
)

type (
	responseData struct {
		status int
		size int
	}
	ResponseWriterLogger struct {
		http.ResponseWriter
		responseData responseData
	}
)

func (r *ResponseWriterLogger) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size = size
	return size, err
}

func (r *ResponseWriterLogger) WriteHeader(statusCode int) {
	r.responseData.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func MiddlewareConveyor(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	for i := len(m)-1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func LoggerWrapper(sugar zap.SugaredLogger) func(h http.HandlerFunc) http.HandlerFunc {
	f := func(h http.HandlerFunc) http.HandlerFunc {	
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			rwl := ResponseWriterLogger{
				w, responseData{},
			}


			h.ServeHTTP(&rwl, r)

			duration := time.Since(start)
			
			sugar.Infoln(
				"uri", uri,
				"method", method,
				"duration", duration,
				"r_status", rwl.responseData.status,
				"r_size", rwl.responseData.size,
			)

		}) 
	}

	return f
}

func MetricReceiverHandler(s repository.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
		// if r.Header.Get("Content-Type") != "text/plain" {
		// 	http.Error(w, "Unsupported content type", http.StatusMethodNotAllowed)
		// 	w.WriteHeader(http.StatusMethodNotAllowed)
		// 	return
		// }
		
		metricType := chi.URLParam(r, "type")
		restPath := chi.URLParam(r, "*")
		reqPath := regexp.MustCompile(`^(.+)/(.+)$`)
		matches := reqPath.FindStringSubmatch(restPath)
		if len(matches) < 3 {
			http.Error(w, "invalid path format . Use: /update/metrictype/metricname/values\n", http.StatusNotFound)
			return
		}
		metricName := matches[1]
		metricValue := matches[2]
		agentIP := strings.Split(r.RemoteAddr, ":")[0]

		switch metricType {
		case model.Counter:
			metricType = model.Counter
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "invalid metric values\n", http.StatusBadRequest)
				return
			}
			err = s.SetMetric(agentIP, metricType, metricName, 0, value)
			if err != nil {
				errStr := fmt.Sprintf("Error while metric adding to db:\n%v", err)
				http.Error(w, errStr, http.StatusBadRequest)
				return 
			}

		case model.Gauge:
			metricType = model.Gauge
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "invalid metric values\n", http.StatusBadRequest)
				return 
			}
			err = s.SetMetric(agentIP, metricType, metricName, value, 0)
			if err != nil {
				errStr := fmt.Sprintf("Error while metric adding to db:\n%v", err)
				http.Error(w, errStr, http.StatusBadRequest)
				return
			}

		default:
			http.Error(w, "invalid metric type\n", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var reply = struct {
			Status string `json:"status"`
		}{Status: "ok"}
		body, _ := json.Marshal(reply)
		w.Write(body)
	})
}

func MetricGetterHandler(s repository.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		if r.Method != http.MethodGet {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
		reqPath := regexp.MustCompile(`^/value/(\w+)/([\w\-.]+)$`)
		matches := reqPath.FindStringSubmatch(r.URL.Path)
		if len(matches) != 3 {
			http.Error(w, "invalid path format. Use: /value/metrictype/metricname/values\n", http.StatusNotFound)
			return
		}
		regMetricType := matches[1]
		regMetricName := matches[2]
		agentIP := strings.Split(r.RemoteAddr, ":")[0]
		var body string

		switch regMetricType {
		case model.Counter:
			{
				_, delta, err := s.GetMetric(agentIP, regMetricType, regMetricName)
				if err != nil {
					errStr := fmt.Sprintf("Error while getting metric:\n%v", err)
					http.Error(w, errStr, http.StatusNotFound)
					return
				}
				body = fmt.Sprintf("%d", delta)
			}
		case model.Gauge:
			{
				value, _, err := s.GetMetric(agentIP, regMetricType, regMetricName)
				if err != nil {
					errStr := fmt.Sprintf("Error while getting metric:\n%v", err)
					http.Error(w, errStr, http.StatusNotFound)
					return
				}
				body = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", value), "0"), ".")

			}
		default:
			http.Error(w, "invalid metric type\n", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, body); err != nil {
			log.Printf("failed to write response body: %v", err)
		}
	})
}

func MetricsGetterHandler(s repository.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		if r.Method != http.MethodGet {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
		memStorage := s.(*repository.MemStorage)
		memStorage.RLock()
		copyMemStorage := s.(*repository.MemStorage).Metrics
		memStorage.RUnlock()

		sortedMetrics := make([]string, 0, len(copyMemStorage))
		for k := range copyMemStorage {
			sortedMetrics = append(sortedMetrics, k)
		}
		sort.Strings(sortedMetrics)

		body := `<!DOCTYPE html>
		<html lang="en">
		<head>
		<meta charset="utf-8">
		<title>Metrics</title>
		<style>
		body { font-family: sans-serif; margin: 20px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
		th { background: #f0f0f0; }
		</style>
		</head>
		<body>
		<h1>All Metrics</h1>
		<table>
		<tr>
			<th>Agent ID</th>
			<th>Name</th>
			<th>Type</th>
			<th>Value</th>
		</tr>
		`
		for _, v := range sortedMetrics {

			body += "<tr>"
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(v))
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(copyMemStorage[v].ID))
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(copyMemStorage[v].MType))
			switch copyMemStorage[v].MType {
			case model.Counter:
				if copyMemStorage[v].Delta != nil {
					body += fmt.Sprintf("<td>%d</td>", *copyMemStorage[v].Delta)
				} else {
					body += "<td>-</td>"
				}
			case model.Gauge:
				if copyMemStorage[v].Value != nil {
					body += fmt.Sprintf("<td>%f</td>", *copyMemStorage[v].Value)
				} else {
					body += "<td>-</td>"
				}
			default:
				body += "<td>?</td>"
			}
			body += "</tr>\n"
		}
		body += `</table>
		</body>
		</html>`

		w.Header().Set("Content-Type", "text/html")
		 w.WriteHeader(http.StatusOK)
		io.WriteString(w, body)
	})
}
