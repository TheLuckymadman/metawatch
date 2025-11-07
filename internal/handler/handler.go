package handler

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/service"
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
	ResponseWriterCompressor struct {
		http.ResponseWriter
		gzip *gzip.Writer
		needCompress bool
	}
)

func (r *ResponseWriterCompressor) Write(b []byte) (int, error) {
	if r.needCompress {
		return r.gzip.Write(b)	
	}
	return r.ResponseWriter.Write(b)
}

func (r *ResponseWriterCompressor) WriteHeader(statusCode int) {
	if strings.Contains(r.ResponseWriter.Header().Get("Content-Type"), "text/html") || strings.Contains(r.Header().Get("Content-Type"), "application/json") {
		r.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		r.needCompress = true
	}
	
	r.ResponseWriter.WriteHeader(statusCode)
}


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

func CompressWrapper(h http.HandlerFunc) http.HandlerFunc {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			log.Println("content decompressing is starting")
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create gzip reader: %v", err), http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			r.Body = io.NopCloser(gz)
			r.Header.Del("Content-Encoding")
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			log.Printf("compressing is requested")
			
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create gzip writer: %v", err), http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			rwc := ResponseWriterCompressor{w, gz, false}
			h(&rwc, r)	
			return 
		}
		h(w, r)
	})
	return f
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

func MetricSetterHandler(s service.Storage) http.HandlerFunc {
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

		err := service.AddMetric(metricName, metricValue, metricType, agentIP, s)
		if err != nil {
			http.Error(w, "invalid metric values\n", http.StatusBadRequest)
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

func JSONSetterHandler(s service.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported content type", http.StatusMethodNotAllowed)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}

		var metric model.Metrics
		jsonDecoder := json.NewDecoder(r.Body)
		if err := jsonDecoder.Decode(&metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		agentIP := strings.Split(r.RemoteAddr, ":")[0]
		err := service.AddObjMetric(metric, agentIP, s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return 
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var reply = struct {
			Status string `json:"status"`
		}{Status: "ok"}
		body, _ := json.Marshal(reply)
		w.Write(body)
	}
}

func MetricGetterHandler(s service.Storage) http.HandlerFunc {
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
		metricType := matches[1]
		metricName := matches[2]
		agentIP := strings.Split(r.RemoteAddr, ":")[0]
		
		body, err := service.GetMetric(metricName, metricType, agentIP, s)
		if err != nil {
			errStr := fmt.Sprintf("There was a error while getting the metric:\n%v", err)
			http.Error(w, errStr, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, body); err != nil {
			log.Printf("failed to write response body: %v", err)
		}
	})
}

func JSONGetterHandler(s service.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported content type", http.StatusMethodNotAllowed)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}

		agentIP := strings.Split(r.RemoteAddr, ":")[0]
		var metricReq model.Metrics
		jsonDecoder := json.NewDecoder(r.Body)
		if err := jsonDecoder.Decode(&metricReq); err != nil {
			log.Printf("Decoding json request failed with %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metricResp, err := service.GetObjMetric(metricReq, agentIP, s)
		if err != nil {
			log.Printf("Getting object metrics failed with %v", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonEncoder := json.NewEncoder(w)
		if err := jsonEncoder.Encode(&metricResp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func MetricsListHandler(s service.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		if r.Method != http.MethodGet {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
		
		memStorage, metricIdx, err := service.ListMetric(s)
		if err != nil {
			errStr := fmt.Sprintf("There was a error while listing metrics:\n%v", err)
			http.Error(w, errStr, http.StatusInternalServerError)
		}

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
		for _, v := range metricIdx {

			body += "<tr>"
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(v))
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(memStorage[v].ID))
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(memStorage[v].MType))
			switch memStorage[v].MType {
			case model.Counter:
				if memStorage[v].Delta != nil {
					body += fmt.Sprintf("<td>%d</td>", *memStorage[v].Delta)
				} else {
					body += "<td>-</td>"
				}
			case model.Gauge:
				if memStorage[v].Value != nil {
					body += fmt.Sprintf("<td>%f</td>", *memStorage[v].Value)
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
